package ops

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/corverroos/play"
	"github.com/corverroos/play/db/cursors"
	"github.com/corverroos/play/db/rounds"
	"github.com/corverroos/play/internal"
	"github.com/corverroos/unsure"
	"github.com/corverroos/unsure/engine"
	"github.com/luno/fate"
	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/j"
	"github.com/luno/jettison/log"
	"github.com/luno/reflex"
)

const (
	consumerJoinRounds       = "play_join_rounds"
	consumerCollectRound     = "play_collect_rounds"
	consumerSubmitNextRound  = "play_submit_next_rounds_%d"
	consumerSubmitFirstRound = "play_submit_first_rounds"
	consumerShareData        = "play_share_round_%d"
	consumerPlayLog          = "play_play_log_%d"
	consumerEngineLog        = "play_engine_log"
	consumerEngineEnded      = "play_engine_ended"
)

func StartLoops(b Backends) {

	go startMatchForever(b)

	reqs := []consumeReq{
		makeJoinRound(b),
		makeCollectRound(b),
		makeSubmitFirst(b),
		makeEngineLogger(b),
		makeExitOnEnded(b),
	}

	for i := 0; i < play.Count(); i++ {
		if i == play.Index() {
			continue
		}
		reqs = append(reqs, makeShareRound(b, i))
		reqs = append(reqs, makeSubmitNext(b, i))
		reqs = append(reqs, makePlayLogger(b, i))
	}

	for _, req := range reqs {
		startConsume(b, req)
	}
}

// makeSubmitNext returns a consumeReq that submits on StatusSubmitted if this player is next.
func makeSubmitNext(b Backends, index int) consumeReq {
	player := b.Players()[index]

	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsType(e.Type, play.StatusSubmitted) {
			return nil
		}

		data, err := player.GetRoundData(ctx, e.ForeignIDInt())
		if err != nil {
			return err
		}

		r, err := rounds.LookupByExternalID(ctx, b.PlayDB().DB, data.ExternalID)
		if err != nil {
			return err
		}

		switch r.Status {
		case play.StatusCollected:
			// I'm still getting shared data. Push back a bit.
			return fate.ErrTempt
		case play.StatusShared:
			// Ready to submit!
		case play.StatusSubmitted:
			// mmm, late, ignore.
			return nil
		case play.StatusExcluded:
			// I'm not playing :(
			return nil
		default:
			return errors.New("unexpected status", j.KV("status", r.Status))
		}

		if !r.State.HasAll() {
			log.Info(ctx, "waiting for data until next submit")
			return fate.ErrTempt
		}

		i, ps, ok := r.State.GetMine()
		if !ok {
			return errors.New("missing my state")
		}

		nextRank, ok := r.State.NextRank(index)
		if !ok {
			// No subsequent player
			return nil
		} else if ps.Rank != nextRank {
			// I'm not next
			return nil
		}

		log.Info(ctx, "submitting next round", j.MKV{"foreign_id": r.ExternalID, "total": r.State.GetTotal()})

		err = b.Engine().SubmitRound(ctx, play.Team(), play.Name(), r.ExternalID, r.State.GetTotal())
		if err != nil {
			return err
		}

		r.State.Players[i].Submitted = true

		err = rounds.ToSubmitted(ctx, b.PlayDB().DB, r.ID, r.Status, r.Version, r.State)
		if err != nil {
			return err
		}

		log.Info(ctx, "submitted next round", j.MKV{"foreign_id": r.ExternalID})

		return fate.Tempt()
	}

	name := reflex.ConsumerName(fmt.Sprintf(consumerSubmitNextRound, index))

	return newConsumeReq(player.Stream, name, f)
}

// makeShareRound returns a consumeReq that update rounds state on StatusCollected if not already shared.
func makeShareRound(b Backends, index int) consumeReq {
	player := b.Players()[index]

	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsAnyType(e.Type, play.StatusCollected, play.StatusExcluded) {
			return nil
		}

		log.Info(ctx, "getting shared player data", j.MKV{"other": index})

		data, err := player.GetRoundData(ctx, e.ForeignIDInt())
		if err != nil {
			return err
		}

		r, err := rounds.LookupByExternalID(ctx, b.PlayDB().DB, data.ExternalID)
		if err != nil {
			return err
		}

		switch r.Status {
		case play.StatusJoined:
			// I'm still collecting, push back a bit
			return fate.ErrTempt
		case play.StatusCollected, play.StatusShared:
			// Yeah, shared!
		case play.StatusSubmitted:
			// mmm, late, ignore.
			return nil
		case play.StatusExcluded:
			// I'm not playing :(
			return nil
		default:
			return errors.New("unexpected status", j.KV("status", r.Status))
		}

		_, _, ok := r.State.GetPlayer(index)
		if ok {
			// Already shared this state
			return nil
		}

		ps := internal.RoundPlayerState{
			Rank:      data.Rank,
			Index:     index,
			Included:  data.Included,
			Collected: true,
			Submitted: data.Submitted,
			Parts:     data.Parts,
		}
		r.State.Players = append(r.State.Players, ps)

		err = rounds.ToShared(ctx, b.PlayDB().DB, r.ID, r.Status, r.Version, r.State)
		if err != nil {
			return err
		}

		log.Info(ctx, "shared player data", j.MKV{"other": index, "included": data.Included})

		return fate.Tempt()
	}

	name := reflex.ConsumerName(fmt.Sprintf(consumerShareData, index))

	return newConsumeReq(player.Stream, name, f)
}

