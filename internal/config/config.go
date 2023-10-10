package config

type Config struct {
	Server Server
	DB     DB
}

type Server struct {
	Port int
}

type DB struct {
	Path string
}
