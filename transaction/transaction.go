package transaction

import (
	"bufio"
	"fmt"
	"os"

	"github.com/masonbrott/sigilkv/core"
)

var TranLogger TransactionLogger

type TransactionLogger interface {
	WriteDelete(key string)
	WritePut(key, value string)
	Err() <-chan error

	ReadEvents() (<-chan Event, <-chan error)

	Run()
}

type Event struct {
	Sequence  uint64
	EventType EventType
	Key       string
	Value     string
}

type EventType byte

const (
	_                     = iota
	EventDelete EventType = iota
	EventPut
)

type FileTransactionLogger struct {
	events       chan<- Event
	errors       <-chan error
	lastSequence uint64
	file         *os.File
}

func (l *FileTransactionLogger) WritePut(key, value string) {
	l.events <- Event{EventType: EventPut, Key: key, Value: value}
}

func (l *FileTransactionLogger) WriteDelete(key string) {
	l.events <- Event{EventType: EventDelete, Key: key}
}

func (l *FileTransactionLogger) Err() <-chan error {
	return l.errors
}

func NewFileTransactionLogger(filename string) (TransactionLogger, error) {
	file, err := os.OpenFile(
		filename,
		os.O_RDWR|os.O_APPEND|os.O_CREATE, 0755)

	if err != nil {
		return nil, fmt.Errorf("cannot open transaction log file: %w", err)
	}

	return &FileTransactionLogger{file: file}, nil
}

func (l *FileTransactionLogger) Run() {
	events := make(chan Event, 16)
	l.events = events

	errors := make(chan error, 1)
	l.errors = errors

	go func() {
		for e := range events {
			l.lastSequence++

			_, err := fmt.Fprintf(
				l.file,
				"%d\t%d\t%s\t%s\n",
				l.lastSequence, e.EventType, e.Key, e.Value)

			if err != nil {
				errors <- err
				return
			}
		}
	}()
}

func (l *FileTransactionLogger) ReadEvents() (<-chan Event, <-chan error) {
	scanner := bufio.NewScanner(l.file)
	outEvent := make(chan Event)
	outError := make(chan error, 1)

	go func() {
		var e Event
		defer close(outEvent)
		defer close(outError)

		for scanner.Scan() {
			line := scanner.Text()
			if _, err := fmt.Sscanf(
				line, "%d\t%d\t%s\t%s",
				&e.Sequence, &e.EventType, &e.Key, &e.Value); err != nil {
				outError <- fmt.Errorf("input parse error: %w", err)
				return
			}

			if l.lastSequence >= e.Sequence {
				outError <- fmt.Errorf("transaction numbers out of sequence")
				return
			}

			l.lastSequence = e.Sequence
			outEvent <- e
		}

		if err := scanner.Err(); err != nil {
			outError <- fmt.Errorf("transaction log read failure: %w", err)
			return
		}
	}()

	return outEvent, outError
}

func InitializeTransactionLog() error {
	var err error

	TranLogger, err = NewFileTransactionLogger("transaction.log")
	if err != nil {
		return fmt.Errorf("failed to create event logger: %w", err)
	}

	events, errors := TranLogger.ReadEvents()

	e, ok := Event{}, true

	for ok && err == nil {
		select {
		case err, ok = <-errors:
		case e, ok = <-events:
			switch e.EventType {
			case EventDelete:
				err = core.Delete(e.Key)
			case EventPut:
				err = core.Put(e.Key, e.Value)
			}
		}
	}

	TranLogger.Run()

	return err
}
