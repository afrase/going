package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"going/internal"
)

var (
	envOut bool
)

type awsCreds struct {
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	Token     string `json:"token"`
}

// credsCmd represents the creds command
var credsCmd = &cobra.Command{
	Use:   "creds",
	Short: "Get AWS SSO credentials",
	Run: func(cmd *cobra.Command, args []string) {
		profile, _ := awsConfig.GetProfile(awsProfile)
		cfg, err := internal.BuildAWSConfig(ctx, profile)
		internal.CheckErr(err)

		c, err := cfg.Credentials.Retrieve(ctx)
		internal.CheckErr(err)

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
			internal.CheckErr(err)
			fmt.Println(string(m))
		}
	},
}

func init() {
	rootCmd.AddCommand(credsCmd)
	credsCmd.Flags().BoolVarP(&envOut, "env", "e", false, "Output in ENV")
}
