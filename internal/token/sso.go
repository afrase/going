package token

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"going/internal/utils"
)

// SSOToken a token representing the cached SSO credentials
type SSOToken struct {
	StartUrl              string    `json:"startUrl"`
	Region                string    `json:"region"`
	AccessToken           string    `json:"accessToken"`
	ExpiresAt             time.Time `json:"expiresAt"`
	ClientId              string    `json:"clientId"`
	ClientSecret          string    `json:"clientSecret"`
	RegistrationExpiresAt time.Time `json:"registrationExpiresAt"`

	filename string
}

func (s *SSOToken) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}

func (s *SSOToken) RegistrationIsExpired() bool {
	return s.RegistrationExpiresAt.Before(time.Now())
}

func (s *SSOToken) Write() error {
	err := utils.StoreCacheFile(s.filename, s, 0600)
	if err != nil {
		return err
	}
	return nil
}

func (s *SSOToken) Delete() error {
	err := os.Remove(s.filename)
	if err != nil {
		return err
	}
	return nil
}

func Read(filename string) (SSOToken, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return SSOToken{}, fmt.Errorf("failed to read cached SSO token file, %w", err)
	}

	t := SSOToken{filename: filename}
	if err := json.Unmarshal(fileBytes, &t); err != nil {
		return SSOToken{}, fmt.Errorf("failed to parse cached SSO token file, %w", err)
	}

	return t, nil
}

func Filename(key string) (string, error) {
	homeDir := utils.UserHomeDir()
	if len(homeDir) == 0 {
		return "", fmt.Errorf("unable to get USER's home directory for cached token")
	}
	hash := sha1.New()
	if _, err := hash.Write([]byte(key)); err != nil {
		return "", fmt.Errorf("unable to compute cached token filepath key SHA1 hash, %w", err)
	}

	cacheFilename := strings.ToLower(hex.EncodeToString(hash.Sum(nil))) + ".json"
	return filepath.Join(homeDir, ".aws", "sso", "cache", cacheFilename), nil
}
