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

	MongoDB MongoDB `yaml: "mongodb"`
}

//Logging configuration for the logging system (At the moment only for the gelf logger)
type Logging struct {
	Gelfurl  string `yaml:"gelf-url"`
	Gelfport int    `yaml:"gelf-port"`
}

//HealthCheck configuration for the healthcheck system
type HealthCheck struct {
	Period int `yaml:"period"`
}

//MongoDB configuration for the mongodb stoirage
type MongoDB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	AuthDB   string `yaml:"authdb"`
	Database string `yaml:"database"`
}
