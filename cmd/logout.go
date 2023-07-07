package cmd

import (
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/ssocreds"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/spf13/cobra"

	"going/internal"
)

// logoutCmd represents the logout command
var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout of the current SSO session",
	Long: `Removes the locally stored SSO tokens from the client-side cache and sends an
API call to the IAM Identity Center service to invalidate the corresponding 
server-side IAM Identity Center sign in session.`,
	Run: func(cmd *cobra.Command, args []string) {
		profile, _ := awsConfig.GetProfile(awsProfile)
		filename, err := ssocreds.StandardCachedTokenFilepath(profile.SSOStartURL)
		internal.CheckErr(err)

		// Load the cached token to send as part of the logout request.
		token, err := internal.LoadSSOTokenCache(filename)
		internal.CheckErr(err)

		cfg, err := config.LoadDefaultConfig(ctx, config.WithCredentialsProvider(aws.AnonymousCredentials{}))
		internal.CheckErr(err)

		client := sso.NewFromConfig(cfg)
		// Ignore any errors because if the session is already invalid it will error.
		_, _ = client.Logout(ctx, &sso.LogoutInput{
			AccessToken: aws.String(token.AccessToken),
		})

		// sso.Client.Logout says it clears the cache file but doesn't so remove it.
		err = os.Remove(filename)
		internal.CheckErr(err)
	},
}

func init() {
	credsCmd.AddCommand(logoutCmd)
}
