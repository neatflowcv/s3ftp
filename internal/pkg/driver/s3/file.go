package s3

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/spf13/afero"
)

type s3File struct {
	afero.File

	filesystem    afero.Fs
	virtualPath   string
	key           string
	bucket        string
	client        *awss3.Client
	uploadOnClose bool
}

func newS3File(
	file afero.File,
	filesystem afero.Fs,
	key, bucket string,
	client *awss3.Client,
	uploadOnClose bool,
) *s3File {
	return &s3File{
		File:          file,
		filesystem:    filesystem,
		virtualPath:   file.Name(),
		key:           key,
		bucket:        bucket,
		client:        client,
		uploadOnClose: uploadOnClose,
	}
}

func (f *s3File) Close() error {
	if f.uploadOnClose {
		data, err := afero.ReadFile(f.filesystem, f.virtualPath)
		if err != nil {
			return fmt.Errorf("read buffer for upload: %w", err)
		}

		ctx := context.Background()

		_, err = f.client.PutObject(ctx, &awss3.PutObjectInput{ //nolint:exhaustruct
			Bucket: &f.bucket,
			Key:    &f.key,
			Body:   bytes.NewReader(data),
		})
		if err != nil {
			return fmt.Errorf("upload to s3: %w", err)
		}
	}

	err := f.File.Close()
	if err != nil {
		return fmt.Errorf("close file buffer: %w", err)
	}

	return nil
}

type s3FileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	dir     bool
}

func (i *s3FileInfo) Name() string       { return i.name }
func (i *s3FileInfo) Size() int64        { return i.size }
func (i *s3FileInfo) Mode() os.FileMode  { return i.mode }
func (i *s3FileInfo) ModTime() time.Time { return i.modTime }
func (i *s3FileInfo) IsDir() bool        { return i.dir }
func (i *s3FileInfo) Sys() any           { return nil }
