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

type resolver struct{}

func (r resolver) Matches(string) bool {
	// accept anything
	return true
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

var Resolver res.Resolver = resolver{}
