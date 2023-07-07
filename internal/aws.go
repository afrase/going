package internal

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"gopkg.in/ini.v1"
)

const profilePrefix = "profile "

// AWSConfig represents the ~/.aws/config file
type AWSConfig struct {
	Profiles []AWSConfigProfile
}

type AWSConfigProfile struct {
	Name        string
	Region      string
	SSOStartURL string
}

func BuildAWSConfig(ctx context.Context, profile AWSConfigProfile) (aws.Config, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile.Name))
	if err != nil {
		return aws.Config{}, err
	}

	// retrieve the credentials to see if they are valid and not expired.
	credentials, err := cfg.Credentials.Retrieve(ctx)
	if err != nil || credentials.Expired() {
		_, err := SSOLogin(ctx, profile)
		if err != nil {
			return aws.Config{}, err
		}
	}

	return cfg, nil
}

func ParseAWSConfig() AWSConfig {
	rawCfg := readAWSConfig()

	var cfg AWSConfig
	for _, section := range rawCfg.Sections() {
		sectionName := section.Name()
		if !strings.HasPrefix(sectionName, profilePrefix) && sectionName != "default" {
			continue
		}

		cfg.Profiles = append(cfg.Profiles, AWSConfigProfile{
			Name:        strings.TrimPrefix(sectionName, profilePrefix),
			Region:      section.Key("region").String(),
			SSOStartURL: section.Key("sso_start_url").String(),
		})
	}

	return cfg
}

func (a *AWSConfig) GetProfile(name string) (AWSConfigProfile, error) {
	for _, profile := range a.Profiles {
		if profile.Name == name {
			return profile, nil
		}
	}
	return AWSConfigProfile{}, fmt.Errorf("no profile named '%s'", name)
}

func (a *AWSConfig) ProfileNames() []string {
	var names []string
	for _, profile := range a.Profiles {
		names = append(names, profile.Name)
	}
	return names
}

func readAWSConfig() *ini.File {
	configFilePath := config.DefaultSharedConfigFilename()
	cfg, err := ini.Load(configFilePath)
	CheckErr(err)

	return cfg
}
