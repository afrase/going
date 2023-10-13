package factory

import (
	"context"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"

	"going/internal/awsconfig"
	"going/internal/utils"
)

type Factory struct {
	Prompt         utils.Prompt
	LocalAWSConfig awsconfig.Config
	Context        context.Context
	ProfileName    string

	config          aws.Config
	selectedProfile awsconfig.Profile
}

func New() *Factory {
	awsCfg, err := awsconfig.Read(&awsconfig.ConfigFileLoader{}, awsconfig.Filename())
	utils.CheckErr(err)
	f := &Factory{
		Prompt:         utils.Prompter{},
		LocalAWSConfig: awsCfg,
	}
	return f
}

func (f *Factory) Config() aws.Config {
	if !reflect.ValueOf(f.config).IsZero() {
		return f.config
	}

	cfg, err := config.LoadDefaultConfig(f.Context,
		config.WithSharedConfigProfile(f.ProfileName),
	)
	utils.CheckErr(err)
	f.config = cfg
	return cfg
}

func (f *Factory) SelectedProfile() awsconfig.Profile {
	if f.selectedProfile != (awsconfig.Profile{}) {
		return f.selectedProfile
	}
	profile, err := f.LocalAWSConfig.GetProfile(f.ProfileName)
	utils.CheckErr(err)
	f.selectedProfile = profile
	return profile
}
