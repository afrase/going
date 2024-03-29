package shell

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/session-manager-plugin/src/datachannel"
	"github.com/aws/session-manager-plugin/src/log"
	"github.com/aws/session-manager-plugin/src/sessionmanagerplugin/session"
	"github.com/manifoldco/promptui"

	// import for side effect of registering the shell session
	_ "github.com/aws/session-manager-plugin/src/sessionmanagerplugin/session/shellsession"
	"github.com/google/uuid"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/cobra"

	"going/internal"
	"going/internal/client"
	"going/internal/factory"
	"going/internal/utils"
)

type shellOptions struct {
	// The command flags
	ClusterInput   string
	ServiceInput   string
	ContainerInput string
	UseSSM         bool

	target client.Container
	client *client.AWSClient
}

var opts = &shellOptions{}

var containerPromptTemplate = &promptui.SelectTemplates{
	Label:    fmt.Sprintf("%s {{ . }}: ", promptui.IconInitial),
	Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
	Inactive: "  {{ .Name }}",
	Selected: fmt.Sprintf(`{{ "%s" | green }} {{ .Name | faint }}`, promptui.IconGood),
	Details: `{{ "Status:" | faint }} {{ .LastStatus }}
{{ "Health:" | faint }} {{ .Health }}`,
}

func NewCmdShell(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Open a shell to a container in ECS",
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.client = client.New(f.Context, f.Config())
		},
		Run: func(cmd *cobra.Command, args []string) {
			// Must be logged in
			err := internal.CheckSSOLogin(f)
			utils.CheckErr(err)

			if opts.ClusterInput == "" {
				opts.ClusterInput = promptForCluster(f)
			}

			if opts.ServiceInput == "" {
				opts.ServiceInput = promptForService(f)
			}

			taskARN := getTaskArn(f)
			opts.target = getContainerDetails(f, taskARN)

			fmt.Printf("cluster: \"%s\" service: \"%s\" container: \"%s\"\n",
				opts.target.ClusterName, opts.target.ServiceName, opts.target.Name)

			if !opts.target.ExecuteAgentRunning {
				fmt.Println("AWS is reporting the \"ExecuteCommandAgent\" is not running, connection will use SSM directly.")
				opts.UseSSM = true
			}

			yes := f.Prompt.YesNoPrompt("Connect to the above container")
			if !yes {
				// just returning seems to sometimes not restore the shells cursor, so exit.
				os.Exit(0)
			}

			if opts.UseSSM {
				getBasicShell(f)
			} else {
				getShellUsingECS(f)
			}
		},
	}

	cmd.Flags().StringVarP(&opts.ClusterInput, "cluster", "c", "", "The cluster name")
	cmd.Flags().StringVarP(&opts.ServiceInput, "service", "s", "", "The service name")
	cmd.Flags().StringVarP(&opts.ContainerInput, "container", "r", "", "The container name")
	cmd.Flags().BoolVar(&opts.UseSSM, "ssm", false, "Use SSM directly to get a shell")

	return cmd
}

func promptForCluster(f *factory.Factory) string {
	c, err := opts.client.ListClusters()
	utils.CheckErr(err)

	var clusters []string
	for _, cluster := range c {
		clusters = append(clusters, cluster.Name)
	}

	return f.Prompt.Select("Select a cluster", clusters)
}

func promptForService(f *factory.Factory) string {
	s, err := opts.client.ListServices(opts.ClusterInput)
	utils.CheckErr(err)

	var services []string
	for _, service := range s {
		services = append(services, service.Name)
	}

	return f.Prompt.Select("Select a service", services)
}

func getTaskArn(f *factory.Factory) string {
	t, err := opts.client.ListTasks(opts.ClusterInput, opts.ServiceInput)
	utils.CheckErr(err)

	switch len(t) {
	case 0:
		yes := f.Prompt.YesNoPrompt("No tasks running. Start one")
		if yes {
			err = opts.client.UpdateService(&ecs.UpdateServiceInput{
				Cluster:      aws.String(opts.ClusterInput),
				Service:      aws.String(opts.ServiceInput),
				DesiredCount: aws.Int32(1),
			})
			utils.CheckErr(err)
			fmt.Println("Set desired count of service to 1. Could take a few minutes to start.")
		}

		os.Exit(1)
		return "" // won't reach
	case 1:
		return t[0]
	default:
		return f.Prompt.Select("Multiple tasks running, please select one", t)
	}
}

func getContainerDetails(f *factory.Factory, taskARN string) client.Container {
	var details client.Container

	if opts.ContainerInput == "" {
		containers, err := opts.client.DescribeContainers(opts.ClusterInput, taskARN)
		utils.CheckErr(err)
		i := f.Prompt.CustomSelect("Select a container", containers, containerPromptTemplate, containerSearch(containers))
		details = containers[i]
	} else {
		c, err := opts.client.DescribeContainer(opts.ClusterInput, taskARN, opts.ContainerInput)
		utils.CheckErr(err)
		details = c
	}

	opts.ContainerInput = details.Name
	return details
}

func containerSearch(containers []client.Container) func(input string, index int) bool {
	return func(input string, index int) bool {
		item := containers[index]
		if fuzzy.MatchFold(input, item.Name) {
			return true
		}
		return false
	}
}

func getBasicShell(f *factory.Factory) {
	target, err := opts.target.SSMTarget()
	utils.CheckErr(err)

	ssmClient := ssm.NewFromConfig(f.Config())
	out, err := ssmClient.StartSession(f.Context, &ssm.StartSessionInput{Target: aws.String(target)})
	utils.CheckErr(err)

	ep, err := ssm.NewDefaultEndpointResolver().ResolveEndpoint(f.Config().Region, ssm.EndpointResolverOptions{})
	utils.CheckErr(err)

	ssmSession := session.Session{
		SessionId:   aws.ToString(out.SessionId),
		StreamUrl:   aws.ToString(out.StreamUrl),
		TokenValue:  aws.ToString(out.TokenValue),
		Endpoint:    ep.URL,
		ClientId:    uuid.NewString(),
		TargetId:    target,
		DataChannel: &datachannel.DataChannel{},
	}

	fmt.Println("Connecting with a basic `sh' shell. After connecting run `/bin/bash' to get a nicer shell.")
	fmt.Println("Don't forget you will have to call `exit' twice to end the connection if you change to bash.")

	utils.CheckErr(ssmSession.Execute(log.Logger(false, ssmSession.ClientId)))
}

func getShellUsingECS(f *factory.Factory) {
	target, err := opts.target.SSMTarget()
	utils.CheckErr(err)

	out, err := opts.client.ExecuteCommand(&ecs.ExecuteCommandInput{
		Cluster:     aws.String(opts.target.ClusterARN),
		Container:   aws.String(opts.target.Name),
		Task:        aws.String(opts.target.TaskARN),
		Command:     aws.String("/bin/bash"),
		Interactive: true,
	})
	utils.CheckErr(err)

	ep, err := ssm.NewDefaultEndpointResolver().ResolveEndpoint(f.Config().Region, ssm.EndpointResolverOptions{})
	utils.CheckErr(err)

	ssmSession := session.Session{
		SessionId:   aws.ToString(out.Session.SessionId),
		StreamUrl:   aws.ToString(out.Session.StreamUrl),
		TokenValue:  aws.ToString(out.Session.TokenValue),
		Endpoint:    ep.URL,
		ClientId:    uuid.NewString(),
		TargetId:    target,
		DataChannel: &datachannel.DataChannel{},
	}

	utils.CheckErr(ssmSession.Execute(log.Logger(false, ssmSession.ClientId)))
}
