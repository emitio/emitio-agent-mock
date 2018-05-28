package main

import (
	"context"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/davecgh/go-spew/spew"
	"github.com/emitio/emitio-agent-mock/pkg/emitio/v1"
)

type server struct{}

func (s *server) Emit(ctx context.Context, req *emitio.EmitRequest) (*emitio.EmitResponse, error) {
	spew.Dump(req)
	return &emitio.EmitResponse{}, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-term
		cancel()
		<-term
		os.Exit(1)
	}()
	l, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(l)
	const (
		network = "tcp"
		address = "127.0.0.1:3648"
	)
	lis, err := net.Listen(network, address)
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("listening")
	}
	srv := grpc.NewServer()
	emitio.RegisterEmitIOServer(srv, &server{})
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()
	err = srv.Serve(lis)
	if err != nil {
		zap.L().With(zap.Error(err)).Fatal("serving")
	}
}
