package awsconfig

import (
	"fmt"
	"path/filepath"
	"strings"

	"gopkg.in/ini.v1"

	"going/internal/utils"
)

const (
	profilePrefix      = "profile "
	defaultProfileName = "default"
)

type Config struct {
	Profiles []Profile
	file     *ini.File
}

type Profile struct {
	Name        string
	SSOStartURL string
	SSORegion   string
}

type FileLoader interface {
	Load(string) (*ini.File, error)
}

type ConfigFileLoader struct{}

func (c *ConfigFileLoader) Load(filename string) (*ini.File, error) {
	return ini.Load(filename)
}

func NewConfig(rawCfg *ini.File) Config {
	cfg := Config{file: rawCfg}
	for _, section := range rawCfg.Sections() {
		sName := section.Name()
		// If the section isn't a profile or default then skip
		if !strings.HasPrefix(sName, profilePrefix) && sName != defaultProfileName {
			continue
		}

		cfg.Profiles = append(cfg.Profiles, Profile{
			Name:        strings.TrimPrefix(sName, profilePrefix),
			SSOStartURL: section.Key("sso_start_url").Value(),
			// If sso_region doesn't exist then fallback to region
			SSORegion: section.Key("sso_region").MustString(section.Key("region").Value()),
		})
	}

	return cfg
}

func Read(loader FileLoader, filename string) (Config, error) {
	rawCfg, err := loader.Load(filename)
	if err != nil {
		return Config{}, err
	}

	cfg := NewConfig(rawCfg)
	return cfg, nil
}

func (c *Config) ProfileNames() []string {
	var names []string
	for _, profile := range c.Profiles {
		names = append(names, profile.Name)
	}
	return names
}

func (c *Config) GetProfile(name string) (Profile, error) {
	for _, profile := range c.Profiles {
		if profile.Name == name {
			return profile, nil
		}
	}

	return Profile{}, fmt.Errorf("no profile named '%s'", name)
}

func Filename() string {
	return filepath.Join(utils.UserHomeDir(), ".aws", "config")
}
