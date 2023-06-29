package utils

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/chzyer/readline"
)

type noBellStdout struct{}

func (n *noBellStdout) Write(p []byte) (int, error) {
	if len(p) == 1 && p[0] == readline.CharBell {
		return 0, nil
	}
	return readline.Stdout.Write(p)
}

func (n *noBellStdout) Close() error {
	return readline.Stdout.Close()
}

var NoBellStdout = &noBellStdout{}

// CheckErr If err is not nil then print to stderr and exist.
func CheckErr(err interface{}) {
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func GetAwsConfig(ctx context.Context, profile string) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
	if err != nil {
		return aws.Config{}, err
	}

	// retrieve the credentials to see if they are valid and not expired.
	credentials, err := cfg.Credentials.Retrieve(ctx)
	if err != nil || credentials.Expired() {
		cmd := exec.Command("aws", "sso", "login", "--profile", profile)
		cmd.Stdout = os.Stdout
		err = cmd.Run()
		if err != nil {
			return aws.Config{}, err
		}
	}

	return cfg, nil
}
