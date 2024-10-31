package slimfig_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/HayoVanLoon/go-slimfig"
	"github.com/HayoVanLoon/go-slimfig/resolver"
)

func Test_Load(t *testing.T) {
	type fields struct {
		configEnv string
		envs      map[string]string
		resolvers []resolver.Resolver
	}
	type want struct {
		value map[string]any
		err   require.ErrorAssertionFunc
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			"no config",
			fields{},
			want{
				map[string]any{},
				require.NoError,
			},
		},
		{
			"happy",
			fields{
				configEnv: "ref1,ref2",
				envs:      map[string]string{prefix_ + "foo": "123"},
				resolvers: []resolver.Resolver{
					TestResolver{
						matchOn: "ref1",
						data: map[string]any{
							"foo": "xxx",
							"a":   1,
							"b":   "x",
							"c":   map[string]any{"d": 4},
						},
					},
					TestResolver{
						matchOn: "ref2",
						data: map[string]any{
							"foo": "yyy",
							"a":   1,
							"b":   2,
							"c":   map[string]any{"e": 5},
						},
					},
				},
			},
			want{
				map[string]any{
					"foo": "123",
					"a":   1,
					"b":   2,
					"c":   map[string]any{"d": 4, "e": 5},
				},
				require.NoError,
			},
		},
		{
			"no resolver",
			fields{
				configEnv: "ref1,ref2",
				resolvers: []resolver.Resolver{
					TestResolver{
						matchOn: "ref1",
					},
				},
			},
			want{
				map[string]any{},
				func(t require.TestingT, err error, _ ...interface{}) {
					require.Error(t, err)
					require.Equal(t, "no resolver for \"ref2\"", err.Error())
				},
			},
		},
		{
			"error resolving",
			fields{
				configEnv: "ref1,ref2",
				resolvers: []resolver.Resolver{
					TestResolver{
						matchOn: "ref1",
					},
					TestResolver{
						matchOn: "ref2",
						err:     fmt.Errorf("oh noes"),
					},
				},
			},
			want{
				map[string]any{},
				func(t require.TestingT, err error, _ ...interface{}) {
					require.Error(t, err)
					require.Equal(t, "error resolving \"ref2\": oh noes", err.Error())
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, clean(func(t *testing.T) {
			setEnvs(tt.fields.envs)
			setEnv(prefix_+slimfig.EnvSuffix, tt.fields.configEnv)
			slimfig.SetResolvers(tt.fields.resolvers...)

			err := slimfig.Load(prefix)
			tt.want.err(t, err)
			actual := slimfig.Config()
			require.Equal(t, tt.want.value, actual)
		}))
	}
}

func Test_merge(t *testing.T) {
	type args struct {
		old map[string]any
		v   map[string]any
	}
	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{
			"add simple value",
			args{
				map[string]any{"a": 1, "c": map[string]any{"d": 3, "e": 4}},
				map[string]any{"b": 2},
			},
			map[string]any{"a": 1, "b": 2, "c": map[string]any{"d": 3, "e": 4}},
		},
		{
			"add too empty map",
			args{
				map[string]any{},
				map[string]any{"b": 2},
			},
			map[string]any{"b": 2},
		},
		{
			"replace simple value",
			args{
				map[string]any{"a": 1, "c": map[string]any{"d": 3, "e": 4}},
				map[string]any{"a": "x"},
			},
			map[string]any{"a": "x", "c": map[string]any{"d": 3, "e": 4}},
		},
		{
			"add map",
			args{
				map[string]any{"a": 1},
				map[string]any{"c": map[string]any{"e": 4}},
			},
			map[string]any{"a": 1, "c": map[string]any{"e": 4}},
		},
		{
			"add values at different levels",
			args{
				map[string]any{"c": map[string]any{"d": 3}},
				map[string]any{"b": 2, "c": map[string]any{"e": 4}},
			},
			map[string]any{"b": 2, "c": map[string]any{"d": 3, "e": 4}},
		},
		{
			"replace map with non-map",
			args{
				map[string]any{"b": 2, "c": map[string]any{"f": map[string]any{"g": "6"}}},
				map[string]any{"c": map[string]any{"f": "x"}},
			},
			map[string]any{"b": 2, "c": map[string]any{"f": "x"}},
		},
		{
			"replace non-map with map",
			args{
				map[string]any{"a": 1, "b": 2, "c": map[string]any{"d": 4}},
				map[string]any{"b": map[string]any{"f": "x"}},
			},
			map[string]any{
				"a": 1,
				"b": map[string]any{"f": "x"},
				"c": map[string]any{"d": 4},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, clean(func(t *testing.T) {
			actual := &tt.args.old
			slimfig.Merge(actual, tt.args.v)
			require.Equal(t, tt.want, *actual)
		}))
	}
}

func Test_loadEnvironment(t *testing.T) {
	type fields struct {
		config map[string]any
		envs   map[string]string
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]any
	}{
		{
			"happy",
			fields{
				map[string]any{
					"a": 1,
					"b": 2,
				},
				map[string]string{
					prefix_ + "foo": "123",
					prefix_ + "BAR": "456",
				},
			},
			map[string]any{
				"a":   1,
				"b":   2,
				"foo": "123",
				"BAR": "456",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, clean(func(t *testing.T) {
			setEnvs(tt.fields.envs)
			slimfig.SetConfig(tt.fields.config)

			slimfig.LoadEnvironment(prefix)
			actual := slimfig.Config()
			require.Equal(t, tt.want, actual)
		}))
	}
}

func Test_addEnv(t *testing.T) {
	type args struct {
		old map[string]any
		k   string
		v   string
	}
	tests := []struct {
		name string
		args args
		want map[string]any
	}{
		{
			"simple",
			args{
				map[string]any{"a": 1},
				"b",
				"2",
			},
			map[string]any{"a": 1, "b": "2"},
		},
		{
			"nested",
			args{
				map[string]any{"a": 1},
				"c__f__g",
				"6",
			},
			map[string]any{
				"a": 1,
				"c": map[string]any{
					"f": map[string]any{"g": "6"},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, clean(func(t *testing.T) {
			actual := &tt.args.old
			slimfig.AddEnv(actual, tt.args.k, tt.args.v)
			require.Equal(t, tt.want, *actual)
		}))
	}
}

func TestSimpleGetters(t *testing.T) {
	config := map[string]any{
		"foo":        "1",
		"foo_int":    1,
		"foo_float":  float64(1),
		"bar_string": "true",
		"bar_bool":   true,
		"bar_false":  "false",
		"bar_zero":   0,
		"bar_empty":  "",
		"bla": map[string]any{
			"moo": 2,
		},
	}

	tests := []struct {
		name string
		fns  []func() any
		want []any
	}{
		{
			"string",
			[]func() any{
				func() any { return slimfig.String("foo", "fallback") },
				func() any { return slimfig.String("foo_int", "fallback") },
				func() any { return slimfig.String("foo_float", "fallback") },
				func() any { return slimfig.String("bar_string", "fallback") },
				func() any { return slimfig.String("bar_bool", "fallback") },
				func() any { return slimfig.String("bla.moo", "fallback") },
				func() any { return slimfig.String("xxx", "fallback") },
			},
			[]any{"1", "1", "1", "true", "true", "2", "fallback"},
		},
		{
			"int",
			[]func() any{
				func() any { return slimfig.Int("foo", -1) },
				func() any { return slimfig.Int("foo_int", -1) },
				func() any { return slimfig.Int("foo_float", -1) },
				func() any { return slimfig.Int("bar_string", -1) },
				func() any { return slimfig.Int("bar_bool", -1) },
				func() any { return slimfig.Int("bla.moo", -1) },
				func() any { return slimfig.Int("xxx", -1) },
			},
			[]any{1, 1, 1, -1, -1, 2, -1},
		},
		{
			"float",
			[]func() any{
				func() any { return slimfig.Float("foo", -1) },
				func() any { return slimfig.Float("foo_int", -1) },
				func() any { return slimfig.Float("foo_float", -1) },
				func() any { return slimfig.Float("bar_string", -1) },
				func() any { return slimfig.Float("bar_bool", -1) },
				func() any { return slimfig.Float("bla.moo", -1) },
				func() any { return slimfig.Float("xxx", -1) },
			},
			[]any{
				float64(1),
				float64(1),
				float64(1),
				float64(-1),
				float64(-1),
				float64(2),
				float64(-1),
			},
		},
		{
			"bool",
			[]func() any{
				func() any { return slimfig.Bool("foo", false) },
				func() any { return slimfig.Bool("foo_int", false) },
				func() any { return slimfig.Bool("foo_float", false) },
				func() any { return slimfig.Bool("bar_string", false) },
				func() any { return slimfig.Bool("bar_bool", false) },
				func() any { return slimfig.Bool("bar_false", true) },
				func() any { return slimfig.Bool("bar_zero", true) },
				func() any { return slimfig.Bool("bar_empty", true) },
				func() any { return slimfig.Bool("xxx", false) },
			},
			[]any{true, true, true, true, true, false, false, true, false},
		},
		{
			"any",
			[]func() any{
				func() any { return slimfig.Any("foo", "fallback") },
				func() any { return slimfig.Any("foo_int", "fallback") },
				func() any { return slimfig.Any("foo_float", "fallback") },
				func() any { return slimfig.Any("bar_string", "fallback") },
				func() any { return slimfig.Any("bar_bool", "fallback") },
				func() any { return slimfig.Any("bla.moo", "fallback") },
				func() any { return slimfig.Any("xxx", "fallback") },
			},
			[]any{"1", 1, float64(1), "true", true, 2, "fallback"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, clean(func(t *testing.T) {
			slimfig.SetConfig(config)
			var actual []any
			for i := range tt.fns {
				actual = append(actual, tt.fns[i]())
			}
			require.Equal(t, tt.want, actual)
		}))
	}
}

func TestStringMap(t *testing.T) {
	config := map[string]any{
		"string-string": map[string]string{"a": "1", "b": "2"},
		"string-any":    map[string]any{"a": 1, "b": float32(2)},
		"not-map":       -1,
		"empty":         map[string]any{},
	}
	fallback := map[string]string{"fallback": "fallback"}

	tests := []struct {
		name string
		want map[string]string
	}{
		{
			"string-string",
			map[string]string{"a": "1", "b": "2"},
		},
		{
			"string-any",
			map[string]string{"a": "1", "b": "2"},
		},
		{"not-map", fallback},
		{"fallback", fallback},
		{"empty", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, clean(func(t *testing.T) {
			slimfig.SetConfig(config)
			actual := slimfig.StringMap(tt.name, fallback)
			require.Equal(t, tt.want, actual)
		}))
	}
}

func TestIntMap(t *testing.T) {
	config := map[string]any{
		"string-int": map[string]int{"a": 1, "b": 2},
		"string-any": map[string]any{"a": 1, "b": float32(2)},
		"int-int":    map[int]int{10: 1, 20: 2},
		"not-map":    -1,
		"empty":      map[string]any{},
	}
	fallback := map[string]int{"fallback": 1}

	tests := []struct {
		name string
		want map[string]int
	}{
		{
			"string-int",
			map[string]int{"a": 1, "b": 2},
		},
		{
			"string-any",
			map[string]int{"a": 1, "b": 2},
		},
		{
			"int-int",
			map[string]int{"10": 1, "20": 2},
		},
		{"not-map", fallback},
		{"fallback", fallback},
		{"empty", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, clean(func(t *testing.T) {
			slimfig.SetConfig(config)
			actual := slimfig.IntMap(tt.name, fallback)
			require.Equal(t, tt.want, actual)
		}))
	}
}

func TestFloatMap(t *testing.T) {
	config := map[string]any{
		"string-float": map[string]float64{"a": 1, "b": 2},
		"string-any":   map[string]any{"a": 1, "b": float32(2)},
		"float-float":  map[float64]float64{10: 1, 20: 2},
		"not-map":      -1,
		"empty":        map[string]any{},
	}
	fallback := map[string]float64{"fallback": 1}

	tests := []struct {
		name string
		want map[string]float64
	}{
		{
			"string-float",
			map[string]float64{"a": 1, "b": 2},
		},
		{
			"string-any",
			map[string]float64{"a": 1, "b": 2},
		},
		{
			"float-float",
			map[string]float64{"10": 1, "20": 2},
		},
		{"not-map", fallback},
		{"fallback", fallback},
		{"empty", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, clean(func(t *testing.T) {
			slimfig.SetConfig(config)
			actual := slimfig.FloatMap(tt.name, fallback)
			require.Equal(t, tt.want, actual)
		}))
	}
}

func TestBoolMap(t *testing.T) {
	config := map[string]any{
		"string-bool": map[string]bool{"a": true, "b": false},
		"string-any":  map[string]any{"a": false, "b": float32(1)},
		"bool-bool":   map[bool]bool{true: false, false: true},
		"not-map":     -1,
		"empty":       map[string]any{},
	}
	fallback := map[string]bool{"fallback": true}

	tests := []struct {
		name string
		want map[string]bool
	}{
		{
			"string-bool",
			map[string]bool{"a": true, "b": false},
		},
		{
			"string-any",
			map[string]bool{"a": false, "b": true},
		},
		{
			"bool-bool",
			map[string]bool{"true": false, "false": true},
		},
		{"not-map", fallback},
		{"fallback", fallback},
		{"empty", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, clean(func(t *testing.T) {
			slimfig.SetConfig(config)
			actual := slimfig.BoolMap(tt.name, fallback)
			require.Equal(t, tt.want, actual)
		}))
	}
}

func clean(f func(t *testing.T)) func(t *testing.T) {
	return func(t *testing.T) {
		cleanUp()
		defer cleanUp()
		defer func() {
			if r := recover(); r != nil {
				cleanUp()
				panic(r)
			}
		}()
		f(t)
	}
}

const (
	prefix  = "TESTSLIMFIG"
	prefix_ = prefix + "_"
)

func cleanUp() {
	slimfig.Reset()
	slimfig.SetResolvers()
	for _, s := range os.Environ() {
		k, _, ok := strings.Cut(s, "=")
		if !ok {
			continue
		}
		if !strings.HasPrefix(k, prefix_) {
			continue
		}
		if err := os.Unsetenv(k); err != nil {
			panic(err)
		}
	}
}

func setEnvs(envs map[string]string) {
	for k, v := range envs {
		setEnv(k, v)
	}
}

func setEnv(k, v string) {
	if err := os.Setenv(k, v); err != nil {
		panic(err)
	}
}

type TestResolver struct {
	matchOn string
	data    map[string]any
	err     error
}

func (t TestResolver) Matches(reference string) bool {
	return reference == t.matchOn
}

func (t TestResolver) Resolve(string) (map[string]any, error) {
	if t.err != nil {
		return nil, t.err
	}
	return t.data, nil
}
