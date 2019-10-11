package client

import (
	"context"
	"flag"

	"github.com/corverroos/play"
	pb "github.com/corverroos/play/playpb"
	"github.com/corverroos/play/playpb/protocp"
	"github.com/corverroos/unsure"
	"github.com/luno/reflex"
	"github.com/luno/reflex/reflexpb"
	"google.golang.org/grpc"
)

var addr = flag.String("play_address", "", "host:port of play gRPC service")

var _ play.Client = (*client)(nil)

type client struct {
	address   string
	rpcConn   *grpc.ClientConn
	rpcClient pb.EngineClient
}

func IsEnabled() bool {
	return *addr != ""
}

type option func(*client)

func WithAddress(address string) option {
	return func(c *client) {
		c.address = address
	}
}

func New(opts ...option) (*client, error) {
	c := client{
		address: *addr,
	}
	for _, o := range opts {
		o(&c)
	}

	var err error
	c.rpcConn, err = unsure.NewClient(c.address)
	if err != nil {
		return nil, err
	}

	c.rpcClient = pb.NewEngineClient(c.rpcConn)

	return &c, nil
}

func (c *client) Ping(ctx context.Context) error {
	_, err := c.rpcClient.Ping(ctx, &pb.Empty{})
	return err
}

func (c *client) Stream(ctx context.Context, after string, opts ...reflex.StreamOption) (reflex.StreamClient, error) {
	sFn := reflex.WrapStreamPB(func(ctx context.Context,
		req *reflexpb.StreamRequest) (reflex.StreamClientPB, error) {
		return c.rpcClient.Stream(ctx, req)
	})
	return sFn(ctx, after, opts...)
}

func (c *client) GetRoundData(ctx context.Context, roundID int64) (play.RoundData, error) {
	res, err := c.rpcClient.GetRoundData(ctx, &pb.GetRoundDataReq{
		RoundID: roundID,
	})
	if err != nil {
		return play.RoundData{}, err
	}
	return protocp.RoundDataFromProto(res), err
}
