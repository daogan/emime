package emime

import (
	"fmt"
	"math/rand"
	"mime"
	"strings"
	"time"

	"github.com/daogan/emime/internal/coding"
)

func lowerTrim(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func decodeHeader(input string) string {
	if !strings.Contains(input, "=?") {
		return input
	}
	dec := new(mime.WordDecoder)
	dec.CharsetReader = coding.NewCharsetReader
	header, err := dec.DecodeHeader(input)
	if err != nil {
		return input
	}
	return header
}

func fixMediaType(mtype string) string {
	segs := strings.Split(mtype, ";")
	mtype = ""
	for i, seg := range segs {
		if i == 0 {
			mtype += seg + ";"
			continue
		}
		// parameter := attribute "=" value
		pair := strings.Split(seg, "=")
		if len(pair) != 2 {
			continue
		}
		pair[0] = strings.TrimSpace(pair[0])
		pair[1] = strings.TrimSpace(pair[1])
		// invalid attribute value
		if pair[0] == "" || pair[1] == "" {
			continue
		}
		if strings.ContainsAny(pair[0], " \t") {
			continue
		}
		// contains space but not quoted
		if strings.ContainsAny(pair[1], " \t") {
			if len(pair[1]) < 2 {
				continue
			}
			if pair[1][0] != '"' || pair[1][len(pair[1])-1] != '"' {
				continue
			}
		}
		// skip special chars
		if strings.ContainsAny(seg, "()<>@,;:\"\\/[]?") {
			continue
		}
		// skip duplicate parameter name
		if strings.Contains(mtype, pair[0]+"=") {
			continue
		}
		mtype += seg + ";"
	}
	if strings.HasSuffix(mtype, ";") {
		mtype = mtype[:len(mtype)-1]
	}
	return mtype
}

func genRandomBoundary() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%028x", rand.Uint64())
}

// wrapLine wraps a long line into multiple lines of max length.
func wrapLine(max int, line string) []byte {
	output := make([]byte, 0)
	lastSpaceIdx := -1
	lastReadIdx := -1
	lineLen := 0
	for i := 0; i < len(line); i++ {
		lineLen++
		if line[i] == ' ' || line[i] == '\t' {
			lastSpaceIdx = i
		}
		if lineLen >= max {
			if lastSpaceIdx > 0 {
				output = append(output, []byte(line[lastReadIdx+1:lastSpaceIdx])...)
				output = append(output, '\r', '\n', '\t') // new line
				lastReadIdx = lastSpaceIdx
				lineLen = i - lastReadIdx // reset current line length
				lastSpaceIdx = -1         // rest space index
			}
		}
	}
	output = append(output, line[lastReadIdx+1:]...)
	return output
}
