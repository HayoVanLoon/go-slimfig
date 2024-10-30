// Package json provides a Slimfig resolver for JSON files.
package json

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/HayoVanLoon/go-slimfig/resolver"
	"github.com/HayoVanLoon/go-slimfig/shared"
)

var _ resolver.Resolver = *new(Resolver)

type Resolver struct{}

func (r Resolver) Matches(reference string) bool {
	// do not filter by file extension
	return shared.MaybeFile(reference)
}

func (r Resolver) Resolve(reference string) (map[string]any, error) {
	reference = strings.TrimPrefix(reference, shared.ProtocolFile)
	f, err := os.Open(reference)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	m := make(map[string]any)
	return m, json.NewDecoder(f).Decode(&m)
}
