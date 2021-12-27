package watch

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrTailShutdown  = errors.New("tail: shutdown")
)

// Tail maintains a tail-capable file buffer.
type Tail struct {
	Filename string

	file    *os.File
	reader  *bufio.Reader
	lastPos int64
}

func TailFile(filename string) (*Tail, error) {
	// Tail the file.
	t := &Tail{
		Filename: filepath.Clean(filename),
	}

	// If the file does not exist, return an error.
	var err error
	t.file, err = MyOpenFile(t.Filename)
	if err != nil {
		return nil, err
	}
	t.reader = bufio.NewReader(t.file)
	return t, nil
}

func (t *Tail) Close() {
	if t.file != nil {
		t.file.Close()
		t.file = nil
	}
	t.reader = nil
}

func (t *Tail) resetAtOffset(offset int64) error {
	_, err := t.file.Seek(offset, 0)
	if err == nil {
		t.reader.Reset(t.file)
	}
	return err
}

func (t *Tail) ProcessLines(process func(line string) error) error {
	var offset int64
	var err error
	var line string

	// Read the file line by line
	for {
		// Grab the current offset.
		t.lastPos, err = t.file.Seek(0, os.SEEK_CUR)
		if err != nil {
			return err
		}
		offset = t.lastPos - int64(t.reader.Buffered())

		// Read the next line.
		line, err = t.reader.ReadString('\n')
		if err == nil {
			if err = process(line); err != nil {
				return err
			}
		} else if err == io.EOF {
			// End of file reached, maybe with a partial line read.
			if line != "" {
				t.resetAtOffset(offset)
			}
		}
		if err != nil {
			return err
		}
	}
	panic("unreachable")
}

// Read all available lines from the file, and advance
// the position of the Tail.
func (t *Tail) ReadLines() ([]string, error) {
	var result []string
	err := t.ProcessLines(func(l string) error{
		result = append(result, l)
		return nil
	})
	return result, err
}

// Send all available lines from the file, and advance
// the position of the Tail.
func (t *Tail) SendLines(lines chan string, shutdown Shutdown) error {
	return t.ProcessLines(func(l string) error{
		select {
		case lines <- l:
			return nil
		case <-shutdown.Dying():
			return ErrTailShutdown
		}
	})
}
