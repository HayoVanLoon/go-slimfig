package slimfig

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/HayoVanLoon/go-slimfig/resolver"
	jsonresolver "github.com/HayoVanLoon/go-slimfig/resolver/json"
)

var resolvers = []resolver.Resolver{jsonresolver.Resolver}

// SetResolvers sets the resolvers for configuration map references. Order
// matters as a reference will be resolved by the first matching resolver.
//
// The default set consists of only the JSON file resolver. When using custom
// resolvers, these must be set using this function before calling Load or
// LoadScheme.
func SetResolvers(rs ...resolver.Resolver) {
	resolvers = rs
}

// EnvSuffix suffix added to the prefix to build the configuration scheme
// environment variable, i.e.:
//
//	scheme := os.Getenv(prefix + "_" + EnvSuffix)
const EnvSuffix = "CONFIG"

// Load loads the configuration scheme from either an environment variable
// starting with the prefix or the given references.
//
// The configuration scheme is a list of references to configuration maps,
// which are resolved during initialisation. The first map forms the base and
// each subsequent one modifies this base by adding or overwriting values.
//
// When both prefix and references are specified, the environment variable
// will take precedence, provided its value is non-empty. The configuration
// scheme environment variable is constructed by appending "_CONFIG" to the
// prefix; for instance "XX_CONFIG" for prefix "XX".
//
// If a prefix has been specified, other environment variable starting with it
// will be pulled into the configuration. First their names are translated into
// (sort of) JSON-path locations and then the configuration is updated at those
// locations. The translation follows these two steps:
//   - the prefix is removed
//   - double underscores are translated into dots
//
// For instance, given prefix "XX", "XX_service__Host_Name" becomes
// "service.Host_Name". Notice that the casing is being preserved.
//
// Initialisation is all-or-nothing, so in case of any error, the configuration
// will remain uninitialised.
//
// This method should only be called once. Subsequent calls will always reset
// the configuration.
//
// When using custom resolvers, these must be set via SetResolvers prior to
// calling this method.
func Load(ctx context.Context, prefix string, references ...string) error {
	if prefix != "" {
		if s := os.Getenv(prefix + "_" + EnvSuffix); s != "" {
			references = strings.Split(s, ",")
		}
	}
	if err := loadScheme(ctx, references); err != nil {
		return err
	}
	if prefix != "" {
		loadEnvironment(prefix)
	}
	return nil
}

type configMap map[string]any

func (m configMap) get(key string) (any, bool) {
	return m.get2(strings.Split(key, "."))
}

func (m configMap) get2(key []string) (any, bool) {
	if len(key) == 0 {
		return nil, false
	}
	v, ok := m[key[0]]
	if !ok {
		return nil, false
	}
	if len(key) == 1 {
		return v, true
	}
	m2, ok := v.(map[string]any)
	if !ok {
		if m2, ok = toMap(v, toAny); !ok {
			return nil, false
		}
	}
	return configMap(m2).get2(key[1:])
}

func (m configMap) getPointer(key string) (*any, bool) {
	v, ok := m[key]
	if !ok {
		return nil, false
	}
	return &v, true
}

var config = configMap{}

func reset() {
	config = configMap{}
}

func loadScheme(ctx context.Context, references []string) error {
	reset()
	rs := make([]resolver.Resolver, len(references))
	for i, ref := range references {
		ref = strings.TrimSpace(ref)
		found := false
		for _, r := range resolvers {
			if found = r.Matches(ref); found {
				rs[i] = r
				break
			}
		}
		if !found {
			return fmt.Errorf("no resolver for %q", ref)
		}
	}

	out := configMap{}
	for i := range rs {
		cfg, err := rs[i].Resolve(ctx, references[i])
		if err != nil {
			return fmt.Errorf("error resolving %q: %w", references[i], err)
		}
		merge(&out, cfg)
	}
	config = out
	return nil
}

func merge(old *configMap, m map[string]any) {
	for k, v := range m {
		ovp, ok := (*old).getPointer(k)
		if !ok {
			(*old)[k] = v
			continue
		}
		vm, ok := v.(map[string]any)
		if !ok {
			(*old)[k] = v
			continue
		}
		ovm, ok := (*ovp).(map[string]any)
		if !ok {
			// maps should be string-any, if it is a map: convert it
			ovm, ok = toMap(*ovp, toAny)
			if !ok {
				(*old)[k] = v
				continue
			}
			(*old)[k] = ovm
		}
		merge((*configMap)(&ovm), vm)
	}
}

func loadEnvironment(prefix string) {
	prefix += "_"
	for _, kv := range os.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if !ok {
			continue
		}
		if _, k, ok = strings.Cut(k, prefix); ok && k != EnvSuffix {
			addEnv(&config, k, v)
		}
	}
}

func addEnv(old *configMap, k string, v string) {
	parts := strings.Split(k, "__")
	if len(parts) == 1 {
		(*old)[k] = v
	}
	m := make(map[string]any)
	p := &m
	for i := 0; i < len(parts)-1; i += 1 {
		p2 := &map[string]any{}
		(*p)[parts[i]] = *p2
		p = p2
	}
	(*p)[parts[len(parts)-1]] = v
	merge(old, m)
}

