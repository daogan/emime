package emime

import "strings"

type Attachment struct {
	AttachmentID string // AttachmentID ID for the attachment.
	ContentType  string // ContentType header without parameters.
	Disposition  string // Content-Disposition header without parameters.
	FileName     string // The file-name from disposition or type header.
	Data         []byte
	Size         int
}

func isAttachment(part *Part) bool {
	if part == nil {
		return false
	}
	if part.Disposition == cdAttachment || part.ContentType == ctAppOctetStream {
		return true
	}
	// if part.Disposition == cdInline {
	//     // inline images/ and application/{pdf/octet-stream} are treated as attachments
	//     if strings.HasPrefix(part.ContentType, "image/") ||
	//         strings.HasPrefix(part.ContentType, "application/") {
	//         return true
	//     }
	// }
	return false
}

func part2Attachment(part *Part) *Attachment {
	if part == nil {
		return nil
	}
	attachment := &Attachment{
		AttachmentID: part.ContentID,
		ContentType:  part.ContentType,
		Disposition:  part.Disposition,
		FileName:     part.FileName,
	}
	if part.ContentType == ctRFC822 {
		// TODO: if `message/rfc822` attachment is not base64 encoded,
		// this will return empty data. Should encode the part sub tree instead.
		if len(part.Parts) > 0 {
			attachment.Data = part.Parts[0].Content
			attachment.Size = len(part.Parts[0].Content)
		}
	} else {
		attachment.Data = part.Content
		attachment.Size = len(part.Content)
	}
	return attachment
}

func appendAttachments(root *Part, attachments *[]*Attachment) {
	if root == nil {
		return
	}
	if isAttachment(root) {
		attachment := part2Attachment(root)
		if attachment != nil {
			*attachments = append(*attachments, attachment)
		}
	}
	for _, part := range root.Parts {
		appendAttachments(part, attachments)
	}
}

// GetAttachments returns all attachments in root.
func GetAttachments(root *Part) []*Attachment {
	var attachments []*Attachment
	appendAttachments(root, &attachments)
	return attachments
}
