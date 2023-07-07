package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/ssocreds"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc/types"
)

const oidcClientName = "going"
const oidcTokenGrantType = "urn:ietf:params:oauth:grant-type:device_code"

type SSOToken struct {
	StartUrl              string    `json:"startUrl"`
	Region                string    `json:"region"`
	AccessToken           string    `json:"accessToken"`
	ExpiresAt             time.Time `json:"expiresAt"`
	ClientId              string    `json:"clientId"`
	ClientSecret          string    `json:"clientSecret"`
	RegistrationExpiresAt time.Time `json:"registrationExpiresAt"`
}

func (s *SSOToken) IsExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}

func (s *SSOToken) RegistrationIsExpired() bool {
	return s.RegistrationExpiresAt.Before(time.Now())
}

func LoadSSOTokenCache(filename string) (SSOToken, error) {
	fileBytes, err := os.ReadFile(filename)
	if err != nil {
		return SSOToken{}, fmt.Errorf("failed to read cached SSO token file, %w", err)
	}

	var t SSOToken
	if err := json.Unmarshal(fileBytes, &t); err != nil {
		return SSOToken{}, fmt.Errorf("failed to parse cached SSO token file, %w", err)
	}

	return t, nil
}

// SSOLogin creates an SSO session token
func SSOLogin(ctx context.Context, profile AWSConfigProfile) (SSOToken, error) {
	var token SSOToken
	tokenCachePath, err := ssocreds.StandardCachedTokenFilepath(profile.SSOStartURL)
	if err != nil {
		return token, err
	}

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithSharedConfigProfile(profile.Name),
		config.WithCredentialsProvider(aws.AnonymousCredentials{}),
	)
	if err != nil {
		return token, err
	}

	client := ssooidc.NewFromConfig(cfg)

	if _, err := os.Stat(tokenCachePath); err != nil {
		token.Region = cfg.Region
		token.StartUrl = profile.SSOStartURL

		if err = registerDevice(ctx, client, &token); err != nil {
			return token, err
		}
		if err = refreshAccessToken(ctx, client, &token); err != nil {
			return token, err
		}
		if err = storeCacheFile(tokenCachePath, token, 0600); err != nil {
			return token, err
		}
		return token, nil
	}
	err = readCacheFile(tokenCachePath, &token)
	if err != nil {
		return token, err
	}

	if !token.IsExpired() {
		return token, nil
	}
	if !token.RegistrationIsExpired() {
		if err = refreshAccessToken(ctx, client, &token); err != nil {
			return token, err
		}
		if err = storeCacheFile(tokenCachePath, token, 0600); err != nil {
			return token, err
		}
		return token, nil
	} else {
		if err = registerDevice(ctx, client, &token); err != nil {
			return token, err
		}
		if err = refreshAccessToken(ctx, client, &token); err != nil {
			return token, err
		}
		if err = storeCacheFile(tokenCachePath, token, 0600); err != nil {
			return token, err
		}
		return token, nil
	}
}

func refreshAccessToken(ctx context.Context, client *ssooidc.Client, token *SSOToken) error {
	deviceAuth, err := client.StartDeviceAuthorization(ctx, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     aws.String(token.ClientId),
		ClientSecret: aws.String(token.ClientSecret),
		StartUrl:     aws.String(token.StartUrl),
	})
	if err != nil {
		return err
	}

	tokenInput := ssooidc.CreateTokenInput{
		ClientId:     aws.String(token.ClientId),
		ClientSecret: aws.String(token.ClientSecret),
		GrantType:    aws.String(oidcTokenGrantType),
		DeviceCode:   deviceAuth.DeviceCode,
	}

	err = openUrlInBrowser(aws.ToString(deviceAuth.VerificationUriComplete))
	if err != nil {
		return err
	}

	fmt.Print("Waiting for authorization")
	for i := 0; i < 10; i++ {
		t, err := client.CreateToken(ctx, &tokenInput)
		if err != nil {
			var tokenError *types.AuthorizationPendingException
			if errors.As(err, &tokenError) {
				fmt.Print(".")
				time.Sleep(3 * time.Second)
				continue
			} else {
				return err
			}
		} else {
			fmt.Print("\nSuccessfully logged in\n")
			token.AccessToken = aws.ToString(t.AccessToken)
			token.ExpiresAt = time.Now().Add(time.Duration(t.ExpiresIn) * time.Second)
			return nil
		}
	}

	return fmt.Errorf("varification took too long")
}

func registerDevice(ctx context.Context, client *ssooidc.Client, token *SSOToken) error {
	device, err := client.RegisterClient(ctx, &ssooidc.RegisterClientInput{
		ClientName: aws.String(oidcClientName),
		ClientType: aws.String("public"),
	})
	if err != nil {
		return err
	}
	token.ClientId = aws.ToString(device.ClientId)
	token.ClientSecret = aws.ToString(device.ClientSecret)
	token.RegistrationExpiresAt = time.Unix(device.ClientSecretExpiresAt, 0)
	return nil
}

func openUrlInBrowser(url string) error {
	fmt.Printf("Opening URL in default browser: %s\n", url)
	err := exec.Command("open", url).Start()
	if err != nil {
		return err
	}
	return nil
}

func readCacheFile(filename string, obj any) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	parser := json.NewDecoder(f)
	if err = parser.Decode(&obj); err != nil {
		return err
	}
	return nil
}

func storeCacheFile(filename string, obj interface{}, fileMode os.FileMode) (err error) {
	tmpFilename := filename + ".tmp-" + strconv.FormatInt(time.Now().UnixNano(), 10)
	if err := writeCacheFile(tmpFilename, fileMode, obj); err != nil {
		return err
	}

	if err := os.Rename(tmpFilename, filename); err != nil {
		return fmt.Errorf("failed to replace old cached SSO token file, %w", err)
	}

	return nil
}

func writeCacheFile(filename string, fileMode os.FileMode, obj interface{}) (err error) {
	var f *os.File
	f, err = os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, fileMode)
	if err != nil {
		return fmt.Errorf("failed to create cached SSO token file %w", err)
	}

	defer func() {
		closeErr := f.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("failed to close cached SSO token file, %w", closeErr)
		}
	}()

	encoder := json.NewEncoder(f)

	if err = encoder.Encode(obj); err != nil {
		return fmt.Errorf("failed to serialize cached SSO token, %w", err)
	}

	return nil
}
