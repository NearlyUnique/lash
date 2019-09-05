package lash

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/xerrors"
)

type (
	// File reading/writing options
	File struct {
		path  string
		scope *Scope
		file  *os.File
		ch    chan string
	}
	fileAppender struct {
		file *File
		wg   *sync.WaitGroup
	}
	//FileAppender for concurrently appending to a file
	FileAppender interface {
		Ch() chan<- string
		AppendLine(line string, args ...interface{})
		Close()
	}
)

// OpenFile and do something with it
func (s *Scope) OpenFile(name string, args ...interface{}) *File {
	return &File{
		scope: s,
		path:  s.EnvStr(name, args...),
	}
}

// String content
func (f *File) String() string {
	if f.scope.err != nil {
		return ""
	}
	b, err := ioutil.ReadFile(f.path)
	if err != nil {
		f.scope.SetErr(&ScopeErr{Type: "File", Action: "String", Err: xerrors.Errorf("path '%s': %w", f.path, err)})
		return ""
	}
	return string(b)
}

// AsJSON reads the file into v (a pointer to a struct)
func (f *File) AsJSON(v interface{}) {
	b, err := ioutil.ReadFile(f.path)
	if err != nil {
		f.scope.SetErr(&ScopeErr{Type: "File", Action: "AsJSON_read", Err: xerrors.Errorf("path '%s': %w", f.path, err)})
		return
	}
	err = json.Unmarshal(b, v)
	if err != nil {
		f.scope.SetErr(&ScopeErr{Type: "File", Action: "AsJSON_unmarshal", Err: xerrors.Errorf("path '%s': %w", f.path, err)})
	}
}

// ReadLines read all lines via a channel one line at a lime
func (f *File) ReadLines() chan string {
	ch := make(chan string)

	go func() {
		defer close(ch)
		file, err := os.Open(f.path)

		if err != nil {
			f.scope.SetErr(&ScopeErr{Type: "File", Action: "ReadLines", Err: err})
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

func (f *File) AppendLine(s string, args ...interface{}) *File {
	if f.scope != nil && f.scope.err != nil {
		return f
	}
	f.open(openBasic)

	_, err := fmt.Fprintln(f.file, f.scope.EnvStr(s, args...))
	if err != nil {
		f.scope.SetErr(&ScopeErr{Type: "File", Action: "AppendLine", Err: err})
	}
	return f
}

//Truncate a file to zero length
func (f *File) Truncate() *File {
	if f.isOpen() {
		err := &ScopeErr{Type: "File", Action: "Truncate"}
		// use underlying file close so we can set the correct error context
		err.Err = f.file.Close()
		f.scope.SetErr(err)
	}
	f.open(openTruncate)
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
		f.scope.SetErr(&ScopeErr{Type: "File", Action: "Close", Err: err})
	}
}

type openFlag int

const (
	openBasic    openFlag = openFlag(os.O_RDWR | os.O_CREATE | os.O_APPEND)
	openTruncate          = openFlag(int(openBasic) | os.O_TRUNC)
)

func (f *File) isOpen() bool {
	return f.file != nil
}
func (f *File) open(flag openFlag) {
	var err error
	if f.file != nil {
		return
	}
	f.file, err = os.OpenFile(f.path, int(flag), 0666)
	f.scope.setErr("File", "AppendLine", err)
}

// Delete the file
func (f *File) Delete() {
	f.Close()
	err := os.Remove(f.path)
	f.scope.setErr("File", "Delete", err)
}

// Mkdir creates the full path to ensure the supplied folder exists
func (f *File) Mkdir() {
	err := os.MkdirAll(f.path, 0666)
	f.scope.setErr("File", "Mkdir", err)
}

// CopyTo a destination, returns the destination
func (f *File) CopyTo(dest string) *File {
	err := copyFile(f.path, dest)
	f.scope.setErr("File", "Copy", err)
	return f.scope.OpenFile(dest)
}

// Ch the underlying channel for appending lines to
func (a fileAppender) Ch() chan<- string {
	return a.file.ch
}

// AppendLine wraps sending to the append channel, Ch
func (a fileAppender) AppendLine(line string, args ...interface{}) {
	a.file.ch <- a.file.scope.EnvStr(line, args...)
}

// Close the underlying channel and the target file
func (a fileAppender) Close() {
	if a.file.ch != nil {
		close(a.file.ch)
		a.wg.Wait()
	}

	if a.file != nil {
		a.file.Close()
		a.file = nil
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	tmp, err := ioutil.TempFile(filepath.Dir(dst), "")
	if err != nil {
		return err
	}
	_, err = io.Copy(tmp, in)
	if err != nil {
		tmp.Close()
		os.Remove(tmp.Name())
		return err
	}
	if err = tmp.Close(); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	perms, err := in.Stat()
	if err != nil {
		return err
	}
	if err = os.Chmod(tmp.Name(), perms.Mode()); err != nil {
		os.Remove(tmp.Name())
		return err
	}
	return os.Rename(tmp.Name(), dst)
}
