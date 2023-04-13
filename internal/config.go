package internal

var AppConfig = &Config{}

type Config struct {
	Force   bool
	Tarball string
}
