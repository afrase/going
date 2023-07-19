package shell

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/spf13/cobra"

	"going/internal"
	"going/internal/client"
	"going/internal/factory"
	"going/internal/utils"
)

type shellOptions struct {
	ClusterInput   string
	ServiceInput   string
	ContainerInput string

	ContainerDetails client.Container

	client *client.AWSClient
}

var opts = &shellOptions{}

func NewCmdShell(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "shell",
		Short: "Open a shell on a node in ECS",
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
			getContainerDetails(f, taskARN)

			fmt.Printf("cluster: %s service: %s container: %s\n",
				opts.ClusterInput, opts.ServiceInput, opts.ContainerDetails.Name)

			yes := f.Prompt.YesNoPrompt("Connect to the above container")
			if !yes {
				return
			}

			// taskId, _ := utils.Last(strings.Split(opts.TaskARN, "/"))
			// target := fmt.Sprintf("ecs:%s_%s_%s", opts.ClusterInput, taskId, aws.ToString(opts.ContainerDetails.RuntimeId))
			// err = ssmclient.ShellPluginSession(f.Config(), target)
			// utils.CheckErr(err)
			shell := exec.Command("aws",
				"ecs", "execute-command",
				"--profile", f.ProfileName,
				"--task", opts.ContainerDetails.TaskARN,
				"--cluster", opts.ContainerDetails.ClusterARN,
				"--container", opts.ContainerDetails.Name,
				"--command", "\"/bin/bash\"",
				"--interactive",
			)
			shell.Stdout = os.Stdout
			shell.Stderr = os.Stderr
			shell.Stdin = os.Stdin
			err = shell.Run()
			utils.CheckErr(err)
		},
	}

	cmd.Flags().StringVarP(&opts.ClusterInput, "cluster", "c", "", "ECS cluster")
	cmd.Flags().StringVarP(&opts.ServiceInput, "service", "s", "", "ECS service")
	cmd.Flags().StringVarP(&opts.ContainerInput, "container", "r", "", "ECS container")

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

	if len(t) > 0 {
		return t[0]
	}

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
}

func getContainerDetails(f *factory.Factory, taskARN string) {
	var details client.Container

	if opts.ContainerInput == "" {
		containers, err := opts.client.DescribeContainers(opts.ClusterInput, taskARN)
		utils.CheckErr(err)
		i := f.Prompt.CustomSelect("Select a container", containers, utils.ContainerTemplate, containerSearch(containers))
		details = containers[i]
	} else {
		c, err := opts.client.DescribeContainer(opts.ClusterInput, taskARN, opts.ContainerInput)
		utils.CheckErr(err)
		details = c
	}

	opts.ContainerDetails = details
	opts.ContainerInput = details.Name
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
