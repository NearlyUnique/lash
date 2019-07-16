package lash

import (
	"bufio"
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

// OpenRead and do something with it
func OpenRead(name string) *File {
	s := DefaultSession
	if s == nil {
		s = NewSession()
	}
	return s.OpenRead(name)
}

// OpenRead and do something with it
func (s *Session) OpenRead(name string) *File {
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
		f.session.err = &SessionErr{Type: "File", Action: "String", Err: err}
		return ""
	}
	return string(b)
}

// ReadLines read all lines via a channel one line at a lime
func (f *File) ReadLines() chan (string) {
	ch := make(chan (string))

	go func() {
		defer close(ch)
		file, err := os.Open(f.path)

		if err != nil {
			f.session.err = &SessionErr{Type: "File", Action: "ReadLines", Err: err}
			return
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		for scanner.Scan() {
			ch <- scanner.Text()
		}
	}()

	return ch
}
