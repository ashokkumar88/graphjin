package provider

import (
	"context"
	"errors"

	plugin "github.com/dosco/graphjin/v2/plugin"
	"github.com/dosco/graphjin/v2/plugin/fs"
	jwt "github.com/golang-jwt/jwt"
)

// JWTConfig struct contains JWT authentication related config values used by
// the GraphJin service
type JWTConfig struct {
	// Provider can be one of auth0, firebase, jwks or other
	Provider string `jsonschema:"title=JWT Provider,enum=auth0,enum=firebase,enum=jwks,enum=other"`

	// The secret key used for signing and encrypting the JWT token
	Secret string `jsonschema:"title=JWT Secret Key"`

	// Public keys can be used instead of using a secret
	// PublicKeyFile points to the file containing the public key
	PubKeyFile string `mapstructure:"public_key_file" jsonschema:"title=Public Key File"`

	// Public key file type can be one of ecdsa or rsa
	PubKeyType string `mapstructure:"public_key_type" jsonschema:"title=Public Key File Type,enum=ecdsa,enum=rsa"`

	// Audience value that the JWT token needs to match
	Audience string `mapstructure:"audience" jsonschema:"title=Match Audience Value"`

	// Issuer value that the JWT token needs to match:
	// Example: http://my-domain.auth0.com
	Issuer string `mapstructure:"issuer" jsonschema:"title=Match Issuer Value,example=http://my-domain.auth0.com"`

	// Sets the url of the JWKS endpoint.
	// Example: https://YOUR_DOMAIN/.well-known/jwks.json
	JWKSURL string `mapstructure:"jwks_url" jsonschema:"title=JWKS Endpoint URL,example=https://YOUR_DOMAIN/.well-known/jwks.json"`

	// Sets in minutes interval between refreshes, overriding the adaptive token refreshing
	JWKSRefresh int `mapstructure:"jwks_refresh" jsonschema:"title=JWKS Refresh Timeout (minutes)"`

	// JWKSMinRefresh sets in minutes fallback value when tokens are refreshed, default
	// to 60 minutes
	JWKSMinRefresh int `mapstructure:"jwks_min_refresh" jsonschema:"title=JWKS Minimum Refresh Timeout (minutes)"`

	// FileSystem
	fs plugin.FS
}

// JWTProvider is the interface to define providers for doing JWT
// authentication.
type JWTProvider interface {
	KeyFunc() jwt.Keyfunc
	VerifyAudience(jwt.MapClaims) bool
	VerifyIssuer(jwt.MapClaims) bool
	SetContextValues(context.Context, jwt.MapClaims) (context.Context, error)
}

func NewProvider(config JWTConfig) (JWTProvider, error) {
	config.fs = fs.NewOsFS()
	switch config.Provider {
	case "auth0":
		return NewAuth0Provider(config)
	case "firebase":
		return NewFirebaseProvider(config)
	case "jwks":
		return NewJWKSProvider(config)
	default:
		return NewGenericProvider(config)
	}
}

func getKey(config JWTConfig) (interface{}, error) {
	var key interface{}
	secret := config.Secret
	publicKeyFile := config.PubKeyFile
	switch {
	case publicKeyFile != "":
		kd, err := config.fs.ReadFile(publicKeyFile)
		if err != nil {
			return nil, err
		}
		switch config.PubKeyType {
		case "ecdsa":
			key, err = jwt.ParseECPublicKeyFromPEM(kd)
		case "rsa":
			key, err = jwt.ParseRSAPublicKeyFromPEM(kd)
		default:
			key, err = jwt.ParseECPublicKeyFromPEM(kd)
		}
		if err != nil {
			return nil, err
		}
		//TODO: Log message informing that a public key will be used

	case secret != "":
		key = []byte(secret)
		//TODO: Log message informing that a secret will be used

	}
	if key == nil {
		return nil, errors.New("undefined key")
	}
	return key, nil
}

func (c *JWTConfig) SetFS(fs plugin.FS) {
	c.fs = fs
}
