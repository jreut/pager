// Package cmd is the interesting part between command line parsing and the persistence layer.
package cmd

import (
	"errors"
)

var ErrConflict = errors.New("conflict")
