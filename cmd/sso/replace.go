package sso

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/spf13/cobra"

	"going/internal"
	"going/internal/factory"
	"going/internal/utils"
)

// TODO: This is terrible but it's a POC, rewrite this later.

func NewCmdReplace(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "replace",
		Example: "  going sso replace /project1/.env /project2/.env",
		Short:   "Replace ENV values for AWS credentials in environment files",
		Long: `This command attempts to update the AWS credentials in an environment file
with current values supplied by the AWS SDK. The keys that are updated are
AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY, and AWS_SESSION_TOKEN if they exist.
If no positional arguments are supplied then $PWD/.env is used.`,
		Run: func(cmd *cobra.Command, args []string) {
			err := internal.CheckSSOLogin(f)
			utils.CheckErr(err)

			if len(args) == 0 {
				pwd, err := os.Getwd()
				utils.CheckErr(err)
				envFile := filepath.Join(pwd, ".env")
				updateEnvFiles(f, envFile)
			} else {
				updateEnvFiles(f, args...)
			}
		},
	}

	return cmd
}

func updateEnvFiles(f *factory.Factory, paths ...string) {
	c, err := f.Config().Credentials.Retrieve(f.Context)
	utils.CheckErr(err)

	for _, path := range paths {
		err := updateEnvFile(path, c)
		utils.CheckErr(err)
	}
}

func updateEnvFile(path string, c aws.Credentials) error {
	stat, err := os.Stat(path)
	if err != nil {
		return err
	}

	f, err := os.OpenFile(path, os.O_RDWR, stat.Mode())
	if err != nil {
		return err
	}

	defer func(f *os.File) {
		closeErr := f.Close()
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}(f)

	r := updateEnv(f, c)
	_, err = f.Seek(0, 0)
	if err != nil {
		return err
	}

	_, err = f.WriteString(r)
	if err != nil {
		return err
	}

	return nil
}

func updateEnv(r io.Reader, creds aws.Credentials) string {
	out := strings.Builder{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "AWS_ACCESS_KEY_ID") {
			line = fmt.Sprintf("AWS_ACCESS_KEY_ID=\"%s\"", creds.AccessKeyID)
		} else if strings.HasPrefix(line, "AWS_SECRET_ACCESS_KEY") {
			line = fmt.Sprintf("AWS_SECRET_ACCESS_KEY=\"%s\"", creds.SecretAccessKey)
		} else if strings.HasPrefix(line, "AWS_SESSION_TOKEN") {
			line = fmt.Sprintf("AWS_SESSION_TOKEN=\"%s\"", creds.SessionToken)
		}
		out.WriteString(line + "\n")
	}

	return out.String()
}
