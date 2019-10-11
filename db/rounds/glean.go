package rounds

import (
	"github.com/corverroos/play/internal"
)

// Note glean doesn't support non-bitx packages nicely.
//   mkdir -p /tmp/github.com/corverroos/play
//   ln -s {unsure_repo}/play/internal /tmp/github.com/corverroos/play/

//go:generate glean -table=play_rounds -src=/tmp

type glean struct {
	internal.Round
}
