// Package memory provides a resolver wrapped around an in-memory map.
package memory

import (
	"context"

	res "github.com/HayoVanLoon/go-slimfig/resolver"
)

var _ res.Resolver = *new(resolver)

type resolver struct {
	reference string
	data      map[string]any
}

func (r resolver) Matches(reference string) bool {
	return reference == r.reference
}

func (r resolver) Resolve(context.Context, string) (map[string]any, error) {
	return r.data, nil
}

func Resolver(reference string, data map[string]any) res.Resolver {
	return resolver{
		reference: reference,
		data:      data,
	}
}