// String looks up a configuration value as a string. If the stored value is
// not a string, it will use the value's standard string representation ("%v").
//
// Returns the fallback when the lookup fails.
func String(key, fallback string) string {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	return toString(a)
}

// Int looks up a configuration value as an integer. If the stored value is not
// an integer, some limited attempts will be made to convert or parse it into
// one.
//
// Returns the fallback when the lookup fails or the value cannot be converted.
func Int(key string, fallback int) int {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	i, ok := toInt(a)
	if !ok {
		return fallback
	}
	return i
}

// Float looks up a configuration value as a floating point. If the stored
// value is not a floating point, some limited attempts will be made to convert
// or parse it into one.
//
// Returns the fallback when the lookup fails or the value cannot be converted.
func Float(key string, fallback float64) float64 {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	f, ok := toFloat64(a)
	if !ok {
		return fallback
	}
	return f
}

// Bool looks up a configuration value as a boolean. If the stored value is not
// a boolean, an attempt will be made to parse it into one using the rules
// declared by strconv.ParseBool.
//
// Returns the fallback when the lookup fails or the value cannot be converted.
func Bool(key string, fallback bool) bool {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	b, ok := toBool(a)
	if !ok {
		return fallback
	}
	return b
}

// Any looks up a configuration value. Returns the fallback when the lookup
// fails.
func Any(key string, fallback any) any {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	return a
}

// StringSlice looks up a configuration value as a slice of strings. If the
// stored value is a slice, but not one of strings, it will convert non-string
// values using their standard string representation ("%v").
//
// Returns the fallback when the lookup fails.
func StringSlice(key string, fallback []string) []string {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	out, ok := toSlice(a, toString2)
	if !ok {
		return fallback
	}
	return out
}

// IntSlice looks up a configuration value as a slice of integers. If the
// stored value is a slice, but not one of integers, it will attempt to convert
// or parse the values. If this fails for any item, the fallback is returned.
// Also returns the fallback when the lookup fails.
func IntSlice(key string, fallback []int) []int {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	out, ok := toSlice(a, toInt)
	if !ok {
		return fallback
	}
	return out
}

// FloatSlice looks up a configuration value as a slice of floating point
// numbers. If the stored value is a slice, but not one of floating points, it
// will attempt to convert or parse the values. If this fails for any item, the
// fallback is returned. Also returns the fallback when the lookup fails.
func FloatSlice(key string, fallback []float64) []float64 {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	out, ok := toSlice(a, toFloat64)
	if !ok {
		return fallback
	}
	return out
}

// BoolSlice looks up a configuration value as a slice of booleans. If the
// stored value is a slice, but not one of booleans, it will attempt to convert
// or parse the values. If this fails for any item, the fallback is returned.
// Also returns the fallback when the lookup fails.
func BoolSlice(key string, fallback []bool) []bool {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	out, ok := toSlice(a, toBool)
	if !ok {
		return fallback
	}
	return out
}

// StringMap looks up a configuration value as a map of strings to strings. If
// the stored value is a map, but with different types, it will convert
// non-string keys and values using their standard string representation
// ("%v"). Returns the fallback when the lookup fails.
func StringMap(key string, fallback map[string]string) map[string]string {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	out, ok := toMap(a, toString2)
	if !ok {
		return fallback
	}
	return out
}

// IntMap looks up a configuration value as a map of strings to integers. If
// the stored value is a map, but not one of string to integers, it will
// attempt to convert or parse the values. If this fails for any entry, the
// fallback is returned. Also returns the fallback when the lookup fails.
func IntMap(key string, fallback map[string]int) map[string]int {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	out, ok := toMap(a, toInt)
	if !ok {
		return fallback
	}
	return out
}

// FloatMap looks up a configuration value as a map of strings to floating
// point numbers. If the stored value is a map, but not one of string to
// floating points, it will attempt to convert or parse the values. If this
// fails for any entry, the fallback is returned. Also returns the fallback
// when the lookup fails.
func FloatMap(key string, fallback map[string]float64) map[string]float64 {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	out, ok := toMap(a, toFloat64)
	if !ok {
		return fallback
	}
	return out
}

// BoolMap looks up a configuration value as a map of strings to booleans. If
// the stored value is a map, but not one of string to booleans, it will
// attempt to convert or parse the values. If this fails for any entry, the
// fallback is returned. Also returns the fallback when the lookup fails.
func BoolMap(key string, fallback map[string]bool) map[string]bool {
	a, ok := config.get(key)
	if !ok {
		return fallback
	}
	out, ok := toMap(a, toBool)
	if !ok {
		return fallback
	}
	return out
}

// JSON returns the current configuration as a JSON. Returns an error when the
// configuration is not JSON-serialisable.
func JSON() (string, error) {
	b := new(bytes.Buffer)
	enc := json.NewEncoder(b)
	enc.SetIndent("", "  ")
	if err := enc.Encode(config); err != nil {
		return "", fmt.Errorf("cannot serialise configuration: %w", err)
	}
	return b.String(), nil
}
