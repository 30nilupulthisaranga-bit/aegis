package infisical

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	infisicalSDK "github.com/infisical/go-sdk"
)

type Config struct {
	SiteURL string `yaml:"site_url"`

	ProjectID string `yaml:"project_id"`

	Environment string `yaml:"environment"`

	SecretPath string `yaml:"secret_path"`

	ClientIDSource     string `yaml:"client_id_source"`
	ClientSecretSource string `yaml:"client_secret_source"`

	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
}

type Client struct {
	sdk       infisicalSDK.InfisicalClientInterface
	cfg       Config
	cached    map[string]string
	lastFetch time.Time
	cacheTTL  time.Duration
}

func New(cfg Config, keychainGet func(string) (string, error)) (*Client, error) {
	if cfg.SiteURL == "" {
		cfg.SiteURL = "https://app.infisical.com"
	}
	if cfg.Environment == "" {
		cfg.Environment = "dev"
	}
	if cfg.SecretPath == "" {
		cfg.SecretPath = "/"
	}
	if cfg.ClientIDSource == "" {
		cfg.ClientIDSource = "keychain"
	}
	if cfg.ClientSecretSource == "" {
		cfg.ClientSecretSource = "keychain"
	}

	clientID, err := resolveCredential("INFISICAL_CLIENT_ID", "infisical_client_id",
		cfg.ClientIDSource, cfg.ClientID, keychainGet)
	if err != nil {
		return nil, fmt.Errorf("infisical client ID: %w", err)
	}

	clientSecret, err := resolveCredential("INFISICAL_CLIENT_SECRET", "infisical_client_secret",
		cfg.ClientSecretSource, cfg.ClientSecret, keychainGet)
	if err != nil {
		return nil, fmt.Errorf("infisical client secret: %w", err)
	}

	sdk := infisicalSDK.NewInfisicalClient(context.Background(), infisicalSDK.Config{
		SiteUrl:          cfg.SiteURL,
		AutoTokenRefresh: true,
	})

	_, err = sdk.Auth().UniversalAuthLogin(clientID, clientSecret)
	if err != nil {
		return nil, fmt.Errorf("infisical auth failed: %w\n\n"+
			"Check your INFISICAL_CLIENT_ID and INFISICAL_CLIENT_SECRET.\n"+
			"Setup guide: https://infisical.com/docs/documentation/platform/identities/universal-auth", err)
	}

	c := &Client{
		sdk:      sdk,
		cfg:      cfg,
		cacheTTL: 5 * time.Minute,
	}

	if err := c.refresh(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Secrets() (map[string]string, error) {
	if time.Since(c.lastFetch) > c.cacheTTL {
		if err := c.refresh(); err != nil {

			fmt.Fprintf(os.Stderr, "[aegis] WARNING: Infisical refresh failed: %v (using cached secrets)\n", err)
		}
	}
	return c.cached, nil
}

func (c *Client) Get(key string) (string, error) {
	secrets, err := c.Secrets()
	if err != nil {
		return "", err
	}
	val, ok := secrets[key]
	if !ok {
		return "", fmt.Errorf("secret %q not found in Infisical project %s / env %s",
			key, c.cfg.ProjectID, c.cfg.Environment)
	}
	return val, nil
}

func (c *Client) LoadIntoEnv() (int, error) {
	secrets, err := c.Secrets()
	if err != nil {
		return 0, err
	}
	for k, v := range secrets {
		os.Setenv(k, v)
	}
	return len(secrets), nil
}

func (c *Client) refresh() error {
	secrets, err := c.sdk.Secrets().List(infisicalSDK.ListSecretsOptions{
		ProjectID:   c.cfg.ProjectID,
		Environment: c.cfg.Environment,
		SecretPath:  c.cfg.SecretPath,

		AttachToProcessEnv: false,
	})
	if err != nil {
		return fmt.Errorf("fetching secrets from Infisical: %w", err)
	}

	m := make(map[string]string, len(secrets))
	for _, s := range secrets {
		m[s.SecretKey] = s.SecretValue
	}
	c.cached = m
	c.lastFetch = time.Now()
	return nil
}

func resolveCredential(envKey, keychainKey, source, configValue string, keychainGet func(string) (string, error)) (string, error) {
	switch strings.ToLower(source) {
	case "env":
		val := os.Getenv(envKey)
		if val == "" {
			return "", fmt.Errorf("env var %s is not set", envKey)
		}
		return val, nil

	case "keychain":
		if keychainGet == nil {
			return "", fmt.Errorf("keychain not available")
		}
		val, err := keychainGet(keychainKey)
		if err != nil {
			return "", fmt.Errorf("keychain key %q not found — run: aegis infisical setup", keychainKey)
		}
		return val, nil

	case "config":
		return "", fmt.Errorf("config source is disabled for security; use keychain or env")

	default:

		if keychainGet != nil {
			if val, err := keychainGet(keychainKey); err == nil && val != "" {
				return val, nil
			}
		}
		if val := os.Getenv(envKey); val != "" {
			return val, nil
		}
		return "", fmt.Errorf(
			"could not find %s — run: aegis infisical setup\n"+
				"Or set env var %s, or add to aegis.yaml:\n"+
				"  infisical:\n    client_id_source: env", keychainKey, envKey)
	}
}
