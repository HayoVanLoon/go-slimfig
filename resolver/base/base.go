package base

import (
	"context"
	"fmt"
)

type (
	Fetcher      func(ctx context.Context, reference string) ([]byte, error)
	Unmarshaller func([]byte, any) error
)

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
