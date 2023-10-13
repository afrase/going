package awsconfig

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/ini.v1"

	"going/internal/utils"
)

func TestFilename(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "in the users home directory",
			want: filepath.Join(utils.UserHomeDir(), ".aws", "config"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Filename(); got != tt.want {
				t.Errorf("Filename() = %v, profiles %v", got, tt.want)
			}
		})
	}
}

func TestNewConfig(t *testing.T) {
	tests := []struct {
		name        string
		configBytes []byte
		profiles    []Profile
	}{
		{
			name: "basic profile with sso_region",
			configBytes: []byte(`[profile test]
sso_start_url = https://my-sso-url
sso_region = us-east-1
region = eu-west-1`),
			profiles: []Profile{
				{Name: "test", SSOStartURL: "https://my-sso-url", SSORegion: "us-east-1"},
			},
		},
		{
			name: "complex config",
			configBytes: []byte(`region = eu-north-1
[default]
region = us-east-1
[foo]
sso_start_url = https://my-sso-url-foo
[profile test profile]
sso_start_url = https://my-sso-url-test
sso_region = us-east-2
[profile test2]
sso_start_url = https://my-sso-url-test2
region = us-west-2`),
			profiles: []Profile{
				{Name: "default", SSORegion: "us-east-1"},
				{Name: "test profile", SSOStartURL: "https://my-sso-url-test", SSORegion: "us-east-2"},
				{Name: "test2", SSOStartURL: "https://my-sso-url-test2", SSORegion: "us-west-2"},
			},
		},
		{
			name: "falls back to region when sso_region is missing",
			configBytes: []byte(`[profile test]
sso_start_url = https://my-sso-url
region = us-east-1`),
			profiles: []Profile{
				{Name: "test", SSOStartURL: "https://my-sso-url", SSORegion: "us-east-1"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, _ := ini.Load(tt.configBytes)
			result := NewConfig(cfg)
			if !reflect.DeepEqual(result.Profiles, tt.profiles) {
				t.Errorf("got=%+v, wanted=%+v", result.Profiles, tt.profiles)
			}
		})
	}
}

type mockConfigFileLoader struct {
	returnError bool
}

func (c *mockConfigFileLoader) Load(_ string) (*ini.File, error) {
	if c.returnError {
		return nil, fmt.Errorf("error loading")
	} else {
		return ini.Empty(), nil
	}
}

func TestRead(t *testing.T) {
	type args struct {
		loader   FileLoader
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr bool
	}{
		{
			name: "returns the config",
			args: args{
				loader: &mockConfigFileLoader{returnError: false},
			},
			want:    Config{},
			wantErr: false,
		},
		{
			name: "returns the error when loading fails",
			args: args{
				loader: &mockConfigFileLoader{returnError: true},
			},
			want:    Config{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Read(tt.args.loader, tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got.Profiles, tt.want.Profiles) {
				t.Errorf("Read() got = %v, want %v", got, tt.want)
			}
		})
	}
}
