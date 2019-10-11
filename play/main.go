package main

import (
	"net/http"

	"github.com/corverroos/play"
	play_ops "github.com/corverroos/play/ops"
	"github.com/corverroos/play/playpb"
	play_server "github.com/corverroos/play/server"
	"github.com/corverroos/play/state"
	"github.com/corverroos/unsure"
	"github.com/luno/jettison/errors"
)

func main() {
	unsure.Bootstrap()

	s, err := state.New()
	if err != nil {
		unsure.Fatal(errors.Wrap(err, "new state error"))
	}

	go serveGRPCForever(s)

	play_ops.StartLoops(s)

	http.HandleFunc("/health", makeHealthCheckHandler())
	go unsure.ListenAndServeForever(play.HTTPAddr(play.Index()), nil)

	unsure.WaitForShutdown()
}

func serveGRPCForever(s *state.State) {
	grpcServer, err := unsure.NewServer(play.GRPCAddr(play.Index()))
	if err != nil {
		unsure.Fatal(errors.Wrap(err, "new grpctls server"))
	}

	playSrv := play_server.New(s)
	playpb.RegisterEngineServer(grpcServer.GRPCServer(), playSrv)

	unsure.RegisterNoErr(func() {
		playSrv.Stop()
		grpcServer.Stop()
	})

	unsure.Fatal(grpcServer.ServeForever())
}

func makeHealthCheckHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}
}
