package config

// Config our service configuration
type Config struct {
	//port of the http server
	Port int `yaml:"port"`
	//port of the https server
	Sslport int `yaml:"sslport"`
	//this is the url how to connect to this service from outside
	ServiceURL string `yaml:"serviceURL"`
	//this is the url where to register this service
	RegistryURL string `yaml:"registryURL"`
	//this is the url where to register this service
	SystemID    string `yaml:"systemID"`
	BackendPath string `yaml:"backendpath"`

	SecretFile string `yaml:"secretfil`

	Logging Logging `yaml:"logging"`

	HealthCheck HealthCheck `yaml:"healthcheck"`
}

type Logging struct {
	Gelfurl  string `yaml:"gelf-url"`
	Gelfport int    `yaml:"gelf-port"`
}

type HealthCheck struct {
	Period int `yaml:"period"`
}
