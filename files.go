package lash

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
)

type (
	// File reading/writing options
	File struct {
		path    string
		session *Session
	}
)

// OpenFile and do something with it
func OpenRead(name string) *File {
	s := DefaultSession
	if s == nil {
		s = NewSession()
	}
	return s.OpenFile(name)
}

// OpenFile and do something with it
func (s *Session) OpenFile(name string) *File {
	return &File{
		session: s,
		path:    name,
	}
}

// String content
func (f *File) String() string {
	if f.session.err != nil {
		return ""
	}
	b, err := ioutil.ReadFile(f.path)
	if err != nil {
		f.session.SetErr(&SessionErr{Type: "File", Action: "String", Err: err})
		return ""
	}
	return string(b)
}

// ReadLines read all lines via a channel one line at a lime
func (f *File) ReadLines() chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		file, err := os.Open(f.path)

		if err != nil {
			f.session.SetErr(&SessionErr{Type: "File", Action: "ReadLines", Err: err})
			return
		}
		defer func() { _ = file.Close() }()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			ch <- scanner.Text()
		}
	}()

	return ch
}

func (f *File) AppendLine(s string) {
	if f.session != nil && f.session.err != nil {
		return
	}
	file, err := os.OpenFile(f.path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		f.session.SetErr(&SessionErr{Type: "File", Action: "AppendLine", Err: err})
		return
	}
	defer func() { _ = file.Close() }()

	_, err = fmt.Fprintln(file, s)
	if err != nil {
		f.session.SetErr(&SessionErr{Type: "File", Action: "AppendLine", Err: err})
		return
	}
}
