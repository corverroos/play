package state

import (
	"github.com/corverroos/play"
	"github.com/corverroos/play/client"
	"github.com/corverroos/play/db"
	"github.com/corverroos/unsure/engine"
	ec "github.com/corverroos/unsure/engine/client"
)

type State struct {
	playDB       *db.PlayDB
	players      map[int]play.Client
	engineClient engine.Client
}

func (s *State) PlayDB() *db.PlayDB {
	return s.playDB
}

func (s *State) Players() map[int]play.Client {
	return s.players
}

func (s *State) Engine() engine.Client {
	return s.engineClient
}

// New returns a new play state.
func New() (*State, error) {
	var (
		s   State
		err error
	)

	s.playDB, err = db.Connect()
	if err != nil {
		return nil, err
	}

	s.players = make(map[int]play.Client)

	for i := 0; i < play.Count(); i++ {
		if i == play.Index() {
			continue
		}

		cl, err := client.New(client.WithAddress(play.GRPCAddr(i)))
		if err != nil {
			return nil, err
		}
		s.players[i] = cl
	}

	s.engineClient, err = ec.New()
	if err != nil {
		return nil, err
	}

	return &s, nil
}
