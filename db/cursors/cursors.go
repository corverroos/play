package cursors

import (
	"database/sql"

	"github.com/luno/reflex"
	"github.com/luno/reflex/rsql"
)

// cursors wrap the play_cursors table providing a reflex cursor store for any
// and all consumers running in play.
var cursors = rsql.NewCursorsTable("play_cursors",
	rsql.WithCursorCursorField("`cursor`"))

// ToStore returns a reflex cursor store backed by the play_cursors table.
func ToStore(dbc *sql.DB) reflex.CursorStore {
	return cursors.ToStore(dbc, rsql.WithCursorAsyncDisabled()) // Have to disable async since it doesn't use fated context.
}
