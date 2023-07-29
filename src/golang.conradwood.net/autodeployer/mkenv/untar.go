package mkenv

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"os"
	"sort"
	"strings"
)

type chownEntry struct {
	filename string
	uid      int
	gid      int
	chmod    int64
}

func Untar(tarfile, targetdir string) error {
	fmt.Printf("extracting tarfile %s to %s\n", tarfile, targetdir)
	f, err := os.Open(tarfile)
	if err != nil {
		return err
	}
	err = Derive_Untar(f, targetdir)
	f.Close()
	return err
}
func Derive_Untar(tarfile io.Reader, targetdir string) error {
	var err error
	var chownList []*chownEntry
	tr := tar.NewReader(tarfile)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fname := targetdir + "/" + hdr.Name
		if hdr.Typeflag == tar.TypeDir {
			// make dir
			err = os.MkdirAll(fname, fs.FileMode(hdr.Mode))
			chownList = append(chownList, &chownEntry{filename: fname, chmod: hdr.Mode, uid: hdr.Uid, gid: hdr.Gid})
		} else if hdr.Typeflag == tar.TypeSymlink {
			//			fmt.Printf("Extracting type %d, %s->%s\n", hdr.Typeflag, hdr.Name, hdr.Linkname)
			// make a symlink
			err = os.Symlink(hdr.Linkname, fname)
		} else if hdr.Typeflag == tar.TypeReg {
			// make a regular file
			tf, xerr := os.Create(fname)
			if xerr != nil {
				return err
			}
			_, err = io.Copy(tf, tr)
			tf.Close()
			chownList = append(chownList, &chownEntry{filename: fname, chmod: hdr.Mode, uid: hdr.Uid, gid: hdr.Gid})
		} else {
			return fmt.Errorf("Type %d in tarfile %s for entry %s not supported", hdr.Typeflag, tarfile, hdr.Name)
		}
		if err != nil {
			return err
		}
	}

	// sort by depth in filesystem and longest filename first, so the order of chowning grows from the leavs to the top
	sort.Slice(chownList, func(i, j int) bool {
		c1 := chownList[i]
		c2 := chownList[j]
		d1 := strings.Count(c1.filename, "/")
		d2 := strings.Count(c2.filename, "/")
		if d1 != d2 {
			return d1 > d2
		}
		return len(chownList[i].filename) > len(chownList[j].filename)
	})

	// last but not least chown files (we do this last, because once we chown it we might not be able to create a file in that directory any more
	for _, ce := range chownList {
		err = os.Chmod(ce.filename, fs.FileMode(ce.chmod))
		if err != nil {
			return err
		}
		err = os.Chown(ce.filename, ce.uid, ce.gid)
		if err != nil {
			return err
		}
	}
	return nil
}