// makePlayLogger returns a consumeReq that logs play events.
func makePlayLogger(b Backends, index int) consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		typ := play.Status(e.Type.ReflexType())
		log.Info(ctx, "play event",
			j.MKV{"index": index, "id": e.ForeignIDInt(), "type": typ, "latency": time.Since(e.Timestamp)})
		return fate.Tempt()
	}

	name := reflex.ConsumerName(fmt.Sprintf(consumerPlayLog, index))

	return newConsumeReq(b.Players()[index].Stream, name, f)
}

// makeExitOnEnded returns a consumeReq that exists on match ended.
func makeExitOnEnded(b Backends) consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsType(e.Type, engine.EventTypeMatchEnded) {
			return nil
		}
		log.Info(ctx, "match ended!")
		unsure.Fatal(errors.New("exit on end :)"))
		return nil
	}

	name := reflex.ConsumerName(consumerEngineEnded)

	return newConsumeReq(b.Engine().Stream, name, f)
}

// makeEngineLogger returns a consumeReq that logs engine events.
func makeEngineLogger(b Backends) consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		typ := engine.EventType(e.Type.ReflexType())
		log.Info(ctx, "engine event",
			j.MKV{"foreign_id": e.ForeignIDInt(), "type": typ, "latency": time.Since(e.Timestamp)})
		return fate.Tempt()
	}

	name := reflex.ConsumerName(consumerEngineLog)

	return newConsumeReq(b.Engine().Stream, name, f)
}

// makeSubmitFirst returns a consumeReq that submits on EventTypeRoundSubmit if first.
func makeSubmitFirst(b Backends) consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsType(e.Type, engine.EventTypeRoundSubmit) {
			return nil
		}

		r, err := getRoundFromEvent(ctx, b, e)
		if errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(err, "missing round")
		} else if err != nil {
			return err
		}

		switch r.Status {
		case play.StatusJoined:
			// I missed some update, push back a bit.
			return fate.ErrTempt
		case play.StatusCollected:
			// Yeah, maybe submit!
		case play.StatusShared:
			// Yeah, maybe submit!
		case play.StatusSubmitted:
			// Reprocessing event, we are done.
			return nil
		case play.StatusExcluded:
			// We are not playing :(
			return nil
		default:
			return errors.New("unexpected status", j.KV("status", r.Status))
		}

		if !r.State.HasAll() {
			log.Info(ctx, "waiting for data until first submit")
			return fate.ErrTempt
		}

		i, ps, ok := r.State.GetMine()
		if !ok {
			return errors.New("missing player state")
		}

		firstRank := r.State.FirstRank()
		if ps.Rank != firstRank {
			// I'm not first
			return nil
		}

		log.Info(ctx, "submitting first round", j.MKV{"foreign_id": r.ExternalID, "total": r.State.GetTotal()})

		err = b.Engine().SubmitRound(ctx, play.Team(), play.Name(), r.ExternalID, r.State.GetTotal())
		if err != nil {
			return err
		}

		r.State.Players[i].Submitted = true

		err = rounds.ToSubmitted(ctx, b.PlayDB().DB, r.ID, r.Status, r.Version, r.State)
		if err != nil {
			return err
		}

		log.Info(ctx, "submitted first round", j.MKV{"foreign_id": r.ExternalID})

		return fate.Tempt()
	}

	name := reflex.ConsumerName(consumerSubmitFirstRound)

	return newConsumeReq(b.Engine().Stream, name, f)
}

