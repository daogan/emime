package emime

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
)

func TestBoundaryReader(t *testing.T) {
	input := "preamble\r\n--boundary\r\n1111\r\n--boundary\r\n2222\r\n--boundary\r\n"
	boundary := "boundary"
	br := bufio.NewReader(strings.NewReader(input))
	bdr := NewBoundaryReader(br, boundary)
	next, err := bdr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if !next {
		t.Fatal("expect next")
	}
	content, err := ioutil.ReadAll(bdr)
	if string(content) != "1111" {
		t.Fatalf("got: %s, want: %s", string(content), "1111")
	}
	next, err = bdr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if !next {
		t.Fatal("expect next")
	}
	content, err = ioutil.ReadAll(bdr)
	if string(content) != "2222" {
		t.Fatalf("got: %s, want: %s", string(content), "2222")
	}
	next, err = bdr.NextPart()
	if err != nil {
		t.Fatal(err)
	}
	if !next {
		t.Fatal("expect next")
	}
	content, err = ioutil.ReadAll(bdr)
	if string(content) != "" {
		t.Fatalf("got: %s, want: %s", string(content), "")
	}
}

func TestBoundaryReaderNext(t *testing.T) {
	preamble := []byte("preamble")
	boundary := "boundary"
	parts := [][]byte{
		[]byte("content of part1"),
		[]byte("content of part2"),
		[]byte("content of part3"),
	}
	buf := make([]byte, 200)
	buf = append(buf, preamble...)
	for i := 0; i < len(parts); i++ {
		bs := fmt.Sprintf("\r\n--%s\r\n", boundary)
		buf = append(buf, []byte(bs)...)
		buf = append(buf, parts[i]...)
	}
	bs := fmt.Sprintf("\r\n--%s--\r\n", boundary)
	buf = append(buf, []byte(bs)...)

	r := bufio.NewReader(bytes.NewReader(buf))
	bdr := NewBoundaryReader(r, boundary)
	for i := 0; i < len(parts); i++ {
		next, err := bdr.NextPart()
		if !next {
			t.Error("expect next")
		}
		if err != nil {
			t.Errorf("got: %v, want <nil>", err)
		}
		content, err := ioutil.ReadAll(bdr)
		if !bytes.Equal(content, parts[i]) {
			t.Fatalf("got: %s, want: %s", string(content), string(parts[i]))
		}
	}
}
