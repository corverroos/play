package rounds

import (
	"context"
	"database/sql"

	"github.com/corverroos/play"
	"github.com/corverroos/play/internal"
	"github.com/luno/jettison/errors"
	"github.com/luno/shift"
)

func LookupByExternalID(ctx context.Context, dbc *sql.DB, externalID int64) (*internal.Round, error) {
	return lookupWhere(ctx, dbc, "external_id=?", externalID)
}

func Joined(ctx context.Context, dbc *sql.DB, externalID int64, included bool) error {
	_, err := fsm.Insert(ctx, dbc, joined{
		ExternalID: externalID,
		State: internal.RoundState{Players: []internal.RoundPlayerState{{
			Index:    play.Index(),
			Included: included,
		}}},
	})
	return err
}

func JoinedToExcluded(ctx context.Context, dbc *sql.DB, id int64,
	prevVersion int) error {

	return to(ctx, dbc, id, play.StatusJoined, play.StatusExcluded, prevVersion,
		excluded{ID: id, Version: prevVersion + 1})
}

func JoinedToCollected(ctx context.Context, dbc *sql.DB, id int64, prevVersion int,
	newState internal.RoundState) error {

	return to(ctx, dbc, id, play.StatusJoined, play.StatusCollected,
		prevVersion, collected{ID: id, State: newState, Version: prevVersion + 1})
}

func ToShared(ctx context.Context, dbc *sql.DB, id int64, from play.Status,
	prevVersion int, newState internal.RoundState) error {

	return to(ctx, dbc, id, from, play.StatusShared, prevVersion,
		shared{ID: id, State: newState, Version: prevVersion + 1})
}

func ToSubmitted(ctx context.Context, dbc *sql.DB, id int64, from play.Status,
	prevVersion int, newState internal.RoundState) error {
	return to(ctx, dbc, id, from, play.StatusSubmitted,
		prevVersion, submitted{ID: id, State: newState, Version: prevVersion + 1})
}

func ToFailed(ctx context.Context, dbc *sql.DB, id int64, from play.Status,
	prevVersion int, errMsg string) error {

	return to(ctx, dbc, id, from, play.StatusFailed, prevVersion,
		failed{ID: id, Version: prevVersion + 1})
}

func ensurePrevVersion(ctx context.Context, tx *sql.Tx, id int64, version int) error {
	var n int
	err := tx.QueryRowContext(ctx, "select exists (select 1 from play_rounds "+
		"where id=? and version=?)", id, version).Scan(&n)
	if err != nil {
		return err
	}

	if n != 1 {
		return errors.New("concurrent update")
	}

	return nil
}

func to(ctx context.Context, dbc *sql.DB, id int64, from, to play.Status,
	prevVersion int, req shift.Updater) error {

	tx, err := dbc.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = ensurePrevVersion(ctx, tx, id, prevVersion)
	if err != nil {
		return err
	}

	notify, err := fsm.UpdateTx(ctx, tx, from, to, req)
	if err != nil {
		return err
	}
	defer notify()

	return tx.Commit()
}
