package token

import (
	"testing"
	"time"
)

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
