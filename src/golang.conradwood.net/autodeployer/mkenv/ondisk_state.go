package mkenv

import (
	"fmt"
	cd "golang.conradwood.net/apis/commondeploy"
	"golang.conradwood.net/go-easyops/utils"
	"os"
	"path/filepath"
)

type ondiskstate struct {
	file           string
	cur            *cd.OnDiskStateList
	write_required bool
}

func (o *ondiskstate) record_new_mount(mountpoint string) error {
	err := o.read_if_necessary()
	if err != nil {
		return err
	}
	oe := &cd.OnDiskMountEntry{Target: mountpoint}
	o.cur.MountEntries = append(o.cur.MountEntries, oe)
	o.write_required = true
	return nil
}
func (o *ondiskstate) remove_mountpoint(mountpoint string) error {
	err := o.read_if_necessary()
	if err != nil {
		return err
	}
	var nr []*cd.OnDiskMountEntry
	for _, me := range o.cur.MountEntries {
		if me.Target == mountpoint {
			continue
		}
		nr = append(nr, me)
	}
	o.cur.MountEntries = nr
	o.write_required = true
	return nil
}
func (o *ondiskstate) get_mount_by_mountpoint(mountpoint string) (*cd.OnDiskMountEntry, error) {
	err := o.read_if_necessary()
	if err != nil {
		return nil, err
	}
	for _, me := range o.cur.MountEntries {
		if me.Target == mountpoint {
			return me, nil
		}
	}
	return nil, nil
}
func (o *ondiskstate) get_all_mounts() ([]*cd.OnDiskMountEntry, error) {
	err := o.read_if_necessary()
	if err != nil {
		return nil, err
	}
	return o.cur.MountEntries, nil
}
func (o *ondiskstate) read_if_necessary() error {
	if o.cur != nil {
		return nil
	}
	if !utils.FileExists(o.file) {
		o.cur = &cd.OnDiskStateList{}
		return nil
	}
	b, err := utils.ReadFile(o.file)
	if err != nil {
		return err
	}
	x := &cd.OnDiskStateList{}
	err = utils.UnmarshalBytes(b, x)
	if err != nil {
		return err
	}
	o.cur = x
	return nil
}
func (o *ondiskstate) write_if_necessary() error {
	if o.cur == nil {
		return nil
	}
	if !o.write_required {
		return nil
	}
	b, err := utils.MarshalBytes(o.cur)
	if err != nil {
		return err
	}
	err = utils.WriteFile(o.file, b)
	if err != nil {
		dir := filepath.Dir(o.file)
		os.MkdirAll(dir, 0777)
		err = utils.WriteFile(o.file, b)
	}
	o.write_required = false
	return err
}

func (o *ondiskstate) ToPrettyString() string {
	err := o.read_if_necessary()
	if err != nil {
		return fmt.Sprintf("error %s", err)
	}
	s := "OnDiskStateFile \"" + o.file + "\"\n" + "  MountEntries:\n"
	for _, me := range o.cur.MountEntries {
		s = s + fmt.Sprintf("    Mountpoint: %s\n", me.Target)
	}
	return s
}
