package defaults

import (
	"math"
)

const (
	// https://github.com/riverqueue/river/discussions/475
	// The database column is an int16
	MaxAttempts = math.MaxInt16

	// We assume if a fluxion cancel is not successful, it was already deleted
	MaxCancelAttempts = 3

	// Default duration is 0 (unset) so we honor kubernetes objects
	DefaultDuration = 0
)
