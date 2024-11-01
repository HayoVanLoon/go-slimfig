// Package json provides a Slimfig resolver for JSON files.
package json

import (
	"context"
	"encoding/json"
	"os"
	"strings"

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
	return m, json.NewDecoder(f).Decode(&m)
}

// Resolver returns a resolver for JSON files. By default, it only matches
// references ending in ".json".
func Resolver(extensions ...string) res.Resolver {
	if len(extensions) == 0 {
		extensions = []string{".json"}
	}
	return resolver{
		extensions: extensions,
	}
}
