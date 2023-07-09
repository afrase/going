package internal

import (
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials/ssocreds"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc"
	"github.com/aws/aws-sdk-go-v2/service/ssooidc/types"

	"going/internal/factory"
	"going/internal/token"
	"going/internal/utils"
)

const oidcClientName = "going"
const oidcTokenGrantType = "urn:ietf:params:oauth:grant-type:device_code"

// CheckSSOLogin make sure we are logged in else does the full SSO login.
func CheckSSOLogin(f *factory.Factory) error {
	c, err := f.Config().Credentials.Retrieve(f.Context)
	if err == nil && (c.CanExpire && !c.Expired()) {
		return nil
	}

	var invalidToken *ssocreds.InvalidTokenError
	if errors.As(err, &invalidToken) || (c.CanExpire && c.Expired()) {
		t := getCacheToken(f)
		client := ssooidc.NewFromConfig(f.Config())

		if t.IsExpired() && t.RegistrationIsExpired() {
			if err := registerDevice(f, client, &t); err != nil {
				return err
			}
			if err := refreshToken(f, client, &t); err != nil {
				return err
			}
		} else if t.IsExpired() && !t.RegistrationIsExpired() {
			if err := refreshToken(f, client, &t); err != nil {
				return err
			}
		}

		if err := t.Write(); err != nil {
			return err
		}
	}

	return nil
}

// SSOLogin do a full SSO login
func SSOLogin(f *factory.Factory) error {
	t := getCacheToken(f)
	client := ssooidc.NewFromConfig(f.Config())
	if err := registerDevice(f, client, &t); err != nil {
		return err
	}
	if err := refreshToken(f, client, &t); err != nil {
		return err
	}

	if err := t.Write(); err != nil {
		return err
	}

	return nil
}

func getCacheToken(f *factory.Factory) token.SSOToken {
	// If this ever returns an error we have problems so just exit
	cacheFile, err := token.Filename(f.SelectedProfile().SSOStartURL)
	utils.CheckErr(err)

	// Read will return an error if the file doesn't exist, or we failed to parse it.
	// In either case we will just write a new token later.
	t, _ := token.Read(cacheFile)
	// These should always be equal
	t.StartUrl = f.SelectedProfile().SSOStartURL
	t.Region = f.SelectedProfile().SSORegion
	return t
}

func registerDevice(f *factory.Factory, client *ssooidc.Client, t *token.SSOToken) error {
	device, err := client.RegisterClient(f.Context, &ssooidc.RegisterClientInput{
		ClientName: aws.String(oidcClientName),
		ClientType: aws.String("public"),
	})
	if err != nil {
		return err
	}

	t.ClientId = aws.ToString(device.ClientId)
	t.ClientSecret = aws.ToString(device.ClientSecret)
	t.RegistrationExpiresAt = time.Unix(device.ClientSecretExpiresAt, 0)

	return nil
}

func refreshToken(f *factory.Factory, client *ssooidc.Client, t *token.SSOToken) error {
	deviceAuth, err := client.StartDeviceAuthorization(f.Context, &ssooidc.StartDeviceAuthorizationInput{
		ClientId:     aws.String(t.ClientId),
		ClientSecret: aws.String(t.ClientSecret),
		StartUrl:     aws.String(t.StartUrl),
	})
	if err != nil {
		return err
	}

	tokenInput := ssooidc.CreateTokenInput{
		ClientId:     aws.String(t.ClientId),
		ClientSecret: aws.String(t.ClientSecret),
		GrantType:    aws.String(oidcTokenGrantType),
		DeviceCode:   deviceAuth.DeviceCode,
	}

	err = utils.OpenUrlInBrowser(aws.ToString(deviceAuth.VerificationUriComplete))
	if err != nil {
		return err
	}

	fmt.Print("Waiting for authorization")
	for i := 0; i < 10; i++ {
		ct, err := client.CreateToken(f.Context, &tokenInput)
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
			t.AccessToken = aws.ToString(ct.AccessToken)
			t.ExpiresAt = time.Now().Add(time.Duration(ct.ExpiresIn) * time.Second)
			return nil
		}
	}

	return fmt.Errorf("varification took too long")
}
