package lash

import (
	"encoding/json"
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

// SetErr is the raw session error if any
func (s *Session) SetErr(err error) {
	if s == nil || err == nil {
		return
	}
	serr, ok := err.(*SessionErr)
	if ok && serr.Err == nil {
		return
	}
	s.err = err
	if s.onErr != nil {
		s.onErr(err)
	}
}

// ClearError removes the raw session error if any
func (s *Session) ClearError() {
	s.err = nil
}

// AsJson turns a map or struct (or json-able) thing into json buffer
func AsJson(v interface{}) []byte {
	s := DefaultSession
	if s == nil {
		s = NewSession()
	}
	return s.AsJson(v)
}

// AsJson turns a map or struct (or json-able) thing into json buffer
func (s *Session) AsJson(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		s.SetErr(&SessionErr{Type: "Session", Action: "AsJSON", Err: err})
		return nil
	}
	return b
}

// Env operations
func (s *Session) Env() RequireEnv {
	return RequireEnv{
		session: s,
	}
}

//func (s *Session) Flags() RequireFlag {
//	return RequireFlag{
//		session: s,
//		Flags:   flag.FlagSet{},
//	}
//}

// Args to the app
func (s *Session) Args() RequireArg {
	return RequireArg{
		session: s,
		Args:    os.Args,
	}

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
