package main

import (
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/neatflowcv/s3ftp/internal/pkg/driver/simple"
)

func version() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	return info.Main.Version
}

func main() {
	log.Println("version", version())

	// Create FTP driver
	driver := simple.NewMainDriver()

	// Create FTP server
	server := ftpserver.NewFtpServer(driver)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down FTP server...")

		err := server.Stop()
		if err != nil {
			log.Printf("Error stopping server: %v", err)
		}

		os.Exit(0)
	}()

	// Start server
	log.Println("Starting FTP server on :2121")
	log.Println("Test credentials: user=test, pass=test")

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("FTP server error: %v", err)
	}
}
