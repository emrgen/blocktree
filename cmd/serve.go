package cmd

import (
	"github.com/emrgen/blocktree"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newServeCmd())
}

func newServeCmd() *cobra.Command {
	var grpcPort, httpPost int
	// serveCmd represents the serve command
	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the blocktree server",
		Run: func(cmd *cobra.Command, args []string) {
			if grpcPort == httpPost {
				panic("gRPC and HTTP ports must be different")
			}

			if grpcPort < 1 || grpcPort > 65535 {
				grpcPort = 1000
			}

			if httpPost < 1 || httpPost > 65535 {
				httpPost = 1001
			}

			server := blocktree.NewServer(blocktree.NewMemStore(), &blocktree.Config{
				GrpcPort: grpcPort,
				HttpPort: httpPost,
			})

			err := server.Start()
			if err != nil {
				panic(err)
			}
		},
	}

	serveCmd.Flags().IntVarP(&grpcPort, "gport", "g", 1000, "gRPC port")
	serveCmd.Flags().IntVarP(&httpPost, "hport", "p", 1001, "HTTP port")

	return serveCmd
}
