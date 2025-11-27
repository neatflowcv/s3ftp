package simple

import (
	"crypto/tls"
	"fmt"

	ftpserver "github.com/fclairamb/ftpserverlib"
)

var _ ftpserver.MainDriver = (*MainDriver)(nil)

const (
	defaultIdleTimeout = 300
)

type MainDriver struct {
	settings *ftpserver.Settings
}

func NewMainDriver() *MainDriver {
	return &MainDriver{
		settings: nil,
	}
}

// GetSettings returns the server settings.
func (d *MainDriver) GetSettings() (*ftpserver.Settings, error) {
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
func (d *MainDriver) ClientConnected(cc ftpserver.ClientContext) (string, error) {
	return fmt.Sprintf("220 Welcome to s3ftp server (ID: %d)", cc.ID()), nil
}

// ClientDisconnected is called when a client disconnects.
func (d *MainDriver) ClientDisconnected(cc ftpserver.ClientContext) {
	// Cleanup if needed
}

// AuthUser authenticates the user and returns a ClientDriver.
func (d *MainDriver) AuthUser( //nolint:ireturn
	cc ftpserver.ClientContext,
	user, pass string,
) (ftpserver.ClientDriver, error) {
	// Simple hardcoded authentication for testing
	if user == "test" && pass == "test" {
		return newClientDriver(), nil
	}

	return nil, ErrInvalidCredentials
}

// GetTLSConfig returns TLS configuration (not used for simple server).
func (d *MainDriver) GetTLSConfig() (*tls.Config, error) {
	return nil, ErrTLSNotConfigured
}
