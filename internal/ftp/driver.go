package ftp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"os"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

const (
	defaultIdleTimeout = 300
)

var (
	// ErrInvalidCredentials is returned when authentication fails.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrTLSNotConfigured is returned when TLS is requested but not configured.
	ErrTLSNotConfigured = errors.New("TLS not configured")
)

var (
	_ ftpserver.MainDriver = (*Driver)(nil)
	_ afero.Fs             = (*Driver)(nil)
)

// Driver implements MainDriver and ClientDriver interfaces.
type Driver struct {
	settings *ftpserver.Settings
	fs       afero.Fs
}

// NewDriver creates a new FTP driver instance.
func NewDriver() *Driver {
	return &Driver{
		fs:       afero.NewMemMapFs(),
		settings: nil,
	}
}

// GetSettings returns the server settings.
func (d *Driver) GetSettings() (*ftpserver.Settings, error) {
	if d.settings == nil {
		d.settings = &ftpserver.Settings{
			ListenAddr:               "0.0.0.0:2121",
			IdleTimeout:              defaultIdleTimeout,
			Banner:                   "Welcome to s3ftp server",
			Listener:                 nil,
			PublicHost:               "",
			PassiveTransferPortRange: nil,
			PublicIPResolver:         nil,
			ConnectionTimeout:        0,
			ActiveTransferPortNon20:  false,
			DisableMLSD:              false,
			DisableMLST:              false,
			DisableMFMT:              false,
			TLSRequired:              0,
			DisableLISTArgs:          false,
			DisableSite:              false,
			DisableActiveMode:        false,
			EnableHASH:               false,
			DisableSTAT:              false,
			DisableSYST:              false,
			EnableCOMB:               false,
			DefaultTransferType:      0,
			ActiveConnectionsCheck:   0,
			PasvConnectionsCheck:     0,
		}
	}

	return d.settings, nil
}

// ClientConnected is called when a client connects.
func (d *Driver) ClientConnected(cc ftpserver.ClientContext) (string, error) {
	return fmt.Sprintf("220 Welcome to s3ftp server (ID: %d)", cc.ID()), nil
}

// ClientDisconnected is called when a client disconnects.
func (d *Driver) ClientDisconnected(cc ftpserver.ClientContext) {
	// Cleanup if needed
}

// AuthUser authenticates the user and returns a ClientDriver.
func (d *Driver) AuthUser( //nolint:ireturn
	cc ftpserver.ClientContext,
	user, pass string,
) (ftpserver.ClientDriver, error) {
	// Simple hardcoded authentication for testing
	if user == "test" && pass == "test" {
		return d, nil
	}

	return nil, ErrInvalidCredentials
}

// GetTLSConfig returns TLS configuration (not used for simple server).
func (d *Driver) GetTLSConfig() (*tls.Config, error) {
	return nil, ErrTLSNotConfigured
}

// ClientDriver interface implementation (afero.Fs)
// The Driver struct already embeds afero.Fs methods through composition

// Name returns the name of the filesystem.
func (d *Driver) Name() string {
	return d.fs.Name()
}

// Create creates a file.
func (d *Driver) Create(name string) (afero.File, error) { //nolint:ireturn
	file, err := d.fs.Create(name)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}

	return file, nil
}

// Mkdir creates a directory.
func (d *Driver) Mkdir(name string, perm os.FileMode) error {
	err := d.fs.Mkdir(name, perm)
	if err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}

	return nil
}

// MkdirAll creates a directory and all parent directories.
func (d *Driver) MkdirAll(path string, perm os.FileMode) error {
	err := d.fs.MkdirAll(path, perm)
	if err != nil {
		return fmt.Errorf("mkdirall: %w", err)
	}

	return nil
}

// Open opens a file.
func (d *Driver) Open(name string) (afero.File, error) { //nolint:ireturn
	file, err := d.fs.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}

	return file, nil
}

// OpenFile opens a file with the specified flag and mode.
func (d *Driver) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) { //nolint:ireturn
	file, err := d.fs.OpenFile(name, flag, perm)
	if err != nil {
		return nil, fmt.Errorf("openfile: %w", err)
	}

	return file, nil
}

// Remove removes a file or directory.
func (d *Driver) Remove(name string) error {
	err := d.fs.Remove(name)
	if err != nil {
		return fmt.Errorf("remove: %w", err)
	}

	return nil
}

// RemoveAll removes a path and all children.
func (d *Driver) RemoveAll(path string) error {
	err := d.fs.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("removeall: %w", err)
	}

	return nil
}

// Rename renames a file.
func (d *Driver) Rename(oldname, newname string) error {
	err := d.fs.Rename(oldname, newname)
	if err != nil {
		return fmt.Errorf("rename: %w", err)
	}

	return nil
}

// Stat returns file info.
func (d *Driver) Stat(name string) (os.FileInfo, error) {
	info, err := d.fs.Stat(name)
	if err != nil {
		return nil, fmt.Errorf("stat: %w", err)
	}

	return info, nil
}

// Chmod changes the mode of the named file.
func (d *Driver) Chmod(name string, mode os.FileMode) error {
	err := d.fs.Chmod(name, mode)
	if err != nil {
		return fmt.Errorf("chmod: %w", err)
	}

	return nil
}

// Chown changes the uid and gid of the named file.
func (d *Driver) Chown(name string, uid, gid int) error {
	err := d.fs.Chown(name, uid, gid)
	if err != nil {
		return fmt.Errorf("chown: %w", err)
	}

	return nil
}

// Chtimes changes the access and modification times of the named file.
func (d *Driver) Chtimes(name string, atime, mtime time.Time) error {
	err := d.fs.Chtimes(name, atime, mtime)
	if err != nil {
		return fmt.Errorf("chtimes: %w", err)
	}

	return nil
}
