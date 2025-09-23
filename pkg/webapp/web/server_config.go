package web

type ServerConfig struct {
	HttpPort          string
	GrpcPort          string
	Environment       string
	UseDefaultRecover bool
	UseDefaultLogger  bool
}
