package slimfig

func Merge(old *map[string]any, v map[string]any) {
	merge((*configMap)(old), v)
}

func LoadEnvironment(prefix string) {
	loadEnvironment(prefix)
}

func AddEnv(old *map[string]any, k string, v string) {
	addEnv((*configMap)(old), k, v)
}

func Config() map[string]any {
	return config
}

func SetConfig(m map[string]any) {
	config = m
}

func Reset() {
	reset()
}
