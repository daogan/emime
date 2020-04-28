package emime

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/quotedprintable"
	"net/textproto"
	"strconv"
	"strings"

	"github.com/daogan/emime/internal/coding"
)

// Part is the node of the parsed tree.
type Part struct {
	PartID string
	Header textproto.MIMEHeader

	Boundary    string
	ContentID   string
	ContentType string
	Disposition string
	FileName    string
	Charset     string

	Content []byte

	HeaderKeys []string

	Parent *Part
	Parts  []*Part
}

func (p *Part) setupHeaders(r *bufio.Reader, defaultContentType string) error {
	header, err := readHeader(r, p)
	if err != nil {
		return err
	}
	p.Header = header

	ctype := header.Get(hContentType)
	if ctype == "" {
		ctype = defaultContentType
	}
	mtype, tparams, err := mime.ParseMediaType(ctype)
	if err != nil {
		mtype, tparams, err = mime.ParseMediaType(fixMediaType(ctype))
		if err != nil {
			return fmt.Errorf("media type, err: %v, Content-Type: %q", err, ctype)
		}
	}
	p.ContentType = mtype
	p.Boundary = tparams[hpBoundary]
	p.ContentID = header.Get(hContentID)
	if p.Charset == "" {
		p.Charset = tparams[hpCharset]
	}

	cdisp := header.Get(hContentDisposition)
	disposition, dparams, err := mime.ParseMediaType(cdisp)
	if err == nil {
		p.Disposition = disposition
		p.FileName = decodeHeader(dparams[hpFileName])
	}
	if p.FileName == "" && tparams[hpName] != "" {
		p.FileName = decodeHeader(tparams[hpName])
	}

	return nil
}

func (p *Part) decodeContent(r io.Reader, encoding string) error {
	if encoding == "" {
		encoding = p.Header.Get(hContentEncoding)
	}
	contentReader := r
	switch lowerTrim(encoding) {
	case cteQuotedPrintable:
		contentReader = quotedprintable.NewReader(contentReader)
	case cteBase64:
		b64cleaner := coding.NewBase64Cleaner(contentReader)
		contentReader = base64.NewDecoder(base64.RawStdEncoding, b64cleaner)
	case cte8Bit, cte7Bit, cteBinary, "":
		// No decoding required.
	default:
		// Unknown encoding.
	}

	// BUG: Bug in official "mime/quotedprintable" lib:
	// quotedprintable reader may return `bufio.ErrBufferFull`
	// before exhausting the reader.
	content, err := ioutil.ReadAll(contentReader)
	if err != nil {
		// If content is corrupt, keep the partial decoded content,
		// silently ignore early errors and continue parsing.
		// return err
	}
	p.Content = content

	return nil
}

// AddChild adds a child node into the part tree.
func (p *Part) AddChild(child *Part) {
	if child != nil {
		p.Parts = append(p.Parts, child)
		child.Parent = p
	}
}

func parseMultiPart(parent *Part, r *bufio.Reader) error {
	bdr := NewBoundaryReader(r, parent.Boundary)
	for partIdx := 0; true; partIdx++ {
		next, err := bdr.NextPart()
		if err != nil && err != io.EOF {
			return err
		}
		if !next {
			break
		}

		p := &Part{}
		if parent.PartID == "" {
			p.PartID = strconv.Itoa(partIdx)
		} else {
			p.PartID = parent.PartID + "." + strconv.Itoa(partIdx)
		}
		br := bufio.NewReader(bdr)
		err = p.setupHeaders(br, defaultContentType)
		if err != nil {
			return err
		}

		if strings.HasPrefix(p.ContentType, ctMessagePrefix) {
			if p.ContentType == ctRFC822 {
				cte := lowerTrim(p.Header.Get(hContentEncoding))
				// `message/rfc822` base64 attachment is treated as a new child part.
				if cte == cteBase64 {
					pp := &Part{
						PartID:      p.PartID + ".0",
						ContentType: ctTextPlain,
					}
					p.AddChild(pp)
					if err := pp.decodeContent(br, cteBase64); err != nil {
						return err
					}
				} else {
					if part, err := parse(p, br); err == nil {
						p.AddChild(part)
					}
				}
			} else {
				// TODO: handle other "message/xxx" types
			}
		}

		parent.AddChild(p)

		if p.Boundary == "" {
			err = p.decodeContent(br, "")
			if err != nil {
				return err
			}
		} else {
			err = parseMultiPart(p, br)
			if err != nil {
				return err
			}
		}
	}
	// burn off epilogues if there are any
	_, _ = io.Copy(ioutil.Discard, r)

	return nil
}

func parse(parent *Part, r io.Reader) (*Part, error) {
	partID := ""
	if parent != nil {
		partID = parent.PartID + ".0"
	}
	root := &Part{PartID: partID}
	br := bufio.NewReader(r)
	err := root.setupHeaders(br, defaultContentType)
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(root.ContentType, ctMultipartPrefix) {
		err = parseMultiPart(root, br)
		if err != nil {
			return nil, err
		}
	} else {
		err = root.decodeContent(br, "")
		if err != nil {
			return nil, err
		}
	}
	return root, nil
}

// Parse parses an email into `Part` tree.
func Parse(r io.Reader) (*Part, error) {
	return parse(nil, r)
}
