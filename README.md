# emime

emime is an email / [MIME](https://datatracker.ietf.org/doc/html/rfc2045) parser written in `Go`.

Multipurpose Internet Mail Extensions (MIME) is a standard that extends the format of email messages to support text in character sets other than ASCII, as well as attachments of audio, video, images, and application programs. Message bodies may consist of multiple parts, and header information may be specified in non-ASCII character sets. 


## Example

```go
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/daogan/emime"
)

func main() {
	r, _ := os.Open("./sample.eml")
	part, err := emime.Parse(r)
	if err != nil {
		fmt.Println("Parse error:", err)
		return
	}

	buf := &bytes.Buffer{}
	err = part.Encode(buf)
	if err != nil {
		fmt.Println("Encode error:", err)
		return
	}
	ioutil.WriteFile("./output.eml", buf.Bytes(), 0644)
}
```
