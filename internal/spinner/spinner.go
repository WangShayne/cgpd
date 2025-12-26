package spinner

import (
	"fmt"
	"io"
	"sync"
	"time"
)

type Spinner struct {
	frames   []string
	interval time.Duration
	writer   io.Writer
	message  string
	stop     chan struct{}
	done     chan struct{}
	mu       sync.Mutex
	running  bool
}

func New(w io.Writer, message string) *Spinner {
	return &Spinner{
		frames:   []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		interval: 80 * time.Millisecond,
		writer:   w,
		message:  message,
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
	}
}

func (s *Spinner) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	go func() {
		defer close(s.done)
		i := 0
		for {
			select {
			case <-s.stop:
				fmt.Fprintf(s.writer, "\r\033[K")
				return
			default:
				fmt.Fprintf(s.writer, "\r%s %s", s.frames[i%len(s.frames)], s.message)
				i++
				time.Sleep(s.interval)
			}
		}
	}()
}

func (s *Spinner) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stop)
	<-s.done
}
