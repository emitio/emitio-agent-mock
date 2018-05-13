package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/emitio/emitio-mock/pkg/emitio"
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
	defer func() {
		fmt.Printf("name=%s has gone away\n", name)
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
	beforeExit := func() {}
	defer beforeExit()
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-term
		cancel()
		<-term
		beforeExit()
		os.Exit(1)
	}()
	l, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(l)
	const (
		network = "unix"
		address = "/var/run/emitio/emitio.sock"
	)
	lis, err := net.Listen(network, address)
	if err != nil {
		beforeExit()
		zap.L().With(zap.Error(err)).Fatal("listening")
	}
	beforeExit = func() {
		os.Remove(address)
	}
	srv := grpc.NewServer()
	emitio.RegisterEmitIOServer(srv, &server{})
	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()
	err = srv.Serve(lis)
	if err != nil {
		beforeExit()
		zap.L().With(zap.Error(err)).Fatal("serving")
	}
}
