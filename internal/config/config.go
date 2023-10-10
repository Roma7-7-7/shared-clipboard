package config

type Config struct {
	Server Server
	Web    Web
}

type Server struct {
	Port int
}

type Web struct {
	TemplatesPath string
}
