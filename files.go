package lash

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
)

type (
	// File reading/writing options
	File struct {
		path    string
		session *Session
		file    *os.File
		ch      chan string
	}
	fileAppender struct {
		file *File
		wg   *sync.WaitGroup
	}
	FileAppender interface {
		Ch() chan<- string
		AppendLine(line string)
		Close()
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
	f.open()

	_, err := fmt.Fprintln(f.file, s)
	if err != nil {
		f.session.SetErr(&SessionErr{Type: "File", Action: "AppendLine", Err: err})
		return
	}
}

//Truncate a file to zero length
func (f *File) Truncate() *File {
	f.open()
	err := &SessionErr{Type: "File", Action: "Truncate"}
	err.Err = f.file.Truncate(0)
	f.session.SetErr(err)

	_, err.Err = f.file.Seek(0, 0)
	f.session.SetErr(err)
	return f
}

//Appender channel to append lines to and a func to call when finished
func (f *File) Appender() FileAppender {
	if f.ch == nil {
		f.ch = make(chan string)
	}
	a := fileAppender{file: f, wg: &sync.WaitGroup{}}
	a.wg.Add(1)
	go func() {
		for line := range a.file.ch {
			a.file.AppendLine(line)
		}
		a.wg.Done()
	}()

	return a
}

// close the internal file if it is open
func (f *File) Close() {
	if f == nil || f.file == nil {
		return
	}
	err := f.file.Close()
	if err != nil {
		f.session.SetErr(&SessionErr{Type: "File", Action: "Close", Err: err})
	}
}

func (f *File) open() {
	var err error
	if f.file == nil {
		f.file, err = os.OpenFile(f.path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			f.session.SetErr(&SessionErr{Type: "File", Action: "AppendLine", Err: err})
			return
		}
	}
}

func (a fileAppender) Ch() chan<- string {
	return a.file.ch
}

func (a fileAppender) AppendLine(line string) {
	a.file.ch <- line
}

func (a fileAppender) Close() {
	if a.file.ch != nil {
		close(a.file.ch)
		a.wg.Wait()
	}

	if a.file != nil {
		a.file.Close()
	}
}
