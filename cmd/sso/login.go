package sso

import (
	"github.com/spf13/cobra"

	"going/internal"
	"going/internal/factory"
	"going/internal/utils"
)

func NewCmdLogin(f *factory.Factory) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Perform an SSO login",
		Run: func(cmd *cobra.Command, args []string) {
			err := internal.SSOLogin(f)
			utils.CheckErr(err)
		},
	}

	return cmd
}
