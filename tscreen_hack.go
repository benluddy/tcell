// +build tcellhack

package tcell

import (
	"io"
)

type ScreenDriver interface {
	ReadCloser() io.ReadCloser
	WriteCloser() io.WriteCloser
	Size() (int, int)
	Resized() <-chan struct{}
}

type ResizeSignal struct{}

func (ResizeSignal) String() string {
	return "Resize"
}

func (ResizeSignal) Signal() {
}

func NewTerminfoScreenWithDriver(driver ScreenDriver) (Screen, error) {
	s, e := NewTerminfoScreen()
	if e != nil {
		return nil, e
	}
	t, ok := s.(*tScreen)
	if !ok {
		return nil, ErrNoScreen
	}
	t.tiosp = &termiosPrivate{
		driver: driver,
		done:   make(chan struct{}),
	}
	return t, nil
}

type termiosPrivate struct {
	driver ScreenDriver
	done   chan struct{}
}

type fder struct {
	io.WriteCloser
}

func (fder) Fd() uintptr {
	return 0
}

func (t *tScreen) termioInit() error {
	if t.tiosp == nil {
		return ErrNoScreen
	}
	t.in = t.tiosp.driver.ReadCloser()
	t.out = fder{t.tiosp.driver.WriteCloser()}
	go func() {
		for {
			select {
			case <-t.tiosp.done:
				return
			case <-t.tiosp.driver.Resized():
				select {
				case t.sigwinch <- ResizeSignal{}:
				default:
				}
			}
		}
	}()
	return nil
}

func (t *tScreen) termioFini() {
	if t.tiosp == nil {
		return
	}
	close(t.tiosp.done)
}

func (t *tScreen) getWinSize() (int, int, error) {
	if t.tiosp == nil {
		return 0, 0, ErrNoScreen
	}
	w, h := t.tiosp.driver.Size()
	return w, h, nil
}

func (t *tScreen) Beep() error {
	return nil
}
