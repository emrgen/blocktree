package blocktree

type DbConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type Config struct {
	GrpcPort int
	HttpPort int
}

func NewConfig() *Config {
	return &Config{}
}
