package coding

import "io"

// table of valid base64 characters of byte value range between [0, 127]
var base64Table = []int8{
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -2, -2, -1, -1, -2, -1, -1,
	-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1,
	-2, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, 62, -1, -1, -1, 63,
	52, 53, 54, 55, 56, 57, 58, 59, 60, 61, -1, -1, -1, -2, -1, -1,
	-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14,
	15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, -1, -1, -1, -1, -1,
	-1, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40,
	41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, -1, -1, -1, -1, -1,
}

type Base64Cleaner struct {
	r      io.Reader
	buffer [1024]byte
}

func NewBase64Cleaner(r io.Reader) *Base64Cleaner {
	return &Base64Cleaner{r: r}
}

// Read method for io.Reader interface.
func (bc *Base64Cleaner) Read(p []byte) (n int, err error) {
	size := len(bc.buffer)
	if size > len(p) {
		size = len(p)
	}
	buf := bc.buffer[:size]
	bn, err := bc.r.Read(buf)
	for i := 0; i < bn; i++ {
		// Strip invalid character range 0x7f ~ 0xff
		if buf[i] > 127 {
			continue
		}
		if base64Table[buf[i]&0x7f] < 0 {
			// Strip invalid characters
			continue
		}
		p[n] = buf[i]
		n++
	}
	return
}
