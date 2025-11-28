package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/joho/godotenv"
	"github.com/neatflowcv/s3ftp/internal/pkg/driver/s3"
)

func version() string {
	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}

	return buildInfo.Main.Version
}

func main() {
	log.Println("version", version())

	loadErr := godotenv.Load(".env")
	if loadErr != nil {
		log.Fatalf("Error loading .env file: %v", loadErr)
	}

	cfg, err := buildConfigFromEnv()
	if err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	driver, err := s3.NewMainDriver(cfg)
	if err != nil {
		log.Fatalf("Error creating FTP driver: %v", err)
	}

	// Create FTP server
	server := ftpserver.NewFtpServer(driver)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down FTP server...")

		stopErr := server.Stop()
		if stopErr != nil {
			log.Printf("Error stopping server: %v", stopErr)
		}

		os.Exit(0)
	}()

	// Start server
	log.Println("Starting FTP server on :2121")
	log.Println("Test credentials: user=test, pass=test")

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("FTP server error: %v", err)
	}
}

const (
	s3UsersDelimiter     = ","
	s3UserCredentialSize = 2
)

var (
	errUsersEmpty       = errors.New("S3_USERS must contain at least one user")
	errInvalidUserEntry = errors.New("invalid S3_USERS entry")
	errEnvMissing       = errors.New("environment variable is required")
	errEnvEmpty         = errors.New("environment variable cannot be empty")
)

func buildConfigFromEnv() (s3.Config, error) {
	requiredKeys := []string{
		"S3_REGION",
		"S3_BUCKET",
		"S3_LISTEN_ADDR",
		"S3_IDLE_TIMEOUT",
		"S3_BANNER",
		"S3_USERS",
		"S3_ACCESS_KEY_ID",
		"S3_SECRET_ACCESS_KEY",
		"S3_ENDPOINT",
	}

	envValues, err := loadRequiredEnv(requiredKeys...)
	if err != nil {
		return s3.Config{}, err
	}

	idleTimeout, err := parseIdleTimeout(envValues["S3_IDLE_TIMEOUT"])
	if err != nil {
		return s3.Config{}, err
	}

	users, err := parseUsers(envValues["S3_USERS"])
	if err != nil {
		return s3.Config{}, err
	}

	return s3.Config{
		Region:          envValues["S3_REGION"],
		Bucket:          envValues["S3_BUCKET"],
		Prefix:          os.Getenv("S3_PREFIX"),
		ListenAddr:      envValues["S3_LISTEN_ADDR"],
		IdleTimeout:     idleTimeout,
		Banner:          envValues["S3_BANNER"],
		Users:           users,
		AccessKeyID:     envValues["S3_ACCESS_KEY_ID"],
		SecretAccessKey: envValues["S3_SECRET_ACCESS_KEY"],
		SessionToken:    os.Getenv("S3_SESSION_TOKEN"),
		Endpoint:        envValues["S3_ENDPOINT"],
	}, nil
}

func loadRequiredEnv(keys ...string) (map[string]string, error) {
	values := make(map[string]string, len(keys))

	for _, key := range keys {
		value, err := requiredEnv(key)
		if err != nil {
			return nil, err
		}

		values[key] = value
	}

	return values, nil
}

func parseIdleTimeout(rawTimeout string) (time.Duration, error) {
	timeoutSeconds, err := strconv.Atoi(rawTimeout)
	if err != nil {
		return 0, fmt.Errorf("S3_IDLE_TIMEOUT must be a number: %w", err)
	}

	return time.Duration(timeoutSeconds) * time.Second, nil
}

func parseUsers(raw string) (map[string]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errUsersEmpty
	}

	users := make(map[string]string)

	pairs := strings.SplitSeq(raw, s3UsersDelimiter)
	for pairValue := range pairs {
		pair := strings.TrimSpace(pairValue)
		if pair == "" {
			continue
		}

		credentials := strings.SplitN(pair, ":", s3UserCredentialSize)
		if len(credentials) != s3UserCredentialSize {
			return nil, fmt.Errorf("%w: %s", errInvalidUserEntry, pair)
		}

		username := strings.TrimSpace(credentials[0])

		password := strings.TrimSpace(credentials[1])
		if username == "" || password == "" {
			return nil, fmt.Errorf("%w: %s", errInvalidUserEntry, pair)
		}

		users[username] = password
	}

	if len(users) == 0 {
		return nil, errUsersEmpty
	}

	return users, nil
}

func requiredEnv(key string) (string, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", fmt.Errorf("%w: %s", errEnvMissing, key)
	}

	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("%w: %s", errEnvEmpty, key)
	}

	return value, nil
}
