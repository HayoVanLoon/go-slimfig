package resolver

type Resolver interface {
	// Matches returns true when the resolver is able to handle this reference.
	Matches(reference string) bool
	// Resolve resolves the reference to a map.
	Resolve(reference string) (map[string]any, error)
}
