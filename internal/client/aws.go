package client

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"

	"going/internal/utils"
)

const (
	// The format for setting the Target of an ECS container in the SSM Session.
	ecsTargetFormat    = "ecs:%s_%s_%s"
	groupServicePrefix = "service:"
)

type AWSClient struct {
	ctx       context.Context
	ecsClient *ecs.Client
}

type Cluster struct {
	Name string
	ARN  string
}

type Service struct {
	Name string
	ARN  string
}

type Task struct {
	ARN         string
	ClusterARN  string
	ClusterName string
	ServiceName string
	Containers  []Container
}

type Container struct {
	Name        string
	ARN         string
	ClusterARN  string
	ClusterName string
	ServiceName string
	TaskARN     string
	RuntimeID   string

	Health     string
	LastStatus string
}

func New(ctx context.Context, cfg aws.Config) *AWSClient {
	return &AWSClient{
		ctx:       ctx,
		ecsClient: ecs.NewFromConfig(cfg),
	}
}

// ListClusters returns all clusters the current AWS profile has access to.
func (c *AWSClient) ListClusters() ([]Cluster, error) {
	pager := ecs.NewListClustersPaginator(c.ecsClient, &ecs.ListClustersInput{})

	var clusters []Cluster
	for pager.HasMorePages() {
		result, err := pager.NextPage(c.ctx)
		if err != nil {
			return nil, err
		}

		for _, a := range result.ClusterArns {
			if name, ok := utils.Last(strings.Split(a, "/")); ok {
				clusters = append(clusters, Cluster{ARN: a, Name: name})
			}
		}
	}

	return clusters, nil
}

// ListServices returns all service for the given cluster.
func (c *AWSClient) ListServices(cluster string) ([]Service, error) {
	pager := ecs.NewListServicesPaginator(c.ecsClient, &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
	})

	var services []Service
	for pager.HasMorePages() {
		result, err := pager.NextPage(c.ctx)
		if err != nil {
			return nil, err
		}

		for _, a := range result.ServiceArns {
			if name, ok := utils.Last(strings.Split(a, "/")); ok {
				services = append(services, Service{ARN: a, Name: name})
			}
		}
	}

	return services, nil
}

// ListTasks returns all task ARNs for the cluster with the given service name.
func (c *AWSClient) ListTasks(cluster string, service string) ([]string, error) {
	pager := ecs.NewListTasksPaginator(c.ecsClient, &ecs.ListTasksInput{
		Cluster:     aws.String(cluster),
		ServiceName: aws.String(service),
	})

	var tasks []string
	for pager.HasMorePages() {
		result, err := pager.NextPage(c.ctx)
		if err != nil {
			return nil, err
		}

		for _, a := range result.TaskArns {
			tasks = append(tasks, a)
		}
	}

	return tasks, nil
}

// DescribeTasks returns all tasks in the cluster for the given task ARNs.
func (c *AWSClient) DescribeTasks(cluster string, taskARNs ...string) ([]Task, error) {
	result, err := c.ecsClient.DescribeTasks(c.ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   taskARNs,
	})
	if err != nil {
		return nil, err
	}

	var tasks []Task
	for _, task := range result.Tasks {
		clusterName, _ := utils.Last(strings.Split(aws.ToString(task.ClusterArn), "/"))
		// The Group appears to be the name of the service prefixed with "service:".
		serviceName := strings.TrimPrefix(aws.ToString(task.Group), groupServicePrefix)

		t := Task{
			ARN:         aws.ToString(task.TaskArn),
			ClusterARN:  aws.ToString(task.ClusterArn),
			ClusterName: clusterName,
			ServiceName: serviceName,
		}

		for _, container := range task.Containers {
			t.Containers = append(t.Containers, Container{
				Name:        aws.ToString(container.Name),
				ARN:         aws.ToString(container.ContainerArn),
				TaskARN:     aws.ToString(container.TaskArn),
				ClusterARN:  aws.ToString(task.ClusterArn),
				ClusterName: clusterName,
				ServiceName: serviceName,
				RuntimeID:   aws.ToString(container.RuntimeId),
				Health:      string(container.HealthStatus),
				LastStatus:  aws.ToString(container.LastStatus),
			})
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}

// DescribeTask returns the first task.
func (c *AWSClient) DescribeTask(cluster string, taskARN string) (Task, error) {
	result, err := c.DescribeTasks(cluster, taskARN)
	if err != nil {
		return Task{}, err
	}

	if len(result) <= 0 {
		return Task{}, fmt.Errorf("no tasks found for cluster %s with task ARN %s", cluster, taskARN)
	}

	return result[0], nil
}

// DescribeContainers get details about the containers for the given cluster and task ARN.
func (c *AWSClient) DescribeContainers(cluster string, taskARN string) ([]Container, error) {
	result, err := c.DescribeTask(cluster, taskARN)
	if err != nil {
		return nil, err
	}

	return result.Containers, nil
}

// DescribeContainer get details about the container for the given cluster and task ARN with the given name.
func (c *AWSClient) DescribeContainer(cluster string, taskARN string, name string) (Container, error) {
	containers, err := c.DescribeContainers(cluster, taskARN)
	if err != nil {
		return Container{}, err
	}

	for _, container := range containers {
		if strings.EqualFold(name, container.Name) {
			return container, nil
		}
	}

	return Container{}, fmt.Errorf("no container with name '%s' in cluster '%s'", name, cluster)
}

// UpdateService calls the ecs.Client.UpdateService method.
func (c *AWSClient) UpdateService(params *ecs.UpdateServiceInput) error {
	_, err := c.ecsClient.UpdateService(c.ctx, params)
	if err != nil {
		return err
	}
	return nil
}

// ExecuteCommand calls the ecs.Client.ExecuteCommand method.
func (c *AWSClient) ExecuteCommand(params *ecs.ExecuteCommandInput) (*ecs.ExecuteCommandOutput, error) {
	output, err := c.ecsClient.ExecuteCommand(c.ctx, params)
	if err != nil {
		return &ecs.ExecuteCommandOutput{}, err
	}
	return output, nil
}

// SSMTarget returns a string that can be used by SSM to target this container.
// The documentation for SSM sessions only ever shows the target being an
// instance ID. By reading through the AWS CLI source I was able to find that
// they call the session-manager-plugin using a special string format for the
// target of a container.
func (c *Container) SSMTarget() (string, error) {
	if c.RuntimeID == "" {
		return "", fmt.Errorf("container has no runtime ID, it most likely is still starting")
	}

	taskId, _ := utils.Last(strings.Split(c.TaskARN, "/"))
	return fmt.Sprintf(ecsTargetFormat, c.ClusterName, taskId, c.RuntimeID), nil
}
