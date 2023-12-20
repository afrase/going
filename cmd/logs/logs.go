package logs

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"going/internal"
	"going/internal/client"
	"going/internal/factory"
	"going/internal/utils"
)

type logOptions struct {
	ClusterInput   string
	ServiceInput   string
	ContainerInput string
	Minutes        int

	target client.Container
	client *client.AWSClient
}

type targetLogConfig struct {
	GroupName    string
	StreamPrefix string
}

var opts = &logOptions{}

var containerPromptTemplate = &promptui.SelectTemplates{
	Label:    fmt.Sprintf("%s {{ . }}: ", promptui.IconInitial),
	Active:   fmt.Sprintf("%s {{ .Name | underline }}", promptui.IconSelect),
	Inactive: "  {{ .Name }}",
	Selected: fmt.Sprintf(`{{ "%s" | green }} {{ .Name | faint }}`, promptui.IconGood),
	Details: `{{ "Status:" | faint }} {{ .LastStatus }}
{{ "Health:" | faint }} {{ .Health }}`,
}

func NewCmdLogs(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use: "logs",
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.client = client.New(f.Context, f.Config())
		},
		Run: func(cmd *cobra.Command, args []string) {
			err := internal.CheckSSOLogin(f)
			utils.CheckErr(err)

			if opts.ClusterInput == "" {
				opts.ClusterInput = promptForCluster(f)
			}

			if opts.ServiceInput == "" {
				opts.ServiceInput = promptForService(f)
			}

			taskARN := getTaskArn(f)
			logDetails, err := getLogGroup(f, taskARN)
			utils.CheckErr(err)

			startTime := time.Now().Add(-time.Duration(opts.Minutes) * time.Minute)
			fmt.Printf("Tailing logs for CloudWatch group \"%s\" with prefix \"%s\"\n\n",
				logDetails.GroupName, logDetails.StreamPrefix)

			err = opts.client.TailLogs(logDetails.GroupName, logDetails.StreamPrefix, startTime, func(e client.LogEvent) {
				fmt.Printf("%s [%s] %s\n", e.StreamName, e.Timestamp, e.Message)
			})
			utils.CheckErr(err)
		},
	}

	cmd.Flags().IntVarP(&opts.Minutes, "minutes", "t", 30, "Number of minutes back to filter logs")
	cmd.Flags().StringVarP(&opts.ClusterInput, "cluster", "c", "", "The cluster name")
	cmd.Flags().StringVarP(&opts.ServiceInput, "service", "s", "", "The service name")
	cmd.Flags().StringVarP(&opts.ContainerInput, "container", "r", "", "The container name")

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
		fmt.Println("No tasks running. We need a task ARN to get a task definition to get the logging information.")
		os.Exit(1)
		return "" // won't reach
	case 1:
		return t[0]
	default:
		return f.Prompt.Select("Multiple tasks running, please select one", t)
	}
}

func getLogGroup(f *factory.Factory, taskARN string) (targetLogConfig, error) {
	var details client.Container
	var config targetLogConfig

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

	opts.target = details

	definition, err := opts.client.DescribeTaskDefinition(details.TaskDefinitionARN)
	utils.CheckErr(err)

	for _, container := range definition.ContainerDefinitions {
		if aws.ToString(container.Name) == details.Name {
			if container.LogConfiguration == nil {
				return config, fmt.Errorf("no log configuration found")
			}
			if group, ok := container.LogConfiguration.Options["awslogs-group"]; ok {
				config.GroupName = group
			} else {
				// We have to have a log group
				return config, fmt.Errorf("no log group found")
			}

			if prefix, ok := container.LogConfiguration.Options["awslogs-stream-prefix"]; ok {
				config.StreamPrefix = prefix
			}
		}
	}

	return config, nil
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
