package cmd

import (
	"errors"
	"fmt"
)

type ExitCode int

const (
	ExitStopped = 3
)

func (e ExitCode) Error() string { return fmt.Sprintf("exit code %d", int(e)) }

var ErrExit = errors.New("silent exit")
