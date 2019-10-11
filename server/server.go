package server

import (
	"context"

	"github.com/corverroos/play/db/events"
	"github.com/corverroos/play/ops"
	pb "github.com/corverroos/play/playpb"
	"github.com/corverroos/play/playpb/protocp"
	"github.com/corverroos/unsure"
	"github.com/luno/reflex"
	"github.com/luno/reflex/reflexpb"
)

var _ pb.EngineServer = (*Server)(nil)

// Server implements the play grpc server.
type Server struct {
	b       Backends
	rserver *reflex.Server
	stream  reflex.StreamFunc
}

// New returns a new server instance.
func New(b Backends) *Server {
	return &Server{
		b:       b,
		rserver: reflex.NewServer(),
		stream:  events.ToStream(b.PlayDB().DB),
	}
}

func (srv *Server) Stop() {
	srv.rserver.Stop()
}

func (srv *Server) Ping(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	return req, nil
}

func (srv *Server) Stream(req *reflexpb.StreamRequest, ss pb.Engine_StreamServer) error {
	return srv.rserver.Stream(srv.stream, req, ss)
}

func (srv *Server) GetRoundData(ctx context.Context, req *pb.GetRoundDataReq) (*pb.RoundData, error) {
	res, err := ops.GetRoundData(unsure.ContextWithFate(ctx, unsure.DefaultFateP()), srv.b, req.RoundID)
	if err != nil {
		return nil, err
	}
	return protocp.RoundDataToProto(&res), nil
}
