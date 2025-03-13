package secretsmngr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	res "github.com/HayoVanLoon/go-slimfig/resolver"

	"github.com/HayoVanLoon/go-slimfig/resolver/base"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

var _ res.Resolver = *new(resolver)

type resolver struct {
	base.Resolver
}

func (r resolver) Matches(reference string) bool {
	return validName(reference) != ""
}

// JSONResolver returns a Secret Manager resolver for secrets containing JSON
// objects.
func JSONResolver(ctx context.Context) (res.Resolver, error) {
	return Resolver(ctx, json.Unmarshal)
}

// Resolver returns a Secrets Manager resolver for secrets that can be
// unmarshalled into maps using the provided function.
func Resolver(ctx context.Context, unmarshal base.Unmarshaller) (res.Resolver, error) {
	var opts []func(*config.LoadOptions) error
	if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}
	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("could not load default config: %w", err)
	}
	cfg.HTTPClient = http.DefaultClient
	c := secretsmanager.NewFromConfig(cfg)
	return WithClient(c, unmarshal), nil
}

// WithClient returns a Secrets Manager resolver with the given client and
// unmarshaller.
func WithClient(c *secretsmanager.Client, unmarshal base.Unmarshaller) res.Resolver {
	return resolver{
		Resolver: base.Resolver{
			Fetch:     fetchFn(c),
			Unmarshal: unmarshal,
		},
	}
}

func fetchFn(c *secretsmanager.Client) base.Fetcher {
	return func(ctx context.Context, reference string) ([]byte, error) {
		req := &secretsmanager.GetSecretValueInput{
			SecretId: aws.String(validName(reference)),
		}
		resp, err := c.GetSecretValue(ctx, req)
		if err != nil {
			return nil, err
		}
		if resp.SecretString != nil {
			return []byte(*resp.SecretString), nil
		}
		return nil, nil
	}
}

const Scheme = "aws-secretsmanager"

func validName(s string) string {
	scheme, key, _ := strings.Cut(s, "://")
	if scheme != Scheme {
		return ""
	}
	return key
}
