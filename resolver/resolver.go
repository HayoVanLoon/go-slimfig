package resolver

import "context"

type Resolver interface {
	// Matches returns true when the resolver is able to handle this reference.
	Matches(reference string) bool
	// Resolve resolves the reference to a map.
	Resolve(ctx context.Context, reference string) (map[string]any, error)
}
