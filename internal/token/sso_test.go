package token

import (
	"reflect"
	"testing"
	"time"
)

func TestFilename(t *testing.T) {
	type args struct {
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Filename(tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Filename() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Filename() got = %v, want %v", got, tt.want)
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
		want    SSOToken
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

func TestSSOToken_IsExpired(t *testing.T) {
	type fields struct {
		ExpiresAt time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "blank",
			fields: fields{},
			want:   true,
		},
		{
			name:   "in the future",
			fields: fields{ExpiresAt: time.Now().Add(time.Minute)},
			want:   false,
		},
		{
			name:   "in the past",
			fields: fields{ExpiresAt: time.Now().Add(-time.Minute)},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SSOToken{
				ExpiresAt: tt.fields.ExpiresAt,
			}
			if got := s.IsExpired(); got != tt.want {
				t.Errorf("IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSSOToken_RegistrationIsExpired(t *testing.T) {
	type fields struct {
		RegistrationExpiresAt time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "blank",
			fields: fields{},
			want:   true,
		},
		{
			name:   "in the future",
			fields: fields{RegistrationExpiresAt: time.Now().Add(time.Minute)},
			want:   false,
		},
		{
			name:   "in the past",
			fields: fields{RegistrationExpiresAt: time.Now().Add(-time.Minute)},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SSOToken{
				RegistrationExpiresAt: tt.fields.RegistrationExpiresAt,
			}
			if got := s.RegistrationIsExpired(); got != tt.want {
				t.Errorf("RegistrationIsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSSOToken_Write(t *testing.T) {
	type fields struct {
		StartUrl              string
		Region                string
		AccessToken           string
		ExpiresAt             time.Time
		ClientId              string
		ClientSecret          string
		RegistrationExpiresAt time.Time
		filename              string
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &SSOToken{
				StartUrl:              tt.fields.StartUrl,
				Region:                tt.fields.Region,
				AccessToken:           tt.fields.AccessToken,
				ExpiresAt:             tt.fields.ExpiresAt,
				ClientId:              tt.fields.ClientId,
				ClientSecret:          tt.fields.ClientSecret,
				RegistrationExpiresAt: tt.fields.RegistrationExpiresAt,
				filename:              tt.fields.filename,
			}
			if err := s.Write(); (err != nil) != tt.wantErr {
				t.Errorf("Write() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
