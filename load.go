package qcl

type Loader func(any) error       // Loader is a function that loads the configuration from a specific source.
type LoadOption func(*LoadConfig) // LoadOption is a function that configures the Load function's LoadConfig. The Load function accepts a variable number of LoadOptions.
type Source string                // Source is the type of the configuration source. The following sources are supported out of the box: Environment and Flag.

type LoadConfig struct {
	Sources []Source
	Loaders map[Source]Loader
}

var DefaultLoadOptions = []LoadOption{
	UseEnv(),
	UseFlags(),
}

// Load modifies the pointer it receives with configuration information from the sources specified in the LoadOptions.
// The Load function are passed to the Load function. The default LoadOptions are:
//
//	 DefaultLoadOptions := []LoadOption{
//		  qcl.UseEnv(qcl.WithPrefix(""), qcl.WithEnvSeparator(","), qcl.WithEnvStructTag("env")),
//		  qcl.UseFlags()
//	 }
//
// If no LoadOptions are passed to the Load function, the default LoadOptions will be used.
//
// Example:
//
//	qcl.Load(&defaultConfig)
//
// is equivalent to:
//
//	qcl.Load(&defaultConfig, qcl.DefaultLoadOptions...)
//
// If any LoadOption is passed to the Load function, the default LoadOptions will not be used.
// The Load function returns a pointer to the configuration struct, and an error.
func Load[T any](defaultConfig *T, opts ...LoadOption) (*T, error) {
	config := new(LoadConfig)
	config.Sources = make([]Source, 0, len(opts))
	config.Loaders = make(map[Source]Loader, len(opts))

	if len(opts) == 0 {
		return Load(defaultConfig, DefaultLoadOptions...)
	}

	for _, opt := range opts {
		opt(config)
	}

	if defaultConfig == nil {
		defaultConfig = new(T)
	}
	for _, source := range config.Sources {
		if load, ok := config.Loaders[source]; ok {
			err := load(defaultConfig)
			if err != nil {
				return nil, err
			}
		}
	}

	return defaultConfig, nil
}
