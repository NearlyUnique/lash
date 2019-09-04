package lash

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type (
	// Scope for lash
	Scope struct {
		err            error
		onErr          OnErrorFunc
		stdout, stderr io.Writer
	}
	// ScopeErr an error that occurred during a scope operation
	ScopeErr struct {
		Type   string
		Action string
		Err    error
	}
	// OnErrorFunc perform some action
	OnErrorFunc func(error)
)

// NewScope for lash, you can have as many as you want, they are separate
// the default OnError handler is Terminate
func NewScope() *Scope {
	s := Scope{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	s.onErr = s.Terminate
	return &s
}

// IsError for this scope
func (s *Scope) IsError() bool {
	if s == nil {
		return false
	}
	return s.err != nil
}

// OnError do something
func (s *Scope) OnError(fn OnErrorFunc) *Scope {
	s.onErr = fn
	return s
}

// Err is the raw scope error if any
func (s *Scope) Err() error {
	return s.err
}

// SetErr is the raw scope error if any
func (s *Scope) SetErr(err error) {
	if s == nil || err == nil {
		return
	}
	serr, ok := err.(*ScopeErr)
	if ok && serr.Err == nil {
		return
	}
	s.err = err
	if s.onErr != nil {
		s.onErr(err)
	}
}

// ClearError removes the raw scope error if any
func (s *Scope) ClearError() {
	s.err = nil
}

// AsJson turns a map or struct (or json-able) thing into json buffer
func (s *Scope) AsJson(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		s.SetErr(&ScopeErr{Type: "Scope", Action: "AsJSON", Err: err})
		return nil
	}
	return b
}

// Env operations
func (s *Scope) Env() RequireEnv {
	return RequireEnv{
		scope: s,
	}
}

//func (s *Scope) Flags() RequireFlag {
//	return RequireFlag{
//		scope: s,
//		Flags:   flag.FlagSet{},
//	}
//}

// Args to the app
func (s *Scope) Args() RequireArg {
	return RequireArg{
		scope: s,
		Args:  os.Args,
	}
}

// SetOutput for Println etc. defaults to os.Stdout
func (s *Scope) SetOutput(writer io.Writer) *Scope {
	s.stdout = writer
	return s
}

// Println to stdout as set by scope.SetOutput
func (s *Scope) Println(msg string, args ...interface{}) {
	_, _ = fmt.Fprintf(s.stdout, s.EnvStr(msg, args...)+"\n")
}

// SetErrOutput for Error writing defaults to os.Stderr
func (s *Scope) SetErrOutput(writer io.Writer) *Scope {
	s.stderr = writer
	return s
}

// Terminate on error, with error code 1
func (s *Scope) Terminate(err error) {
	if err == nil {
		return
	}
	_, _ = fmt.Fprintln(s.stderr, err.Error())
	os.Exit(1)
}

// Warn about to standard out errors but continue
func (s *Scope) Warn(err error) {
	if err == nil {
		return
	}
	_, _ = fmt.Fprintln(s.stderr, err.Error())
}

// Ignore on error
func Ignore(error) {
}

// fail sets the error action and details
func (e *ScopeErr) fail(action string, err error) error {
	e.Action = action
	e.Err = err
	return e
}

// Error interface
func (e *ScopeErr) Error() string {
	return fmt.Sprintf("%s:%s:%v", e.Type, e.Action, e.Err)
}
