package shell

import (
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/spf13/cobra"

	"going/internal"
	"going/internal/factory"
	"going/internal/utils"
)

type shellOptions struct {
	Cluster   string
	Service   string
	Container string
	TaskARN   string

	client *ecs.Client
}

var opts = &shellOptions{}

func NewCmdShell(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Open a shell on a node in ECS",
		Run: func(cmd *cobra.Command, args []string) {
			// Must be logged in
			err := internal.CheckSSOLogin(f)
			utils.CheckErr(err)

			if opts.Cluster == "" {
				opts.Cluster = promptForCluster(f)
			}
			if opts.Service == "" {
				opts.Service = promptForService(f)
			}

			opts.TaskARN = getTaskArn(f)

			if opts.Container == "" {
				opts.Container = promptForContainer(f)
			}

			fmt.Printf("cluster: %s service: %s container: %s\n", opts.Cluster, opts.Service, opts.Container)
			yes := f.Prompt.YesNoPrompt("Connect to the above instance")
			if !yes {
				return
			}

			shell := exec.Command("aws",
				"ecs", "execute-command",
				"--profile", f.ProfileName,
				"--task", opts.TaskARN,
				"--cluster", opts.Cluster,
				"--container", opts.Container,
				"--command", "\"/bin/bash\"",
				"--interactive",
			)
			shell.Stdout = os.Stdout
			shell.Stderr = os.Stderr
			shell.Stdin = os.Stdin
			err = shell.Run()
			utils.CheckErr(err)
		},
		PreRun: func(cmd *cobra.Command, args []string) {
			opts.client = ecs.NewFromConfig(f.Config())
		},
	}

	cmd.Flags().StringVarP(&opts.Cluster, "cluster", "c", "", "ECS cluster")
	cmd.Flags().StringVarP(&opts.Service, "service", "s", "", "ECS service")
	cmd.Flags().StringVarP(&opts.Container, "container", "r", "", "ECS container")

	return cmd
}

func promptForCluster(f *factory.Factory) string {
	pager := ecs.NewListClustersPaginator(opts.client, &ecs.ListClustersInput{})

	var clusters []string
	for pager.HasMorePages() {
		result, err := pager.NextPage(f.Context)
		utils.CheckErr(err)
		for _, arn := range result.ClusterArns {
			parts := strings.Split(arn, "/")
			if cluster, ok := utils.Last(parts); ok {
				clusters = append(clusters, cluster)
			}
		}
	}

	return f.Prompt.Select("Select a cluster", clusters)
}

func promptForService(f *factory.Factory) string {
	pager := ecs.NewListServicesPaginator(opts.client, &ecs.ListServicesInput{
		Cluster: aws.String(opts.Cluster),
	})

	var services []string
	for pager.HasMorePages() {
		result, err := pager.NextPage(f.Context)
		utils.CheckErr(err)
		for _, a := range result.ServiceArns {
			parts := strings.Split(a, "/")
			if service, ok := utils.Last(parts); ok {
				services = append(services, service)
			}
		}
	}

	return f.Prompt.Select("Select a service", services)
}

func getTaskArn(f *factory.Factory) string {
	result, err := opts.client.ListTasks(f.Context, &ecs.ListTasksInput{
		Cluster:     aws.String(opts.Cluster),
		ServiceName: aws.String(opts.Service),
	})
	utils.CheckErr(err)

	return result.TaskArns[0]
}

func promptForContainer(f *factory.Factory) string {
	result, err := opts.client.DescribeTasks(f.Context, &ecs.DescribeTasksInput{
		Cluster: aws.String(opts.Cluster),
		Tasks:   []string{opts.TaskARN},
	})
	utils.CheckErr(err)

	var tasks []string
	for _, c := range result.Tasks[0].Containers {
		tasks = append(tasks, aws.ToString(c.Name))
	}

	sort.Strings(tasks)
	return f.Prompt.Select("Select a container", tasks)
}
