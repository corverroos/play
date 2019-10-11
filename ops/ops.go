package ops

import (
	"context"

	"github.com/corverroos/play"
	"github.com/corverroos/play/db/rounds"
	"github.com/luno/jettison/errors"
)

func GetRoundData(ctx context.Context, b Backends, roundID int64) (play.RoundData, error) {
	round, err := rounds.Lookup(ctx, b.PlayDB().DB, roundID)
	if err != nil {
		return play.RoundData{}, err
	}

	_, ps, ok := round.State.GetMine()
	if !ok {
		return play.RoundData{}, errors.New("No own player state yet?!")
	}

	return play.RoundData{
		ExternalID: round.ExternalID,
		Included:   ps.Included,
		Submitted:  ps.Submitted,
		Rank:       ps.Rank,
		Parts:      ps.Parts,
	}, nil
}
