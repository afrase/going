package cmd

import (
	"context"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"gopkg.in/ini.v1"

	"going/utils"
)

var awsProfile string
var ctx = context.Background()

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "going",
	Short: "A tool for working with AWS at LinkSquares",
	Long:  ``,
	// Run:   func(cmd *cobra.Command, args []string) {},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	utils.CheckErr(err)
}

func init() {
	cobra.OnInitialize(initProfile)
	rootCmd.PersistentFlags().StringVarP(&awsProfile, "profile", "p", "", "AWS profile to use")
}

func parseAwsProfiles() []string {
	usr, _ := user.Current()
	configFilePath := filepath.Join(usr.HomeDir, "/.aws/config")
	cfg, err := ini.Load(configFilePath)
	utils.CheckErr(err)

	var profiles []string
	for _, s := range cfg.SectionStrings() {
		if s == ini.DefaultSection {
			continue
		}

		s, _ = strings.CutPrefix(s, "profile ")
		profiles = append(profiles, s)
	}

	return profiles
}

func initProfile() {
	if awsProfile == "" {
		profiles := parseAwsProfiles()
		prompt := promptui.Select{Label: "Select a profile", Items: profiles, Stdout: utils.NoBellStdout}
		_, result, err := prompt.Run()
		utils.CheckErr(err)
		awsProfile = result
	}
}
