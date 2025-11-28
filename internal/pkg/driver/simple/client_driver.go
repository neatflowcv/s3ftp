package simple

import (
	"fmt"
	"log"
	"os"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

var _ ftpserver.ClientDriver = (*ClientDriver)(nil)

type ClientDriver struct {
	fs afero.Fs
}

func newClientDriver() *ClientDriver {
	return &ClientDriver{
		fs: afero.NewMemMapFs(),
	}
}

func (d *ClientDriver) Name() string {
	return "simple"
}

// Create creates a file.
func (d *ClientDriver) Create(name string) (afero.File, error) {
	log.Println("Create", name)

	file, err := d.fs.Create(name)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	return newFile(file), nil
}

// Mkdir creates a directory.
func (d *ClientDriver) Mkdir(name string, perm os.FileMode) error {
	err := d.fs.Mkdir(name, perm)
	if err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	return nil
}

// MkdirAll creates a directory and all parent directories.
func (d *ClientDriver) MkdirAll(path string, perm os.FileMode) error {
	err := d.fs.MkdirAll(path, perm)
	if err != nil {
		return fmt.Errorf("mkdirall: %w", err)
	}

	return nil
}

// Open opens a file.
func (d *ClientDriver) Open(name string) (afero.File, error) {
	log.Println("Open", name)

	file, err := d.fs.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return newFile(file), nil
}

// OpenFile opens a file with the specified flag and mode.
func (d *ClientDriver) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	log.Println("OpenFile", name, flag, perm)

	file, err := d.fs.OpenFile(name, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("openfile: %w", err)
	}

	return newFile(file), nil
}

// Remove removes a file or directory.
func (d *ClientDriver) Remove(name string) error {
	log.Println("Remove", name)

	err := d.fs.Remove(name)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	return nil
}

// RemoveAll removes a path and all children.
func (d *ClientDriver) RemoveAll(path string) error {
	log.Println("RemoveAll", path)

	err := d.fs.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("removeall: %w", err)
	}

	return nil
}

// Rename renames a file.
func (d *ClientDriver) Rename(oldname, newname string) error {
	log.Println("Rename", oldname, newname)

	err := d.fs.Rename(oldname, newname)
	if err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

// Stat returns file info.
func (d *ClientDriver) Stat(name string) (os.FileInfo, error) {
	log.Println("Stat", name)

	info, err := d.fs.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}

	return info, nil
}

// Chmod changes the mode of the named file.
func (d *ClientDriver) Chmod(name string, mode os.FileMode) error {
	log.Println("Chmod", name, mode)

	err := d.fs.Chmod(name, mode)
	if err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	return nil
}

// Chown changes the uid and gid of the named file.
func (d *ClientDriver) Chown(name string, uid, gid int) error {
	log.Println("Chown", name, uid, gid)

	err := d.fs.Chown(name, uid, gid)
	if err != nil {
		return fmt.Errorf("chown: %w", err)
	}

	return nil
}

// Chtimes changes the access and modification times of the named file.
func (d *ClientDriver) Chtimes(name string, atime, mtime time.Time) error {
	log.Println("Chtimes", name, atime, mtime)

	err := d.fs.Chtimes(name, atime, mtime)
	if err != nil {
		return fmt.Errorf("chtimes: %w", err)
	}

	return nil
}
