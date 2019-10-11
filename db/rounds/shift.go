package rounds

import (
	"github.com/corverroos/play"
	"github.com/corverroos/play/db/events"
	"github.com/corverroos/play/internal"
	"github.com/luno/shift"
)

//go:generate shiftgen -inserter=joined -updaters=collected,submitted,excluded,failed,shared -table=play_rounds

var fsm = shift.NewFSM(events.GetTable()).
	Insert(play.StatusJoined, joined{},
		play.StatusExcluded, play.StatusCollected, play.StatusFailed).
	Update(play.StatusCollected, collected{},
		play.StatusShared, play.StatusSubmitted /* if single player */, play.StatusFailed).
	Update(play.StatusShared, shared{},
		play.StatusShared, play.StatusSubmitted, play.StatusFailed).
	Update(play.StatusSubmitted, submitted{},
		play.StatusFailed).
	Update(play.StatusExcluded, excluded{},
		play.StatusFailed).
	Update(play.StatusFailed, failed{}).
	Build()

type joined struct {
	ExternalID int64
	Version    int
	State      internal.RoundState
}

type collected struct {
	ID      int64
	Version int
	State   internal.RoundState
}

type shared struct {
	ID      int64
	Version int
	State   internal.RoundState
}

type submitted struct {
	ID      int64
	Version int
	State   internal.RoundState
}

type excluded struct {
	ID      int64
	Version int
}

type failed struct {
	ID      int64
	Version int
}