// makeCollectRound returns a consumeReq that collects rounds on EventTypeRoundCollect if included not already collected.
func makeCollectRound(b Backends) consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsType(e.Type, engine.EventTypeRoundCollect) {
			return nil
		}

		eid := e.ForeignIDInt()

		r, err := getRoundFromEvent(ctx, b, e)
		if errors.Is(err, sql.ErrNoRows) {
			return errors.Wrap(err, "missing round")
		} else if err != nil {
			return err
		}

		switch r.Status {
		case play.StatusJoined:
			// Yeah, collect!
		case play.StatusCollected, play.StatusShared, play.StatusSubmitted:
			// Reprocessing event, we are done.
			return nil
		case play.StatusExcluded:
			// I'm not playing :(
			return nil
		default:
			return errors.New("unexpected status", j.KV("status", r.Status))
		}

		i, ps, ok := r.State.GetMine()
		if !ok {
			return errors.New("missing player state")
		}

		if !ps.Included {
			// Go directly to jail; do not pass go, do not collect $200
			return rounds.JoinedToExcluded(ctx, b.PlayDB().DB, eid, r.Version)
		}

		log.Info(ctx, "collecting round", j.MKV{"foreign_id": eid})

		res, err := b.Engine().CollectRound(ctx, play.Team(), play.Name(), eid)
		if err != nil {
			return err
		}

		parts := make(map[int]int)
		for _, pl := range res.Players {
			index, err := play.NameToIndex(pl.Name)
			if err != nil {
				return errors.Wrap(err, "invalid name")
			}
			parts[index] = pl.Part
		}
		r.State.Players[i].Rank = res.Rank
		r.State.Players[i].Parts = parts

		err = rounds.JoinedToCollected(ctx, b.PlayDB().DB, eid, r.Version, r.State)
		if err != nil {
			return err
		}

		log.Info(ctx, "collected round", j.MKV{"foreign_id": eid, "rank": res.Rank, "parts": parts})

		return fate.Tempt()
	}

	name := reflex.ConsumerName(consumerCollectRound)

	return newConsumeReq(b.Engine().Stream, name, f)
}

// makeJoinRound returns a consumeReq that joins rounds on EventTypeRoundJoin if not already joined.
func makeJoinRound(b Backends) consumeReq {
	f := func(ctx context.Context, f fate.Fate, e *reflex.Event) error {
		if !reflex.IsType(e.Type, engine.EventTypeRoundJoin) {
			return nil
		}

		eid := e.ForeignIDInt()

		// Ensure we have not started yet.
		_, err := getRoundFromEvent(ctx, b, e)
		if errors.Is(err, sql.ErrNoRows) {
			// We do not expect a round yet, noop.
		} else if err != nil {
			return err
		} else {
			// Round exists, we are cool.
			return nil
		}

		log.Info(ctx, "joining round", j.MKV{"foreign_id": eid})

		incl, err := b.Engine().JoinRound(ctx, play.Team(), play.Name(), eid)
		if errors.Is(err, engine.ErrAlreadyJoined) {
			incl = true
		} else if errors.Is(err, engine.ErrAlreadyExcluded) {
			incl = false
		} else if err != nil {
			return err
		}

		err = rounds.Joined(ctx, b.PlayDB().DB, eid, incl)
		if err != nil {
			return err
		}

		log.Info(ctx, "joined round", j.MKV{"foreign_id": eid, "included": incl})

		return fate.Tempt()
	}

	name := reflex.ConsumerName(consumerJoinRounds)

	return newConsumeReq(b.Engine().Stream, name, f)
}

func getRoundFromEvent(ctx context.Context, b Backends, e *reflex.Event) (*internal.Round, error) {
	return rounds.LookupByExternalID(ctx, b.PlayDB().DB, e.ForeignIDInt())
}

func startMatchForever(b Backends) {
	for {
		ctx := unsure.ContextWithFate(context.Background(), unsure.DefaultFateP())

		err := b.Engine().StartMatch(ctx, play.Team(), play.Count())

		if errors.Is(err, engine.ErrActiveMatch) {
			// Match active, just ignore
		} else if err != nil {
			log.Error(ctx, errors.Wrap(err, "start match error"))
		} else {
			log.Info(ctx, "match started")
		}

		time.Sleep(time.Second * 10)
	}
}

type consumeReq struct {
	stream reflex.StreamFunc
	name   reflex.ConsumerName
	f      func(ctx context.Context, f fate.Fate, e *reflex.Event) error
	copts  []reflex.ConsumerOption
	sopts  []reflex.StreamOption
}

func newConsumeReq(stream reflex.StreamFunc, name reflex.ConsumerName, f func(ctx context.Context, f fate.Fate, e *reflex.Event) error,
	opts ...reflex.StreamOption) consumeReq {
	return consumeReq{
		stream: stream,
		name:   name,
		f:      f,
		sopts:  opts,
	}
}

func startConsume(b Backends, req consumeReq) {
	consumer := reflex.NewConsumer(req.name, req.f, req.copts...)
	consumable := reflex.NewConsumable(req.stream, cursors.ToStore(b.PlayDB().DB))
	go unsure.ConsumeForever(unsure.FatedContext, consumable.Consume, consumer, req.sopts...)
}
