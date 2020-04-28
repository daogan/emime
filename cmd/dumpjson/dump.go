package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/textproto"
	"os"

	"github.com/daogan/emime"
)

type KVPair struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// AttachmentResource is the body of a single MIME message part.
type AttachmentResource struct {
	AttachmentID string `json:"attachmentId,omitempty"`
	Data         []byte `json:"data,omitempty"`
	Size         int    `json:"size"`
}

// MessagePart represents the payload of the parsed email structure.
// Referenced fom Gmail API:
// https://developers.google.com/gmail/api/v1/reference/users/messages
type MessagePart struct {
	PartID   string              `json:"partId"`
	MimeType string              `json:"mimeType"`
	FileName string              `json:"filename"`
	Headers  []*KVPair           `json:"headers"`
	Body     *AttachmentResource `json:"body"`
	Parts    []*MessagePart      `json:"parts,omitempty"`
}

func map2List(mp map[string][]string, keys []string) []*KVPair {
	keyvals := make([]*KVPair, 0)
	// `mp` is a reference to header, so make a copy
	tmap := make(map[string][]string)
	for k, v := range mp {
		tmap[k] = v
	}
	for _, key := range keys {
		ck := textproto.CanonicalMIMEHeaderKey(key)
		if len(tmap[ck]) < 1 {
			continue
		}
		val := tmap[ck][0]
		tmap[ck] = tmap[ck][1:]
		keyvals = append(keyvals, &KVPair{Name: key, Value: val})
	}
	return keyvals
}

func transPart(root *emime.Part) (*MessagePart, error) {
	if root == nil {
		return nil, nil
	}
	mp := &MessagePart{
		PartID:   root.PartID,
		MimeType: root.ContentType,
		FileName: root.FileName,
		Body: &AttachmentResource{
			AttachmentID: root.ContentID,
			Data:         root.Content,
			Size:         len(root.Content),
		},
		Headers: map2List(root.Header, root.HeaderKeys),
	}

	var parts []*MessagePart
	for i := 0; i < len(root.Parts); i++ {
		if part, err := transPart(root.Parts[i]); err == nil {
			parts = append(parts, part)
		}
	}
	mp.Parts = parts

	return mp, nil
}

// ParseMessagePart parses an email into `MessagePart`.
func ParseMessagePart(r io.Reader) (*MessagePart, error) {
	part, err := emime.Parse(r)
	if err != nil {
		return nil, err
	}
	return transPart(part)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: dump <path/to/file.eml>")
		return
	}
	filename := os.Args[1]
	r, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	mp, err := ParseMessagePart(r)
	if err != nil {
		fmt.Println(err)
	}
	if mp != nil {
		b, _ := json.MarshalIndent(mp, "", "\t")
		fmt.Println(string(b))
	}
}
