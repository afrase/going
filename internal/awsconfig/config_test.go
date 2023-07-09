package awsconfig

import (
	"path/filepath"
	"reflect"
	"testing"

	"gopkg.in/ini.v1"

	"going/internal/utils"
)

func TestConfig_GetProfile(t *testing.T) {
	type fields struct {
		Profiles []Profile
		file     *ini.File
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Profile
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Profiles: tt.fields.Profiles,
				file:     tt.fields.file,
			}
			got, err := c.GetProfile(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetProfile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetProfile() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfig_ProfileNames(t *testing.T) {
	type fields struct {
		Profiles []Profile
		file     *ini.File
	}
	tests := []struct {
		name   string
		fields fields
		want   []string
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				Profiles: tt.fields.Profiles,
				file:     tt.fields.file,
			}
			if got := c.ProfileNames(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ProfileNames() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
				t.Errorf("Filename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRead(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name    string
		args    args
		want    Config
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Read(tt.args.filename)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Read() got = %v, want %v", got, tt.want)
			}
		})
	}
}
