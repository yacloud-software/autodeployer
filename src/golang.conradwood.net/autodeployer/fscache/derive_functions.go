package fscache

import (
	"compress/bzip2"
	"io"
)

func register_default_functions(f *fscache) {
	f.RegisterDeriveFunction("unbzip2", Derive_unbzip2)
}

// unpbzip2 a file, registered by default as "unbzip2"
func Derive_unbzip2(r io.Reader, w io.Writer) error {
	br := bzip2.NewReader(r)
	_, err := io.Copy(w, br)
	return err
}
