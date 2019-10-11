package play

//go:generate stringer -type=Status -trimprefix=Status

type Status int

func (t Status) Valid() bool {
	return t > StatusUnknown && t < StatusSentinel
}

func (t Status) ReflexType() int {
	return int(t)
}

func (t Status) Enum() int {
	return int(t)
}

func (t Status) ShiftStatus() {
}

const (
	StatusUnknown   Status = 0
	StatusJoined    Status = 1
	StatusExcluded  Status = 2
	StatusCollected Status = 3
	StatusShared    Status = 4
	StatusSubmitted Status = 5
	StatusFailed    Status = 6
	StatusSentinel  Status = 7
)

type RoundData struct {
	ExternalID int64
	Included   bool
	Submitted  bool
	Rank       int
	Parts      map[int]int
}
