package base

import (
	"context"
	"fmt"
)

type (
	// A Fetcher fetches the data indicated by the reference.
	Fetcher func(ctx context.Context, reference string) ([]byte, error)
	// An Unmarshaller unmarshalls bytes into a map[string]any. Its second
	// parameter type is any only so common Unmarshallers (like json.Unmarshal)
	// can be used without wrapping.
	Unmarshaller func([]byte, any) error
)

// A Resolver partially implements a Resolver. Matching will have to be
// implemented separately.
type Resolver struct {
	Fetch     Fetcher
	Unmarshal Unmarshaller
}

func (r Resolver) Resolve(ctx context.Context, reference string) (map[string]any, error) {
	data, err := r.Fetch(ctx, reference)
	if err != nil {
		return nil, fmt.Errorf("error fetching secret: %w", err)
	}
	return r.parse(data)
}

func (r Resolver) parse(data []byte) (map[string]any, error) {
	m := make(map[string]any)
	if err := r.Unmarshal(data, &m); err != nil {
		return nil, err
	}
	return m, nil
}
