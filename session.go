package lash

import (
	"fmt"
	"os"
)

type (
	// Session for lash
	Session struct {
		err   error
		onErr OnErrorFunc
	}
	// SessionErr an error that occurred during a session operation
	SessionErr struct {
		Type   string
		Action string
		Err    error
	}
	// OnErrorFunc perform some action
	OnErrorFunc func(error)
)

var (
	DefaultSession *Session
)

func init() {
	DefaultSession = NewSession()
}

// NewSession for lash, you can have as many as you want, they are separate
// the default OnError handler is Terminate
func NewSession() *Session {
	return &Session{
		onErr: Terminate,
	}
}

// Error for the global session
func Error() string {
	if DefaultSession != nil {
		return DefaultSession.ErrorString()
	}
	return ""
}

// ErrorString for this session
func (s *Session) ErrorString() string {
	if s.err != nil {
		return s.err.Error()
	}
	return ""
}

// IsError for this session
func (s *Session) IsError() bool {
	return s.err != nil
}

// OnError do something
func (s *Session) OnError(fn OnErrorFunc) *Session {
	s.onErr = fn
	return s
}

// Err is the raw session error if any
func (s *Session) Err() error {
	return s.err
}

// ClearError removes the raw session error if any
func (s *Session) ClearError() {
	s.err = nil
}

// Terminate on error, with error code 1
func Terminate(err error) {
	if err == nil {
		return
	}
	_, _ = fmt.Fprintln(os.Stderr, err.Error())
	os.Exit(1)
}

// Ignore errors, do nothing
func Ignore(err error) {}

// Warn about to standard out errors but continue
func Warn(err error) {
	if err == nil {
		return
	}
	_, _ = fmt.Fprintln(os.Stderr, err.Error())
}

// fail sets the error action and details
func (e *SessionErr) fail(action string, err error) error {
	e.Action = action
	e.Err = err
	return e
}

// Error interface
func (e *SessionErr) Error() string {
	return fmt.Sprintf("%s:%s:%v", e.Type, e.Action, e.Err)
}
