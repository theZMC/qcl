package qcl

type Loader func(any) error       // Loader is a function that loads the configuration from a specific source.
type LoadOption func(*LoadConfig) // LoadOption is a function that configures the Load function's LoadConfig. The Load function accepts a variable number of LoadOptions.
type Source string                // Source is the type of the configuration source. The following sources are supported out of the box: Environment, File, and Flag.

type LoadConfig struct {
	Sources []Source
	Loaders map[Source]Loader
}

var defaultOptions = []LoadOption{
	UseEnv(),
	UseFlags(),
	InThisOrder(Environment, Flag),
}

// Load loads the configuration from the specified sources. The configuration is loaded in the same order as the
// sources are specified with later sources overriding earlier ones.
//
// Example:
//
//	ocl.Load(&defaultConfig, ocl.UseConfigFile("config.yaml", ocl.YAML), ocl.UseEnv())
//
// will load the configuration from the config file first, and then from the environment variables. If the same
// configuration field is set in both the config file and the environment variables, the value from the environment
// variables will be used. If the config file is not found, the configuration will be loaded from the environment
// variables. If the environment variables are not set for a field, the value specified in the defaultConfig struct
// will be used.
func Load[T any](defaultConfig *T, opts ...LoadOption) (*T, error) {
	config := new(LoadConfig)
	config.Sources = make([]Source, 0, len(opts))
	config.Loaders = make(map[Source]Loader, len(opts))

	if len(opts) == 0 {
		opts = defaultOptions
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

// InThisOrder allows you to specify the order in which the configuration sources will be loaded. By default, the order
// is determined by the order in which the LoadOptions are passed to the Load function. This function allows you to
// override that order after the fact.
//
// Example:
//
//	ocl.Load(
//		&defaultConfig,
//		ocl.UseConfigFile("config.yaml", ocl.YAML),
//		ocl.UseEnv(),
//		ocl.InThisOrder(ocl.Environment, ocl.File),
//	)
//
// will load the configuration from the environment variables first, and then from the config file.
func InThisOrder(sources ...Source) LoadOption {
	return func(o *LoadConfig) {
		o.Sources = sources
	}
}
