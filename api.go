package play

import (
	"context"

	"github.com/luno/reflex"
)

// Client defines the root play service interface.
type Client interface {
	Ping(context.Context) error
	Stream(ctx context.Context, after string, opts ...reflex.StreamOption) (reflex.StreamClient, error)
	GetRoundData(ctx context.Context, roundID int64) (RoundData, error)
}
