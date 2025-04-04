package blocktree

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	v1 "github.com/emrgen/blocktree/apis/v1"
	"github.com/gobuffalo/packr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
)

// Server is the blocktree server
type Server struct {
	Config *Config
	store  Store
}

// NewServer creates a new server
func NewServer(store Store, config *Config) *Server {
	return &Server{
		store:  store,
		Config: config,
	}
}

// Start starts the server
func (s *Server) Start() error {
	grpcPort := fmt.Sprintf(":%d", s.Config.GrpcPort)
	httpPort := fmt.Sprintf(":%d", s.Config.HttpPort)

	// Create the gRPC server
	grpcServer := grpc.NewServer()

	// Connect the rest gateway to the grpc server
	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:   true,
					EmitUnpopulated: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}
	endpoint := "localhost" + grpcPort

	api := NewApiWithPublisher(s.store, NewNullPublisher())

	// Register the server with the gRPC server
	v1.RegisterBlocktreeServer(grpcServer, newGrpcApi(api))
	// Register the server with the rest gateway
	if err := v1.RegisterBlocktreeHandlerFromEndpoint(context.TODO(), mux, endpoint, opts); err != nil {
		return err
	}

	// Serve the OpenAPI documentation
	apiMux := http.NewServeMux()
	openAPIBox := packr.NewBox("docs/v1")
	docsPath := "/v1/docs/"
	apiMux.Handle(docsPath, http.StripPrefix(docsPath, http.FileServer(openAPIBox)))
	apiMux.Handle("/", mux)

	// Add CORS support
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Authorization"},
		AllowCredentials: true,
	})

	// Create the HTTP server
	httpServer := &http.Server{
		Addr:    httpPort,
		Handler: c.Handler(apiMux),
	}

	// make sure to wait for the servers to stop before exiting
	var wg sync.WaitGroup
	gl, err := net.Listen("tcp", grpcPort)
	if err != nil {
		return err
	}

	rl, err := net.Listen("tcp", httpPort)
	if err != nil {
		return err
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		logrus.Info("starting rest gateway on: ", httpPort)
		logrus.Info("click on the following link to view the API documentation: http://localhost", httpPort, "/v1/docs/")
		if err := httpServer.Serve(rl); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				logrus.Errorf("error starting rest gateway: %v", err)
			}
		}
		logrus.Infof("rest gateway stopped")
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		logrus.Info("starting grpc server on: ", grpcPort)
		if err := grpcServer.Serve(gl); err != nil {
			logrus.Infof("grpc failed to start: %v", err)
		}
		logrus.Infof("grpc server stopped")
	}()

	time.Sleep(1 * time.Second)
	logrus.Infof("Press Ctrl+C to stop the server")

	// listen for interrupt signal to gracefully shut down the server
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, unix.SIGTERM, unix.SIGINT, unix.SIGTSTP)
	<-sigs
	// clean Ctrl+C output
	fmt.Println()

	logrus.Info("shutting down server")

	grpcServer.Stop()
	err = httpServer.Shutdown(context.Background())
	if err != nil {
		logrus.Errorf("error stopping rest gateway: %v", err)
	}

	// wait for the servers to stop
	wg.Wait()

	return nil
}
