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

func DefaultConfig() *Config {
	return &Config{
		GrpcPort: 1000,
		HttpPort: 1001,
	}
}

func NewConfig(grpcPort, httpPort int) *Config {
	return &Config{
		GrpcPort: grpcPort,
		HttpPort: httpPort,
	}
}
