package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"

	"going/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

var (
	cluster   string
	service   string
	container string
	taskArn   string
)

// shellCmd represents the shell command
var shellCmd = &cobra.Command{
	Use:    "shell",
	Short:  "Open a shell on a node in ECS",
	PreRun: preRun,
	Run:    run,
}

func run(_ *cobra.Command, _ []string) {
	fmt.Printf("cluster: %s service: %s container: %s\n", cluster, service, container)

	prompt := promptui.Prompt{Label: "Connect to the above instance", IsConfirm: true, Default: "n"}
	validate := func(s string) error {
		if len(s) == 1 && strings.Contains("YyNn", s) || prompt.Default != "" && len(s) == 0 {
			return nil
		}
		return errors.New("invalid input")
	}
	prompt.Validate = validate
	_, err := prompt.Run()
	aborted := errors.Is(err, promptui.ErrAbort)
	if aborted || err != nil {
		return
	}

	shell := exec.Command("aws",
		"ecs", "execute-command",
		"--profile", awsProfile,
		"--task", taskArn,
		"--cluster", cluster,
		"--container", container,
		"--command", "\"/bin/bash\"",
		"--interactive",
	)
	shell.Stdout = os.Stdout
	shell.Stderr = os.Stderr
	shell.Stdin = os.Stdin
	err = shell.Run()
	utils.CheckErr(err)
}

func init() {
	rootCmd.AddCommand(shellCmd)
	shellCmd.Flags().StringVarP(&cluster, "cluster", "c", "", "ECS cluster")
	shellCmd.Flags().StringVarP(&service, "service", "s", "", "ECS service")
	shellCmd.Flags().StringVarP(&container, "container", "r", "", "ECS container")
}

func preRun(_ *cobra.Command, _ []string) {
	cfg, err := utils.GetAwsConfig(ctx, awsProfile)
	utils.CheckErr(err)

	if cluster == "" {
		cluster = promptForCluster(cfg)
	}
	if service == "" {
		service = promptForService(cfg)
	}

	taskArn = getTaskArn(cfg)

	if container == "" {
		container = promptForContainer(cfg)
	}
}

func promptForCluster(cfg aws.Config) string {
	svc := ecs.NewFromConfig(cfg)
	output, err := svc.ListClusters(ctx, &ecs.ListClustersInput{})
	utils.CheckErr(err)

	var clusters []string
	for _, arn := range output.ClusterArns {
		parts := strings.Split(arn, "/")
		clusters = append(clusters, parts[len(parts)-1])
	}

	// sort.Strings(clusters)
	prompt := promptui.Select{Label: "Select a cluster", Items: clusters, Stdout: utils.NoBellStdout}
	_, result, err := prompt.Run()
	utils.CheckErr(err)
	return result
}

func promptForService(cfg aws.Config) string {
	svc := ecs.NewFromConfig(cfg)
	output, err := svc.ListServices(ctx,
		&ecs.ListServicesInput{
			Cluster: aws.String(cluster),
		},
	)
	utils.CheckErr(err)

	var services []string
	for _, a := range output.ServiceArns {
		parts := strings.Split(a, "/")
		services = append(services, parts[len(parts)-1])
	}

	// sort.Strings(services)
	prompt := promptui.Select{Label: "Select a service", Items: services, Stdout: utils.NoBellStdout}
	_, result, err := prompt.Run()
	utils.CheckErr(err)
	return result
}

func getTaskArn(cfg aws.Config) string {
	svc := ecs.NewFromConfig(cfg)
	result, err := svc.ListTasks(ctx,
		&ecs.ListTasksInput{
			Cluster:     aws.String(cluster),
			ServiceName: aws.String(service),
		},
	)
	utils.CheckErr(err)

	return result.TaskArns[0]
}

func promptForContainer(cfg aws.Config) string {
	svc := ecs.NewFromConfig(cfg)
	output, err := svc.DescribeTasks(ctx,
		&ecs.DescribeTasksInput{
			Cluster: aws.String(cluster),
			Tasks:   []string{taskArn},
		},
	)
	utils.CheckErr(err)

	var tasks []string
	for _, c := range output.Tasks[0].Containers {
		tasks = append(tasks, aws.ToString(c.Name))
	}

	sort.Strings(tasks)
	prompt := promptui.Select{Label: "Select a container", Items: tasks, Stdout: utils.NoBellStdout}
	_, result, err := prompt.Run()
	utils.CheckErr(err)
	return result
}
