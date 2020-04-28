package emime

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/pkg/errors"
)

const (
	peakSize = 4096
)

// BoundaryReader struct
type BoundaryReader struct {
	br        *bufio.Reader
	partsRead int
	finished  bool

	nl               []byte // "\r\n" or "\n" (set after seeing first boundary line)
	nlDashBoundary   []byte // nl + "--boundary"
	dashBoundaryDash []byte // "--boundary--"
	dashBoundary     []byte // "--boundary"
}

// NewBoundaryReader returns a new BoundaryReader.
func NewBoundaryReader(r *bufio.Reader, boundary string) *BoundaryReader {
	b := []byte("\r\n--" + boundary + "--")
	return &BoundaryReader{
		br:               r,
		nl:               b[:2],
		nlDashBoundary:   b[:len(b)-2],
		dashBoundaryDash: b[2:],
		dashBoundary:     b[2 : len(b)-2],
	}
}

// Read reads from buffer until the next boundary.
func (b *BoundaryReader) Read(dest []byte) (int, error) {
	peek, err := b.br.Peek(peakSize)
	if err != nil && err != io.EOF && err != bufio.ErrBufferFull {
		return 0, errors.WithStack(err)
	}
	var nRead int
	idx := findBoundary(peek, b.nlDashBoundary)
	if idx == 0 {
		// boundary found in begining of buffer
		return 0, io.EOF
	} else if idx > 0 {
		// boundary found in middle of buffer
		nRead = idx
	} else {
		// boundary not found in buffer,
		// move back a safe distance to avoid reading partial boundary
		nRead = len(peek) - len(b.nlDashBoundary)
		if nRead <= 0 {
			if err == io.EOF {
				// read all remaining data on last call
				if len(peek) > 0 {
					nRead = len(peek)
				} else {
					return 0, io.EOF
				}
			}
		}
	}
	if nRead > 0 {
		if nRead > len(dest) {
			nRead = len(dest)
		}
		n, err := b.br.Read(dest[:nRead])
		return n, err
	}
	return 0, nil
}

// NextPart consumes the next part boundary.
func (b *BoundaryReader) NextPart() (bool, error) {
	if b.finished {
		return false, nil
	}
	for {
		line, err := b.br.ReadSlice('\n')
		if err != nil && err != io.EOF {
			return false, errors.Errorf("boundary: NextPart: %v", err)
		}

		if b.isDelimiter(line) {
			b.partsRead++
			return true, nil
		}
		if b.isTerminator(line) {
			b.finished = true
			return false, io.EOF
		}
		if err == io.EOF {
			return false, io.EOF
		}
		if bytes.Equal(line, b.nl) {
			// skip blank lines
			continue
		}
		if b.partsRead == 0 {
			// burn off preamble before the first part if there's any
			continue
		}
		b.finished = true
		return false, fmt.Errorf("boundary: unexpected line in NextPart(): %q", line)
	}
}

// matches `^--boundary[ \t]*[\r\n]$`
func (b *BoundaryReader) isDelimiter(line []byte) bool {
	if !bytes.HasPrefix(line, b.dashBoundary) {
		return false
	}
	rest := line[len(b.dashBoundary):]
	rest = bytes.TrimLeft(rest, " \t")
	// switch to ending of \n instead of \r\n on first boundary if is the case
	if b.partsRead == 0 && len(rest) == 1 && rest[0] == '\n' {
		b.nl = b.nl[1:]
		b.nlDashBoundary = b.nlDashBoundary[1:]
	}
	// more tolerant ending
	if len(rest) > 0 {
		if rest[0] == '\r' || rest[0] == '\n' {
			return true
		}
	}
	return bytes.Equal(rest, b.nl)
}

// matches `^--boundary--*`,
// more strict should matche `^--boundary--[ \t]*(\r\n)?$`
func (b *BoundaryReader) isTerminator(line []byte) bool {
	if !bytes.HasPrefix(line, b.dashBoundaryDash) {
		return false
	}
	// matches `^--boundary--[ \t]*(\r\n)?$`
	// rest := line[len(b.dashBoundaryDash):]
	// rest = bytes.TrimLeft(rest, " \t")
	// return len(rest) == 0 || bytes.Equal(rest, b.nl)

	return true // more tolerant ending
}

// matches --boundary$, --boundary--, --boundary\r\n, --boundary[ \t]
// but not --boundary-2 or --boundary2
func findBoundary(buf []byte, nlDashBoundary []byte) int {
	idx := bytes.Index(buf, nlDashBoundary)
	if idx >= 0 {
		rest := buf[idx+len(nlDashBoundary):]
		// matches --boundary$
		if len(rest) == 0 {
			// need more data to verify EOF
			return idx
		}
		c := rest[0]
		// matches --boundary\r\n, --boundary[ \t]
		if c == ' ' || c == '\t' || c == '\r' || c == '\n' {
			return idx
		}
		// matches --boundary--
		if bytes.HasPrefix(rest, []byte("--")) {
			return idx
		}
		// boundary found but not match, fast forward
		return idx + len(nlDashBoundary)
	}
	return -1
}
