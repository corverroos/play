package ops

import (
	"github.com/corverroos/play"
	"github.com/corverroos/play/db"
	"github.com/corverroos/unsure/engine"
)

type Backends interface {
	Engine() engine.Client
	PlayDB() *db.PlayDB
	Players() map[int]play.Client
}
