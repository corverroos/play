package internal

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"math"
	"time"

	"github.com/corverroos/play"
)

type Round struct {
	ID         int64
	ExternalID int64
	Status     play.Status
	CreatedAt  time.Time
	UpdatedAt  time.Time
	Version    int
	State      RoundState
}

type RoundState struct {
	Players []RoundPlayerState
}

func (rs RoundState) Value() (driver.Value, error) {
	return json.MarshalIndent(rs, "", " ")
}

func (rs *RoundState) Scan(src interface{}) error {
	var s sql.NullString
	if err := s.Scan(src); err != nil {
		return err
	}
	*rs = RoundState{}
	if !s.Valid {
		return nil
	}
	return json.Unmarshal([]byte(s.String), rs)
}

func (rs RoundState) GetMine() (int, RoundPlayerState, bool) {
	return rs.GetPlayer(play.Index())
}

func (rs RoundState) GetPlayer(index int) (int, RoundPlayerState, bool) {
	for i, m := range rs.Players {
		if m.Index == index {
			return i, m, true
		}
	}
	return 0, RoundPlayerState{}, false
}

func (rs RoundState) HasAll() bool {
	return len(rs.Players) == play.Count()
}

func (rs RoundState) FirstRank() int {
	minRank := int(math.MaxInt32)
	for _, m := range rs.Players {
		if m.Rank < minRank && m.Included {
			minRank = m.Rank
		}
	}

	return minRank
}

func (rs RoundState) NextRank(index int) (int, bool) {
	var thisRank int
	for _, m := range rs.Players {
		if m.Index == index {
			thisRank = m.Rank
		}
	}

	nextRank := int(math.MaxInt32)
	var has bool
	for _, m := range rs.Players {
		if m.Rank > thisRank && m.Rank < nextRank && m.Included {
			nextRank = m.Rank
			has = true
		}
	}

	return nextRank, has
}

func (rs RoundState) GetTotal() int {
	var total int
	for _, m := range rs.Players {
		if !m.Included {
			continue
		}
		total += m.Parts[play.Index()]
	}
	return total
}

type RoundPlayerState struct {
	Index     int
	Rank      int
	Parts     map[int]int
	Included  bool
	Collected bool
	Submitted bool
}
