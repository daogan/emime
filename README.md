# emime

emime is an email / [MIME](https://tools.ietf.org/html/rfc2045) parser written in `Go`.


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
