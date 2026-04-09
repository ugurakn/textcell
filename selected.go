package textcell

import (
	"io"
	"strings"
)

type selectedText struct {
	// lines must be sorted in ASC order by selectedLnArea.y
	lines          []*selectedLnArea
	pivotX, pivotY int
	dir            byte // 'R' or 'L', '\0' if none
}

type selectedLnArea struct {
	start, end int
	y          int
}

func newSelectedText(pivotX, pivotY int) *selectedText {
	return &selectedText{
		lines:  make([]*selectedLnArea, 0, 8),
		dir:    0,
		pivotX: pivotX,
		pivotY: pivotY,
	}
}

// calcDirLnCount returns direction of selection and
// the number of lines with selected text on it.
func (s *selectedText) calcDirLnCount(cx, cy int) (byte, int) {
	if s.pivotY == cy {
		if s.pivotX == cx {
			return 0, 0
		}
		if s.pivotX < cx {
			return 'R', 1
		} else {
			return 'L', 1
		}
	}
	if v := s.pivotY - cy; v < 0 {
		return 'R', (v * -1) + 1
	} else {
		return 'L', v + 1
	}
}

// // Copy buffer

// copyBuf holds the last copied [selectedText].
type copyBuf struct {
	buf []rune
	// wi and ri are the write and read indices, respectively.
	wi, ri int
}

func newCopyBuf(size int) *copyBuf {
	return &copyBuf{
		buf: make([]rune, size),
		wi:  0,
		ri:  0,
	}
}

// debug returns copy buffer as a single-line string.
func (cb *copyBuf) debug() string {
	return strings.ReplaceAll(string(cb.buf[:cb.wi]), "\n", "<LF>")
}

// readLine returns the next unread line (without LF)
// as a reslice of copy buffer.
func (cb *copyBuf) readLine() ([]rune, error) {
	if cb.ri == cb.wi {
		return nil, io.EOF
	}
	start := cb.ri
	for i := range cb.buf[start:] {
		cb.ri++
		if cb.buf[start+i] == '\n' {
			// return without LF
			return cb.buf[start : cb.ri-1], nil
		}
	}
	return cb.buf[start:cb.ri], io.EOF
}

// writeLine appends the contents of p to copy buffer as a line.
func (cb *copyBuf) writeLine(p []rune) {
	if cb.wi != 0 {
		cb.buf[cb.wi] = '\n'
		cb.wi++
	}
	cb.wi += copy(cb.buf[cb.wi:], p)
}

func (cb *copyBuf) resetRead() {
	cb.ri = 0
}
