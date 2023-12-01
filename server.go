package blocktree

type Server struct {
	Config *Config
}

func NewServer(config *Config) *Server {
	return &Server{
		Config: config,
	}
}

func (s *Server) Start() error {
	return nil
}
