package internal

// AppConfig is the global configuration for the application
var AppConfig = &Config{}
var ClusterName = "iuf"

// Config is the configuration for the application
type Config struct {
	Force   bool
	Tarball string
}
