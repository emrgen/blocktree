package blocktree

type Config struct {
	grpcPort int
	httpPort int
}

func NewConfig() *Config {
	return &Config{}
}
