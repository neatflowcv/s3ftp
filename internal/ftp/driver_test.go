package ftp_test

import (
	"bytes"
	"io"
	"net"
	"testing"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/neatflowcv/s3ftp/internal/ftp"
)

const testFileName = "test.txt"

func TestDriverSettings(t *testing.T) {
	t.Parallel()

	driver := ftp.NewDriver()

	settings, err := driver.GetSettings()
	if err != nil {
		t.Fatalf("GetSettings failed: %v", err)
	}

	if settings.ListenAddr != "0.0.0.0:2121" {
		t.Errorf("Expected ListenAddr 0.0.0.0:2121, got %s", settings.ListenAddr)
	}
}

func TestDriverFileOperations(t *testing.T) {
	t.Parallel()

	driver := ftp.NewDriver()
	testContent := "Hello, FTP Server!"

	// Create a file
	file, err := driver.Create(testFileName)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	_, err = file.WriteString(testContent)
	if err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	// Read the file
	file, err = driver.Open(testFileName)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	var buf bytes.Buffer

	_, err = io.Copy(&buf, file)
	if err != nil {
		t.Fatalf("Read failed: %v", err)
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	if buf.String() != testContent {
		t.Errorf("Expected content %q, got %q", testContent, buf.String())
	}
}

func TestDriverStat(t *testing.T) {
	t.Parallel()

	driver := ftp.NewDriver()
	testContent := "Hello, FTP Server!"

	file, err := driver.Create(testFileName)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	_, err = file.WriteString(testContent)
	if err != nil {
		t.Fatalf("WriteString failed: %v", err)
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	info, err := driver.Stat(testFileName)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Size() != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), info.Size())
	}
}

func TestDriverRemove(t *testing.T) {
	t.Parallel()

	driver := ftp.NewDriver()

	file, err := driver.Create(testFileName)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = file.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	err = driver.Remove(testFileName)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify file is removed
	_, err = driver.Stat(testFileName)
	if err == nil {
		t.Error("File should be removed")
	}
}

func TestDriverAuthValid(t *testing.T) {
	t.Parallel()

	driver := ftp.NewDriver()
	mockCC := &mockClientContext{}

	clientDriver, err := driver.AuthUser(mockCC, "test", "test")
	if err != nil {
		t.Fatalf("AuthUser with valid credentials failed: %v", err)
	}

	if clientDriver == nil {
		t.Fatal("AuthUser should return a ClientDriver")
	}
}

func TestDriverAuthInvalid(t *testing.T) {
	t.Parallel()

	driver := ftp.NewDriver()
	mockCC := &mockClientContext{}

	_, err := driver.AuthUser(mockCC, "test", "wrong")
	if err == nil {
		t.Error("AuthUser with invalid credentials should fail")
	}
}

func TestDriverClientConnected(t *testing.T) {
	t.Parallel()

	driver := ftp.NewDriver()
	mockCC := &mockClientContext{}

	msg, err := driver.ClientConnected(mockCC)
	if err != nil {
		t.Fatalf("ClientConnected failed: %v", err)
	}

	if msg == "" {
		t.Error("ClientConnected should return a welcome message")
	}
}

// mockClientContext implements ClientContext for testing.
type mockClientContext struct{}

func (m *mockClientContext) Path() string            { return "/" }
func (m *mockClientContext) SetPath(path string)     {}
func (m *mockClientContext) SetListPath(path string) {}
func (m *mockClientContext) SetDebug(debug bool)     {}
func (m *mockClientContext) Debug() bool             { return false }
func (m *mockClientContext) ID() uint32              { return 1 }
func (m *mockClientContext) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 12345, Zone: ""}
}

func (m *mockClientContext) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 2121, Zone: ""}
}
func (m *mockClientContext) GetClientVersion() string                             { return "" }
func (m *mockClientContext) Close() error                                         { return nil }
func (m *mockClientContext) HasTLSForControl() bool                               { return false }
func (m *mockClientContext) HasTLSForTransfers() bool                             { return false }
func (m *mockClientContext) GetLastCommand() string                               { return "" }
func (m *mockClientContext) GetLastDataChannel() ftpserver.DataChannel            { return 0 }
func (m *mockClientContext) SetTLSRequirement(req ftpserver.TLSRequirement) error { return nil }
func (m *mockClientContext) SetExtra(extra any)                                   {}
func (m *mockClientContext) Extra() any                                           { return nil }
