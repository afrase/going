package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"going/utils"
)

var (
	envOut bool
)

type AwsCreds struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Token     string `json:"token"`
}

// credsCmd represents the creds command
var credsCmd = &cobra.Command{
	Use:   "creds",
	Short: "Get AWS SSO credentials",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := utils.GetAwsConfig(ctx, awsProfile)
		utils.CheckErr(err)

		c, err := cfg.Credentials.Retrieve(ctx)
		utils.CheckErr(err)

		if envOut {
			fmt.Printf("AWS_ACCESS_KEY_ID=\"%s\"\n"+
				"AWS_SECRET_ACCESS_KEY=\"%s\"\n"+
				"AWS_SESSION_TOKEN=\"%s\"\n", c.AccessKeyID, c.SecretAccessKey, c.SessionToken)
		} else {
			creds := AwsCreds{
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

func init() {
	rootCmd.AddCommand(credsCmd)
	credsCmd.Flags().BoolVarP(&envOut, "env", "e", false, "Output in ENV")
}
