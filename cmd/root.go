package cmd

import (
	"github.com/spf13/cobra"

	"going/cmd/shell"
	"going/cmd/sso"
	"going/internal/factory"
)

func NewCmdRoot(version string) *cobra.Command {
	f := factory.New()

	cmd := &cobra.Command{
		Use:     "going",
		Short:   "A tool for working with AWS",
		Version: version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			f.Context = cmd.Context()
			if f.ProfileName == "" {
				p := f.Prompt.Select("Select a profile", f.LocalAWSConfig.ProfileNames())
				f.ProfileName = p
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&f.ProfileName, "profile", "p", "", "The AWS profile to use")

	cmd.AddCommand(shell.NewCmdShell(f))
	cmd.AddCommand(sso.NewCmdSSO(f))

	return cmd
}
