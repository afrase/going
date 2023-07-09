package sso

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sso"
	"github.com/spf13/cobra"

	"going/internal/factory"
	"going/internal/token"
	"going/internal/utils"
)

func NewCmdLogout(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Logout of the current SSO session",
		Long: `Removes the locally stored SSO tokens from the client-side cache and sends an
API call to the IAM Identity Center service to invalidate the corresponding
server-side IAM Identity Center sign in session.`,
		Run: func(cmd *cobra.Command, args []string) {
			cacheFile, err := token.Filename(f.SelectedProfile().SSOStartURL)
			utils.CheckErr(err)

			t, _ := token.Read(cacheFile)
			client := sso.NewFromConfig(f.Config())

			// Ignore any errors because if the session is already invalid it will error.
			_, _ = client.Logout(f.Context, &sso.LogoutInput{
				AccessToken: aws.String(t.AccessToken),
			})

			// sso.Client.Logout says it clears the cache file but doesn't so remove it.
			err = t.Delete()
			utils.CheckErr(err)
		},
	}

	return cmd
}
