package play

import (
	"flag"
	"fmt"
	"strconv"
	"strings"
)

var (
	index = flag.Int("index", 0, "index of the play instance")
	team  = flag.String("team", "play", "team name")
	count = flag.Int("count", 3, "number of players in the team")
)

func Name() string {
	return fmt.Sprintf("play%d", *index)
}

func NameToIndex(name string) (int, error) {
	index := strings.Replace(name, "play", "", 1)
	return strconv.Atoi(index)
}

func Index() int {
	return *index
}

func Team() string {
	return *team
}

func Count() int {
	return *count
}

const portOffset = 41566

func HTTPAddr(index int) string {
	return fmt.Sprintf("localhost:%d", portOffset+(index*2))
}

func GRPCAddr(index int) string {
	return fmt.Sprintf("localhost:%d", portOffset+(index*2)+1)
}
