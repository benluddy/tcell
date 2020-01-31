// +build solaris

// Copyright 2019 The TCell Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tcell

import (
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/gdamore/tcell/terminfo"
	"golang.org/x/sys/unix"
)

type termiosPrivate struct {
	tio *unix.Termios
	in  *os.File
	out *os.File
}

func (t *termiosPrivate) Init(cb CellBuffer, sigwinch chan<- os.Signal, _ *terminfo.Terminfo) (io.Reader, io.Writer, error) {
	var e error
	var raw *unix.Termios

	if t.in, e = os.OpenFile("/dev/tty", os.O_RDONLY, 0); e != nil {
		goto failed
	}
	if t.out, e = os.OpenFile("/dev/tty", os.O_WRONLY, 0); e != nil {
		goto failed
	}

	t.tio, e = unix.IoctlGetTermios(int(t.out.Fd()), unix.TCGETS)
	if e != nil {
		goto failed
	}

	// make a local copy, to make it raw
	raw = &unix.Termios{
		Cflag: t.tio.Cflag,
		Oflag: t.tio.Oflag,
		Iflag: t.tio.Iflag,
		Lflag: t.tio.Lflag,
		Cc:    t.tio.Cc,
	}

	raw.Iflag &^= (unix.IGNBRK | unix.BRKINT | unix.PARMRK | unix.INLCR |
		unix.IGNCR | unix.ICRNL | unix.IXON)
	raw.Oflag &^= unix.OPOST
	raw.Lflag &^= (unix.ECHO | unix.ECHONL | unix.ICANON | unix.ISIG | unix.IEXTEN)
	raw.Cflag &^= (unix.CSIZE | unix.PARENB)
	raw.Cflag |= unix.CS8

	// This is setup for blocking reads.  In the past we attempted to
	// use non-blocking reads, but now a separate input loop and timer
	// copes with the problems we had on some systems (BSD/Darwin)
	// where close hung forever.
	raw.Cc[unix.VMIN] = 1
	raw.Cc[unix.VTIME] = 0

	e = unix.IoctlSetTermios(int(t.out.Fd()), unix.TCSETS, raw)
	if e != nil {
		goto failed
	}

	signal.Notify(sigwinch, syscall.SIGWINCH)

	if w, h, e := t.GetWinSize(); e == nil && w != 0 && h != 0 {
		cb.Resize(w, h)
	}

	return t.in, t.out, nil

failed:
	if t.in != nil {
		t.in.Close()
	}
	if t.out != nil {
		t.out.Close()
	}
	return nil, nil, e
}

func (t *termiosPrivate) Fini() {
	if t.out != nil && t.tio != nil {
		unix.IoctlSetTermios(int(t.out.Fd()), unix.TCSETSF, t.tio)
		t.out.Close()
	}
	if t.in != nil {
		t.in.Close()
	}
}

func (t *termiosPrivate) GetWinSize() (int, int, error) {
	wsz, err := unix.IoctlGetWinsize(int(t.out.Fd()), unix.TIOCGWINSZ)
	if err != nil {
		return -1, -1, err
	}
	return int(wsz.Col), int(wsz.Row), nil
}
