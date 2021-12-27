package watch

import (
	"log"
)

// Shutdown is an interface that provides a safe shutdown mechansim.
// The interface is compatible with tomb.Tomb.
//
// Most channel operations should include a select on Shutdown.Dying()
// inorder to cleanly exit:
//
// select {
//   case <-shutdown.Dying():
//     return
// }
//
type Shutdown interface {
	// Waits for a close event.
	Dying() <-chan struct{}

	// Kill the application, which triggers all waiters on shutdown.Dying()
	// to receive a close event.
	Kill(error)
}

type shutdownImpl struct {
	done chan struct{}
}

func (s *shutdownImpl) Dying() <-chan struct{} {
	return s.done
}

func (s *shutdownImpl) Kill(reason error) {
	select {
	case <-s.done:
		// closed
		return
	default:
		log.Println("shutdown: ", reason)
		close(s.done)
	}
}

// NewShutdown creates a shutdown object which encapsulates a single
// channel used to signal that the process is terminating.
func NewShutdown() Shutdown {
	return &shutdownImpl{make(chan struct{}, 1)}
}
