// +build windows

// Copyright 2015 The TCell Authors
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
	"github.com/gdamore/tcell/terminfo"
	"io"
	"os"
)

// On Windows we don't have support for termios.  We probably could, and
// may should, in a cygwin type environment.  Its not clear how to make
// this all work nicely with both cygwin and Windows console, so we
// decline to do so here.

func (t *termiosPrivate) Init(cb CellBuffer, sigwinch chan<- os.Signal, _ *terminfo.Terminfo) (io.Reader, io.Writer, error) {
	return nil, nil, ErrNoScreen
}

func (t *termiosPrivate) Fini() {
	return
}

func (t *termiosPrivate) GetWinSize() (int, int, error) {
	return 0, 0, ErrNoScreen
}

type termiosPrivate struct{}
