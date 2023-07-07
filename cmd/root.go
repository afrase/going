package cmd

import (
	"context"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"going/internal"
)

var awsProfile string
var awsConfig internal.AWSConfig
var ctx = context.Background()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "going",
	Short: "A tool for working with AWS at LinkSquares",
	// Run: func(cmd *cobra.Command, args []string) {
	// 	profile, err := config.LoadSharedConfigProfile(ctx, awsProfile)
	// 	internal.CheckErr(err)
	// 	fmt.Printf("%v\n", profile)
	// },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	internal.CheckErr(err)
}

func init() {
	cobra.OnInitialize(initProfile)
	rootCmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", "", "AWS profile to use")
}

func initProfile() {
	awsConfig = internal.ParseAWSConfig()

	if awsProfile == "" {
		prompt := promptui.Select{Label: "Select a profile", Items: awsConfig.ProfileNames(), Stdout: internal.NoBellStdout}
		_, result, err := prompt.Run()
		internal.CheckErr(err)
		awsProfile = result
	}
}
