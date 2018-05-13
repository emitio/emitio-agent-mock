package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/emitio/emitio-agent-mock/pkg/emitio"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type server struct{}

func (s *server) Emit(stream emitio.EmitIO_EmitServer) error {
	i, err := stream.Recv()
	if err != nil {
		return err
	}
	h, ok := i.Inputs.(*emitio.EmitInput_Header)
	if !ok {
		return errors.New("first message must be header")
	}
	name := h.Header.Name
	if name == "" {
		return errors.New("forwarder must specify name")
	}
	zap.L().With(zap.String("name", name)).Info("new stream")
	defer func() {
		zap.L().With(zap.String("name", name)).Info("stream has gone away")
	}()
	for {
		i, err := stream.Recv()
		if err != nil {
			return err
		}
		fmt.Printf("name=%s recv=%+v\n", name, i)
	}
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
