package tea

import (
	"time"
)

// Cmd represents an asynchronous command that yields a Msg upon completion.
type Cmd func() Msg

// Batch combines multiple commands to execute concurrently.
func Batch(cmds ...Cmd) Cmd {
	valid := make([]Cmd, 0, len(cmds))
	for _, c := range cmds {
		if c != nil {
			valid = append(valid, c)
		}
	}
	if len(valid) == 0 {
		return nil
	}
	if len(valid) == 1 {
		return valid[0]
	}

	return func() Msg {
		return batchMsg(valid)
	}
}

type batchMsg []Cmd

// Sequence runs commands sequentially in a dedicated goroutine.
func Sequence(cmds ...Cmd) Cmd {
	return func() Msg {
		return sequenceMsg(cmds)
	}
}

type sequenceMsg []Cmd

// Tick returns a Cmd that sends a message produced by fn after duration d.
func Tick(d time.Duration, fn func(t time.Time) Msg) Cmd {
	return func() Msg {
		t := <-time.After(d)
		if fn == nil {
			return nil
		}
		return fn(t)
	}
}

// Quit returns a QuitMsg to stop the program loop.
func Quit() Msg {
	return QuitMsg{}
}

// execBatch executes batch messages concurrently and sends results to msgs chan.
func execBatch(cmds []Cmd, send func(Msg)) {
	for _, cmd := range cmds {
		if cmd == nil {
			continue
		}
		go func(c Cmd) {
			if msg := c(); msg != nil {
				send(msg)
			}
		}(cmd)
	}
}
