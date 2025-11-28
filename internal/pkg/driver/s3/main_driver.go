package s3

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"maps"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	ftpserver "github.com/fclairamb/ftpserverlib"
)

const (
	defaultListenAddr  = "0.0.0.0:2121"
	defaultIdleTimeout = 5 * time.Minute
	defaultBanner      = "Welcome to s3ftp server backed by Amazon S3"
)

var (
	_ ftpserver.MainDriver = (*MainDriver)(nil)

	errRegionRequired = errors.New("aws region must be provided")
	errBucketRequired = errors.New("aws bucket must be provided")
	errUsersRequired  = errors.New("at least one user must be configured")
)

// Config groups all required inputs for the S3 FTP driver.
type Config struct {
	Region          string
	Bucket          string
	Prefix          string
	Endpoint        string
	ListenAddr      string
	IdleTimeout     time.Duration
	Banner          string
	Users           map[string]string
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// MainDriver implements ftpserver.MainDriver backed by S3.
type MainDriver struct {
	cfg       Config
	s3Client  *awss3.Client
	settings  *ftpserver.Settings
	userStore map[string]string
}

// NewMainDriver builds an S3 main driver from the provided configuration.
func NewMainDriver(cfg Config) (*MainDriver, error) {
	if cfg.Region == "" {
		return nil, errRegionRequired
	}

	if cfg.Bucket == "" {
		return nil, errBucketRequired
	}

	if len(cfg.Users) == 0 {
		return nil, errUsersRequired
	}

	opts := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}

	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, cfg.SessionToken),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}

	var s3OptFns []func(*awss3.Options)

	if cfg.Endpoint != "" {
		endpoint := cfg.Endpoint
		region := cfg.Region

		s3OptFns = append(s3OptFns, func(o *awss3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			if o.Region == "" && region != "" {
				o.Region = region
			}
		})
	}

	driver := &MainDriver{ //nolint:exhaustruct
		cfg:       normalizeConfig(cfg),
		s3Client:  awss3.NewFromConfig(awsCfg, s3OptFns...),
		userStore: make(map[string]string, len(cfg.Users)),
	}

	maps.Copy(driver.userStore, cfg.Users)

	return driver, nil
}

func normalizeConfig(cfg Config) Config {
	cfg.Prefix = strings.Trim(cfg.Prefix, "/")

	return cfg
}

// GetSettings returns FTP server settings.
func (d *MainDriver) GetSettings() (*ftpserver.Settings, error) {
	if d.settings != nil {
		return d.settings, nil
	}

	listenAddr := d.cfg.ListenAddr
	if listenAddr == "" {
		listenAddr = defaultListenAddr
	}

	banner := d.cfg.Banner
	if banner == "" {
		banner = defaultBanner
	}

	timeout := d.cfg.IdleTimeout
	if timeout == 0 {
		timeout = defaultIdleTimeout
	}

	d.settings = &ftpserver.Settings{ //nolint:exhaustruct
		ListenAddr:  listenAddr,
		IdleTimeout: int(timeout / time.Second),
		Banner:      banner,
	}

	return d.settings, nil
}

// ClientConnected sends welcome.
func (d *MainDriver) ClientConnected(clientCtx ftpserver.ClientContext) (string, error) {
	banner := d.cfg.Banner
	if banner == "" {
		banner = defaultBanner
	}

	return fmt.Sprintf("220 %s (session: %d)", banner, clientCtx.ID()), nil
}

// ClientDisconnected currently does nothing.
func (d *MainDriver) ClientDisconnected(ftpserver.ClientContext) {}

// AuthUser validates credentials and returns an S3 client driver.
func (d *MainDriver) AuthUser( //nolint:ireturn
	_ ftpserver.ClientContext,
	user, pass string,
) (ftpserver.ClientDriver, error) {
	expected, ok := d.userStore[user]
	if !ok || expected != pass {
		return nil, ErrInvalidCredentials
	}

	return newClientDriver(d.s3Client, d.cfg.Bucket, d.cfg.Prefix), nil
}

// GetTLSConfig returns nil because TLS termination is not configured here.
func (d *MainDriver) GetTLSConfig() (*tls.Config, error) {
	return nil, ErrTLSNotConfigured
}
