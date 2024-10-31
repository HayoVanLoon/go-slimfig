package base

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	pb "cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"

	res "github.com/HayoVanLoon/go-slimfig/resolver"
	"github.com/HayoVanLoon/go-slimfig/resolver/base"
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

// Resolver returns a Secret Manager resolver for secrets that can be
// unmarshalled into maps using the provided function.
func Resolver(ctx context.Context, unmarshal base.Unmarshaller) (res.Resolver, error) {
	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	return WithClient(c, unmarshal), nil
}

// WithClient returns a Secret Manager resolver with the given client and
// unmarshaller.
func WithClient(c *secretmanager.Client, unmarshal base.Unmarshaller) res.Resolver {
	return resolver{
		Resolver: base.Resolver{
			Fetch:     fetchFn(c),
			Unmarshal: unmarshal,
		},
	}
}

func fetchFn(c *secretmanager.Client) base.Fetcher {
	return func(ctx context.Context, reference string) ([]byte, error) {
		req := &pb.AccessSecretVersionRequest{
			Name: validName(reference),
		}
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		resp, err := c.AccessSecretVersion(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("error fetching secret: %w", err)
		}
		return resp.Payload.Data, nil
	}
}

func validName(s string) string {
	xs := strings.Split(s, "/")
	switch len(xs) {
	case 4:
		if xs[0] == "projects" && xs[2] == "secrets" {
			return s + "/versions/latest"
		}
	case 6:
		if xs[0] == "projects" && xs[2] == "secrets" {
			return s
		}
	case 8:
		if xs[0] == "projects" && xs[2] == "locations" && xs[4] == "secrets" {
			return s
		}
	}
	return ""
}
