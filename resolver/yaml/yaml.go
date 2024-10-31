// Package yaml provides a Slimfig resolver for YAML files.
package yaml

import (
	"context"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	res "github.com/HayoVanLoon/go-slimfig/resolver"
)

const ProtocolFile = "file://"

var _ res.Resolver = *new(resolver)

type resolver struct {
	extensions []string
}

func (r resolver) Matches(reference string) bool {
	for _, ext := range r.extensions {
		if strings.HasSuffix(reference, ext) {
			return true
		}
	}
	return false
}

func (r resolver) Resolve(_ context.Context, reference string) (map[string]any, error) {
	reference = strings.TrimPrefix(reference, ProtocolFile)
	f, err := os.Open(reference) //nolint:gosec
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	m := make(map[string]any)
	return m, yaml.NewDecoder(f).Decode(&m)
}

// Resolver returns a resolver for YAML files. By default, it only matches
// references ending in ".yaml".
func Resolver(extensions ...string) res.Resolver {
	if len(extensions) == 0 {
		extensions = []string{".yaml"}
	}
	return resolver{
		extensions: extensions,
	}
}
