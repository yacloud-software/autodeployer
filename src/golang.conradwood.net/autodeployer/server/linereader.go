package main

import (
	"fmt"
)

type LineReader struct {
	buf []byte
}

// returns a string or nil of no full line atm
func (lr *LineReader) gotBytes(a []byte, lenctr int) error {
	if len(lr.buf) > 4096 {
		return fmt.Errorf("Warning: Line > 4096 characters")
	}
	lr.buf = append(lr.buf, a[:lenctr]...)

	return nil
}
func (lr *LineReader) clearBuf() {
	lr.buf = []byte{}
}

func (lr *LineReader) getBuf() string {
	return string(lr.buf)
}

// return complete lines or ""
func (lr *LineReader) getLine() string {
	for idx, c := range lr.buf {
		if c == '\n' {
			s := string(lr.buf[:idx])
			nidx := idx + 1
			if nidx >= len(lr.buf) {
				lr.buf = []byte{}
			} else {
				lr.buf = lr.buf[nidx:]
			}
			return s
		}
	}
	return ""
}
