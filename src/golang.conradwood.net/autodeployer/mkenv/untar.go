package mkenv

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
)

func Untar(tarfile, targetdir string) error {
	fmt.Printf("extracting tarfile %s to %s", tarfile, targetdir)
	f, err := os.Open(tarfile)
	if err != nil {
		return err
	}
	defer f.Close()
	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fmt.Printf("Extracting type %d, %s\n", hdr.Typeflag, hdr.Name)
	}
	return nil
}
