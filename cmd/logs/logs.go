package logs

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
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
	Hours          int

	client *client.AWSClient
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
			group, err := getLogGroup(f, taskARN)
			utils.CheckErr(err)

			logClient := cloudwatchlogs.NewFromConfig(f.Config())
			pager := cloudwatchlogs.NewFilterLogEventsPaginator(logClient, &cloudwatchlogs.FilterLogEventsInput{
				LogGroupName: aws.String(group),
				StartTime:    aws.Int64(time.Now().Add(-time.Duration(opts.Hours) * time.Hour).UnixMilli()),
			})

			for pager.HasMorePages() {
				result, err := pager.NextPage(f.Context)
				utils.CheckErr(err)

				for _, event := range result.Events {
					fmt.Printf("%s [%s] %s\n",
						aws.ToString(event.LogStreamName),
						time.UnixMilli(aws.ToInt64(event.Timestamp)),
						aws.ToString(event.Message))
				}
			}

		},
	}

	cmd.Flags().IntVarP(&opts.Hours, "hours", "t", 1, "Number of hours back to filter logs")
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

func getLogGroup(f *factory.Factory, taskARN string) (string, error) {
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

	definition, err := opts.client.DescribeTaskDefinition(details.TaskDefinitionARN)
	utils.CheckErr(err)
	for _, container := range definition.ContainerDefinitions {
		if aws.ToString(container.Name) == details.Name {
			for k, v := range container.LogConfiguration.Options {
				if k == "awslogs-group" {
					return v, nil
				}
			}
		}
	}

	return "", fmt.Errorf("no log group found")
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
