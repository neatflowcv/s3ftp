package simple

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/afero"
)

var _ afero.File = (*File)(nil)

type File struct {
	core afero.File
}

func newFile(core afero.File) *File {
	return &File{
		core: core,
	}
}

// Close implements afero.File.
func (f *File) Close() error {
	log.Println("Close", f.core.Name())

	err := f.core.Close()
	if err != nil {
		return fmt.Errorf("close file: %w", err)
	}

	return nil
}

// Name implements afero.File.
func (f *File) Name() string {
	return f.core.Name()
}

// Read implements afero.File.
func (f *File) Read(p []byte) (int, error) {
	log.Println("Read", f.core.Name())

	n, err := f.core.Read(p)
	if err != nil {
		return 0, fmt.Errorf("read file: %w", err)
	}

	return n, nil
}

// ReadAt implements afero.File.
func (f *File) ReadAt(p []byte, off int64) (int, error) {
	log.Println("ReadAt", f.core.Name(), off)

	n, err := f.core.ReadAt(p, off)
	if err != nil {
		return 0, fmt.Errorf("readat file: %w", err)
	}

	return n, nil
}

// Readdir implements afero.File.
func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	log.Println("Readdir", f.core.Name(), count)

	infos, err := f.core.Readdir(count)
	if err != nil {
		return nil, fmt.Errorf("readdir file: %w", err)
	}

	return infos, nil
}

// Readdirnames implements afero.File.
func (f *File) Readdirnames(n int) ([]string, error) {
	log.Println("Readdirnames", f.core.Name(), n)

	names, err := f.core.Readdirnames(n)
	if err != nil {
		return nil, fmt.Errorf("readdirnames file: %w", err)
	}

	return names, nil
}

// Seek implements afero.File.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	log.Println("Seek", f.core.Name(), offset, whence)

	pos, err := f.core.Seek(offset, whence)
	if err != nil {
		return 0, fmt.Errorf("seek file: %w", err)
	}

	return pos, nil
}

// Stat implements afero.File.
func (f *File) Stat() (os.FileInfo, error) {
	log.Println("Stat", f.core.Name())

	info, err := f.core.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat file: %w", err)
	}

	return info, nil
}

// Sync implements afero.File.
func (f *File) Sync() error {
	log.Println("Sync", f.core.Name())

	err := f.core.Sync()
	if err != nil {
		return fmt.Errorf("sync file: %w", err)
	}

	return nil
}

// Truncate implements afero.File.
func (f *File) Truncate(size int64) error {
	log.Println("Truncate", f.core.Name(), size)

	err := f.core.Truncate(size)
	if err != nil {
		return fmt.Errorf("truncate file: %w", err)
	}

	return nil
}

// Write implements afero.File.
func (f *File) Write(p []byte) (int, error) {
	log.Println("Write", f.core.Name())

	n, err := f.core.Write(p)
	if err != nil {
		return 0, fmt.Errorf("write file: %w", err)
	}

	return n, nil
}

// WriteAt implements afero.File.
func (f *File) WriteAt(p []byte, off int64) (int, error) {
	log.Println("WriteAt", f.core.Name(), off)

	n, err := f.core.WriteAt(p, off)
	if err != nil {
		return 0, fmt.Errorf("writeat file: %w", err)
	}

	return n, nil
}

// WriteString implements afero.File.
func (f *File) WriteString(s string) (int, error) {
	log.Println("WriteString", f.core.Name(), s)

	n, err := f.core.WriteString(s)
	if err != nil {
		return 0, fmt.Errorf("writestring file: %w", err)
	}

	return n, nil
}
