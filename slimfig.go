package slimfig

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/HayoVanLoon/go-slimfig/resolver"
	"github.com/HayoVanLoon/go-slimfig/resolver/json"
)

var resolvers = []resolver.Resolver{json.Resolver}

// SetResolvers sets the resolvers for configuration map references. Order
// matters as a reference will be resolved by the first matching resolver.
//
// The default set consists of only the JSON file resolver.
func SetResolvers(rs ...resolver.Resolver) {
	resolvers = rs
}

const EnvSuffix = "CONFIG"

type configMap map[string]any

func (m configMap) get(key []string) (any, bool) {
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
	return configMap(m2).get(key[1:])
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
// 'XX_CONFIG', where 'XX' is the prefix.
//
// It will first load the configuration map(s) provided by the XX_CONFIG
// variable. Each successive map is applied as a patch on the existing
// configuration.
//
// After that, it will continue with adding other environment variables that
// have this prefix. The key (after removing the prefix) is then used as a
// JSON-path, where double underscores are translated into dots. For instance
// "XX_service__host_name" becomes "service.host_name".
//
// It is advised to call this method only once. Subsequent calls will first
// reset the configuration. When using custom resolvers, these must be set via
// SetResolvers prior to calling this method.
func Load(prefix string) error {
	s := os.Getenv(prefix + "_CONFIG")
	if s != "" {
		if err := loadFrom(strings.Split(s, ",")); err != nil {
			return err
		}
	}
	loadEnvironment(prefix)
	return nil
}

func loadFrom(refs []string) error {
	reset()
	rs := make([]resolver.Resolver, len(refs))
	for _, ref := range refs {
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

	for i := range rs {
		cfg, err := rs[i].Resolve(refs[i])
		if err != nil {
			return fmt.Errorf("error resolving %q: %w", refs[i], err)
		}
		merge(&config, cfg)
	}
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

func String(key, fallback string) string {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	return toString(a)
}

func Int(key string, fallback int) int {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	i, ok := toInt(a)
	if !ok {
		return fallback
	}
	return i
}

func Float(key string, fallback float64) float64 {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	f, ok := toFloat64(a)
	if !ok {
		return fallback
	}
	return f
}

func Bool(key string, fallback bool) bool {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	b, ok := toBool(a)
	if !ok {
		return fallback
	}
	return b
}

func Any(key string, fallback any) any {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	return a
}

func StringSlice(key string, fallback []string) []string {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	switch x := a.(type) {
	case []string:
		return x
	case []any:
		var out []string
		for i := range x {
			out = append(out, toString(x[i]))
		}
		return out
	default:
		return fallback
	}
}

func IntSlice(key string, fallback []int) []int {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	switch x := a.(type) {
	case []int:
		return x
	case []any:
		var out []int
		for i := range x {
			if j, ok := toInt(x[i]); ok {
				out = append(out, j)
			}
		}
		return out
	default:
		return fallback
	}
}

func FloatSlice(key string, fallback []float64) []float64 {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	switch x := a.(type) {
	case []float64:
		return x
	case []any:
		var out []float64
		for i := range x {
			if j, ok := toFloat64(x[i]); ok {
				out = append(out, j)
			}
		}
		return out
	default:
		return fallback
	}
}

func BoolSlice(key string, fallback []bool) []bool {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	switch x := a.(type) {
	case []bool:
		return x
	case []any:
		var out []bool
		for i := range x {
			if j, ok := toBool(x[i]); ok {
				out = append(out, j)
			}
		}
		return out
	default:
		return fallback
	}
}

func StringMap(key string, fallback map[string]string) map[string]string {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	switch x := a.(type) {
	case map[string]string:
		return x
	case map[string]any:
		m := make(map[string]string)
		for k, v := range x {
			m[k] = toString(v)
		}
		return m
	default:
		return fallback
	}
}

func IntMap(key string, fallback map[string]int) map[string]int {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	switch x := a.(type) {
	case map[string]int:
		return x
	case map[string]any:
		m := make(map[string]int)
		for k, v := range x {
			if i, ok := toInt(v); ok {
				m[k] = i
			}
		}
		return m
	default:
		return fallback
	}
}

func FloatMap(key string, fallback map[string]float64) map[string]float64 {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	switch x := a.(type) {
	case map[string]float64:
		return x
	case map[string]float32:
		m := make(map[string]float64)
		for k, v := range x {
			m[k] = float64(v)
		}
		return m
	case map[string]any:
		m := make(map[string]float64)
		for k, v := range x {
			if i, ok := toFloat64(v); ok {
				m[k] = i
			}
		}
		return m
	default:
		return fallback
	}
}

func BoolMap(key string, fallback map[string]bool) map[string]bool {
	parts := strings.Split(key, ".")
	a, ok := config.get(parts)
	if !ok {
		return fallback
	}
	switch x := a.(type) {
	case map[string]bool:
		return x
	case map[string]any:
		m := make(map[string]bool)
		for k, v := range x {
			if i, ok := toBool(v); ok {
				m[k] = i
			}
		}
		return m
	default:
		return fallback
	}
}

func toString(a any) string {
	if s, ok := a.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", a)
}

func toInt(a any) (int, bool) {
	// uint and similar might cause overflows
	switch x := a.(type) {
	case int:
		return x, true
	case int8:
		return int(x), true
	case int16:
		return int(x), true
	case int32:
		return int(x), true
	case int64:
		return int(x), true
	case float32:
		return int(x), true
	case float64:
		return int(x), true
	}
	i, err := strconv.Atoi(toString(a))
	if err != nil {
		return 0, false
	}
	return i, true
}

func toFloat64(a any) (float64, bool) {
	switch x := a.(type) {
	case float32:
		return float64(x), true
	case float64:
		return x, true
	}
	f, err := strconv.ParseFloat(toString(a), 64)
	if err != nil {
		return 0, false
	}
	return f, true
}

func toBool(a any) (bool, bool) {
	if b, ok := a.(bool); ok {
		return b, true
	}
	b, err := strconv.ParseBool(fmt.Sprintf("%v", a))
	if err != nil {
		return false, false
	}
	return b, true
}
