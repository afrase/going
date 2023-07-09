package sso

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"going/internal/factory"
	"going/internal/utils"
)

var envOut bool

type awsCreds struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Token     string `json:"token"`
}

func NewCmdSSO(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sso",
		Short: "Get AWS SSO credentials",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := f.Config().Credentials.Retrieve(f.Context)
			utils.CheckErr(err)

			if envOut {
				fmt.Printf("AWS_ACCESS_KEY_ID=\"%s\"\n"+
					"AWS_SECRET_ACCESS_KEY=\"%s\"\n"+
					"AWS_SESSION_TOKEN=\"%s\"\n", c.AccessKeyID, c.SecretAccessKey, c.SessionToken)
			} else {
				creds := awsCreds{
					AccessKey: c.AccessKeyID,
					SecretKey: c.SecretAccessKey,
					Token:     c.SessionToken,
				}
				m, err := json.Marshal(creds)
				utils.CheckErr(err)
				fmt.Println(string(m))
			}
		},
	}

	cmd.Flags().BoolVarP(&envOut, "env", "e", false, "Output in ENV format")

	cmd.AddCommand(NewCmdLogin(f))
	cmd.AddCommand(NewCmdLogout(f))

	return cmd
}
