package emime

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"io"
	"io/ioutil"
	"mime/quotedprintable"
	"net/textproto"
	"strings"

	"github.com/daogan/emime/internal/coding"
)

// Encode encodes the Part tree back to plain text.
func (p *Part) Encode(writer io.Writer) error {
	b := bufio.NewWriter(writer)
	cte := p.setupPart()
	p.encodeHeader(b)

	if len(p.Content) > 0 {
		b.Write(crnl)
		if err := p.encodeContent(b, cte); err != nil {
			return err
		}
	}
	if len(p.Parts) == 0 {
		return b.Flush()
	}

	// Encode `message/rfc822`.
	if p.ContentType == ctRFC822 {
		if len(p.Parts) > 0 {
			cte = lowerTrim(p.Header.Get(hContentEncoding))
			if cte != cteBase64 {
				b.Write(crnl)
			}
			if err := p.Parts[0].Encode(b); err != nil {
				return err
			}
		}
		// `message/rfc822` part has only one child, just return.
		return b.Flush()
	}

	// Encode children.
	endMarker := []byte("\r\n--" + p.Boundary + "--")
	boundary := endMarker[:len(endMarker)-2]
	for i := 0; i < len(p.Parts); i++ {
		b.Write(boundary)
		b.Write(crnl)
		if err := p.Parts[i].Encode(b); err != nil {
			return err
		}
	}
	b.Write(endMarker)
	b.Write(crnl)
	return b.Flush()
}

func (p *Part) encodeHeader(b *bufio.Writer) {
	// make a copy to avoid modifying original header map
	tHeader := make(textproto.MIMEHeader)
	for k, v := range p.Header {
		tHeader[k] = v
	}
	for _, k := range p.HeaderKeys {
		ck := textproto.CanonicalMIMEHeaderKey(k)
		// duplicate Content-Type header may be deleted
		if len(tHeader[ck]) < 1 {
			continue
		}
		val := tHeader[ck][0]
		tHeader[ck] = tHeader[ck][1:]
		line := k + ":_" + val + "\r\n"
		wb := wrapLine(76, line)
		wb[len(k)+1] = ' '
		b.Write(wb)
	}
}

func (p *Part) encodeContent(b *bufio.Writer, cte string) (err error) {
	content := p.Content
	// encode text with stated charset
	if strings.HasPrefix(p.ContentType, "text") {
		input := bytes.NewReader(p.Content)
		if r, err := coding.NewCharsetEncoder(p.Charset, input); err == nil {
			enc, err := ioutil.ReadAll(r)
			if err == nil {
				content = enc
			}
		}
	}

	switch lowerTrim(cte) {
	case cteBase64:
		enc := base64.StdEncoding
		text := make([]byte, enc.EncodedLen(len(content)))
		base64.StdEncoding.Encode(text, content)
		// Wrap lines.
		lineLen := 76
		for len(text) > 0 {
			if lineLen > len(text) {
				lineLen = len(text)
			}
			if _, err = b.Write(text[:lineLen]); err != nil {
				return err
			}
			b.Write(crnl)
			text = text[lineLen:]
		}
	case cteQuotedPrintable:
		qp := quotedprintable.NewWriter(b)
		if _, err = qp.Write(content); err != nil {
			return err
		}
		err = qp.Close()
	default:
		_, err = b.Write(content)
	}
	return err
}

func (p *Part) setupPart() (cte string) {
	if p.Header == nil {
		p.Header = make(textproto.MIMEHeader)
	}
	// Restore Content-Transfer-Encoding for base64 rfc822 attachment
	if len(p.Header) == 0 && p.Parent != nil && p.Parent.ContentType == ctRFC822 {
		cte = p.Parent.Header.Get(hContentEncoding)
	} else {
		cte = p.Header.Get(hContentEncoding)
	}
	cte = lowerTrim(cte)
	switch cte {
	case cte7Bit:
	case cteBase64:
	case cteQuotedPrintable:
	case cte8Bit, cteBinary:
		cte = cte7Bit
	default:
		// RFC 2045: 7bit is assumed if CTE header not present.
		cte = cte7Bit
	}
	if p.ContentType != ctRFC822 && len(p.Parts) > 0 && p.Boundary == "" {
		p.Boundary = genRandomBoundary()
	}
	return cte
}
