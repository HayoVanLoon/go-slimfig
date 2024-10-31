package slimfig

import (
	"fmt"
	"os"
	"strings"

	"github.com/HayoVanLoon/go-slimfig/resolver"
	"github.com/HayoVanLoon/go-slimfig/resolver/json"
)

var resolvers = []resolver.Resolver{json.Resolver}

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
		return nil, false
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

// Load loads the configuration scheme from the environment variable
// 'XX_CONFIG', where 'XX' is the given prefix. Initialisation follows an
// all-or-nothing principle, so following an error, the configuration will
// remain uninitialised.
//
// The configuration scheme is a list of references to configuration maps,
// where each subsequent map modifies the configuration by adding or
// overwriting values.
//
// After that, it will look for other environment variables with the given
// prefix. Their names are than translated into JSON-path(-like) keys by:
//   - removing the prefix
//   - translating double underscores into dots
//
// For instance, given prefix "XX", "XX_service__host_name" becomes "service.host_name".
//
// This method should only be called once. Subsequent calls will always reset
// the configuration.
//
// When using custom resolvers, these must be set via SetResolvers prior to
// calling this method.
func Load(prefix string) error {
	s := os.Getenv(prefix + "_" + EnvSuffix)
	if s != "" {
		if err := LoadScheme(strings.Split(s, ",")); err != nil {
			return err
		}
	}
	loadEnvironment(prefix)
	return nil
}

// LoadScheme loads the provided configuration scheme. Unlike Load, it does not
// check inspect environment variables. Initialisation follows an
// all-or-nothing principle, so following an error, the configuration will
// remain uninitialised.
//
// The configuration scheme is a list of references to configuration maps,
// where each subsequent map modifies the configuration by adding or
// overwriting values.
//
// It is advised to call this method only once. Subsequent calls will first
// reset the configuration. When using custom resolvers, these must be set via
// SetResolvers prior to calling this method.
func LoadScheme(references []string) error {
	reset()
	rs := make([]resolver.Resolver, len(references))
	for _, ref := range references {
		ref = strings.TrimSpace(ref)
		found := false
		for i, r := range resolvers {
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
		cfg, err := rs[i].Resolve(references[i])
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
			(*old)[k] = v
			continue
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
