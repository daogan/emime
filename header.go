package emime

import (
	"bufio"
	"bytes"
	"io"
	"net/textproto"

	"github.com/pkg/errors"
)

const (
	// Header names
	hContentDisposition = "Content-Disposition"
	hContentEncoding    = "Content-Transfer-Encoding"
	hContentID          = "Content-ID"
	hContentType        = "Content-Type"

	// Content-Type
	ctMultipartPrefix = "multipart/"
	ctMessagePrefix   = "message/"
	ctTextPlain       = "text/plain"
	ctTextHTML        = "text/html"
	ctRFC822          = "message/rfc822"
	ctAppOctetStream  = "application/octet-stream"

	// Content-Transfter-Encoding
	cte7Bit            = "7bit"
	cte8Bit            = "8bit"
	cteBase64          = "base64"
	cteBinary          = "binary"
	cteQuotedPrintable = "quoted-printable"

	// Content-Desposition
	cdAttachment = "attachment"
	cdInline     = "inline"

	// header parameters
	hpFileName = "filename"
	hpName     = "name"
	hpBoundary = "boundary"
	hpCharset  = "charset"

	// rfc2045: Content-Type Defaults
	defaultContentType = `text/plain; charset=us-ascii`
)

var crnl = []byte{'\r', '\n'}

func readHeader(r *bufio.Reader, p *Part) (textproto.MIMEHeader, error) {
	buf := &bytes.Buffer{}
	tp := textproto.NewReader(r)
	firstHeader := true
	for {
		line, err := tp.ReadLineBytes()
		if err != nil {
			if err == io.EOF {
				buf.Write(crnl)
				break
			}
			return nil, errors.WithStack(err)
		}
		spaceIdx := bytes.IndexAny(line, " \t\r\n")
		// start with space, continuation
		if spaceIdx == 0 {
			buf.WriteByte(' ')
			buf.Write(textproto.TrimBytes(line))
			continue
		}
		colonIdx := bytes.IndexByte(line, ':')
		if colonIdx == 0 {
			// illegal line, skip
			continue
		}
		// contains colon, new header entry
		if colonIdx > 0 {
			if !firstHeader {
				buf.Write(crnl)
			}
			buf.Write(textproto.TrimBytes(line))
			firstHeader = false
			// Keep header keys in order
			headerKey := string(textproto.TrimBytes(line[:colonIdx]))
			p.HeaderKeys = append(p.HeaderKeys, headerKey)
		} else {
			if len(line) > 0 {
				// illegal line, treat as unintented continuation
				buf.WriteByte(' ')
				buf.Write(textproto.TrimBytes(line))
				continue
			} else {
				// empty line, end of header
				buf.Write(crnl)
				break
			}
		}
	}
	buf.Write(crnl) // end of header marker
	tp = textproto.NewReader(bufio.NewReader(buf))
	hdr, err := tp.ReadMIMEHeader()
	return hdr, err
}
