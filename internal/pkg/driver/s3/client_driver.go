package s3

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awss3 "github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
	ftpserver "github.com/fclairamb/ftpserverlib"
	"github.com/spf13/afero"
)

const (
	memFileName       = "payload"
	defaultDriverName = "s3"
	dirPermission     = 0o755
)

var (
	_ ftpserver.ClientDriver                  = (*clientDriver)(nil)
	_ ftpserver.ClientDriverExtensionFileList = (*clientDriver)(nil)
)

type clientDriver struct {
	s3      *awss3.Client
	bucket  string
	prefix  string
	builder func() afero.Fs
}

var errUnexpectedFileType = errors.New("unexpected file type for population")

func newClientDriver(s3Client *awss3.Client, bucket, prefix string) *clientDriver {
	return &clientDriver{
		s3:      s3Client,
		bucket:  bucket,
		prefix:  prefix,
		builder: afero.NewMemMapFs,
	}
}

func (d *clientDriver) Name() string {
	return defaultDriverName
}

func (d *clientDriver) Create(name string) (afero.File, error) { //nolint:ireturn
	return d.newWritableFile(name)
}

func (d *clientDriver) Mkdir(name string, perm os.FileMode) error {
	return d.ensureDirMarker(name)
}

func (d *clientDriver) MkdirAll(path string, perm os.FileMode) error {
	return d.ensureDirMarker(path)
}

func (d *clientDriver) Open(name string) (afero.File, error) { //nolint:ireturn
	key := d.objectKey(name)
	if key == "" {
		return nil, fs.ErrNotExist
	}

	data, err := d.downloadObject(key)
	if err != nil {
		return nil, err
	}

	mem := d.builder()

	file, err := mem.Create(memFileName)
	if err != nil {
		return nil, fmt.Errorf("create buffer: %w", err)
	}

	if len(data) > 0 {
		var writeErr error

		_, writeErr = file.Write(data)
		if writeErr != nil {
			return nil, fmt.Errorf("fill buffer: %w", writeErr)
		}

		var seekErr error

		_, seekErr = file.Seek(0, io.SeekStart)
		if seekErr != nil {
			return nil, fmt.Errorf("reset buffer: %w", seekErr)
		}
	}

	return newS3File(file, mem, key, d.bucket, d.s3, false), nil
}

func (d *clientDriver) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) { //nolint:ireturn
	writeMode := flag&(os.O_WRONLY|os.O_RDWR|os.O_APPEND|os.O_CREATE|os.O_TRUNC) != 0
	if !writeMode {
		return d.Open(name)
	}

	file, err := d.newWritableFile(name)
	if err != nil {
		return nil, err
	}

	appendMode := flag&os.O_APPEND != 0
	if !appendMode {
		return file, nil
	}

	key := d.objectKey(name)
	if key == "" {
		return file, nil
	}

	populateErr := d.populateWritableFile(file, key)
	if populateErr != nil && !errors.Is(populateErr, fs.ErrNotExist) {
		return nil, populateErr
	}

	return file, nil
}

func (d *clientDriver) Remove(name string) error {
	key := d.objectKey(name)
	if key == "" {
		return fs.ErrInvalid
	}

	ctx := context.Background()

	_, err := d.s3.DeleteObject(ctx, &awss3.DeleteObjectInput{ //nolint:exhaustruct
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %q: %w", name, err)
	}

	return nil
}

func (d *clientDriver) RemoveAll(path string) error {
	key := d.objectKey(path)
	if key == "" {
		return fs.ErrInvalid
	}

	return d.deleteByPrefix(key)
}

func (d *clientDriver) Rename(oldname, newname string) error {
	oldKey := d.objectKey(oldname)

	newKey := d.objectKey(newname)
	if oldKey == "" || newKey == "" {
		return fs.ErrInvalid
	}

	ctx := context.Background()

	copySource := url.PathEscape(fmt.Sprintf("%s/%s", d.bucket, oldKey))

	_, err := d.s3.CopyObject(ctx, &awss3.CopyObjectInput{ //nolint:exhaustruct
		Bucket:     aws.String(d.bucket),
		CopySource: aws.String(copySource),
		Key:        aws.String(newKey),
	})
	if err != nil {
		return fmt.Errorf("copy object: %w", err)
	}

	return d.Remove(oldname)
}

func (d *clientDriver) Stat(name string) (os.FileInfo, error) {
	clean := cleanPath(name)
	if clean == "." || clean == "/" {
		return &s3FileInfo{
			name:    "/",
			size:    0,
			mode:    os.ModeDir | dirPermission,
			modTime: time.Now(),
			dir:     true,
		}, nil
	}

	key := d.objectKey(name)
	if key == "" {
		return nil, fs.ErrNotExist
	}

	ctx := context.Background()

	out, err := d.s3.HeadObject(ctx, &awss3.HeadObjectInput{ //nolint:exhaustruct
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	if err == nil {
		return &s3FileInfo{
			name:    path.Base(clean),
			size:    aws.ToInt64(out.ContentLength),
			mode:    0,
			modTime: aws.ToTime(out.LastModified),
			dir:     false,
		}, nil
	}

	if !isNotFound(err) {
		return nil, fmt.Errorf("head object %q: %w", name, err)
	}

	ok, dirErr := d.directoryExists(key)
	if dirErr != nil {
		return nil, dirErr
	}

	if ok {
		return &s3FileInfo{
			name:    path.Base(clean),
			size:    0,
			mode:    os.ModeDir | dirPermission,
			modTime: time.Now(),
			dir:     true,
		}, nil
	}

	return nil, fs.ErrNotExist
}

func (d *clientDriver) Chmod(name string, mode os.FileMode) error { return nil }

func (d *clientDriver) Chown(name string, uid, gid int) error { return nil }

func (d *clientDriver) Chtimes(name string, atime, mtime time.Time) error { return nil }

// ReadDir implements ftpserver.ClientDriverExtensionFileList.
func (d *clientDriver) ReadDir(name string) ([]os.FileInfo, error) { //nolint:funlen
	prefix := d.objectKey(name)

	var prefixPtr *string

	if prefix != "" {
		p := ensureTrailingSlash(prefix)
		prefix = p
		prefixPtr = aws.String(p)
	}

	input := &awss3.ListObjectsV2Input{ //nolint:exhaustruct
		Bucket:    aws.String(d.bucket),
		Delimiter: aws.String("/"),
		Prefix:    prefixPtr,
	}

	ctx := context.Background()

	var infos []os.FileInfo

	paginator := awss3.NewListObjectsV2Paginator(d.s3, input)
	seenDirs := make(map[string]struct{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("list objects: %w", err)
		}

		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)
			if key == prefix {
				continue
			}

			shortName := trimPrefix(key, prefix)
			if strings.Contains(shortName, "/") {
				continue
			}

			infos = append(infos, &s3FileInfo{
				name:    shortName,
				size:    aws.ToInt64(obj.Size),
				mode:    0,
				modTime: aws.ToTime(obj.LastModified),
				dir:     false,
			})
		}

		for _, cp := range page.CommonPrefixes {
			dirName := trimPrefix(aws.ToString(cp.Prefix), prefix)

			dirName = strings.TrimSuffix(dirName, "/")
			if dirName == "" {
				continue
			}

			if _, ok := seenDirs[dirName]; ok {
				continue
			}

			seenDirs[dirName] = struct{}{}

			infos = append(infos, &s3FileInfo{
				name:    dirName,
				size:    0,
				mode:    os.ModeDir | dirPermission,
				modTime: time.Time{},
				dir:     true,
			})
		}
	}

	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name() < infos[j].Name()
	})

	return infos, nil
}

func (d *clientDriver) newWritableFile(name string) (afero.File, error) { //nolint:ireturn
	key := d.objectKey(name)
	if key == "" {
		return nil, fs.ErrInvalid
	}

	mem := d.builder()

	file, err := mem.Create(memFileName)
	if err != nil {
		return nil, fmt.Errorf("create buffer: %w", err)
	}

	return newS3File(file, mem, key, d.bucket, d.s3, true), nil
}

func (d *clientDriver) populateWritableFile(fileHandle afero.File, key string) error {
	bufferedFile, ok := fileHandle.(*s3File)
	if !ok {
		return errUnexpectedFileType
	}

	data, err := d.downloadObject(key)
	if err != nil {
		return err
	}

	if len(data) == 0 {
		return nil
	}

	var writeErr error

	_, writeErr = fileHandle.Write(data)
	if writeErr != nil {
		return fmt.Errorf("populate existing data: %w", writeErr)
	}

	_, seekErr := fileHandle.Seek(0, io.SeekEnd)
	if seekErr != nil {
		return fmt.Errorf("seek to end: %w", seekErr)
	}

	bufferedFile.key = key

	return nil
}

func (d *clientDriver) downloadObject(key string) ([]byte, error) {
	ctx := context.Background()

	out, err := d.s3.GetObject(ctx, &awss3.GetObjectInput{ //nolint:exhaustruct
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return nil, fs.ErrNotExist
		}

		return nil, fmt.Errorf("get object %q: %w", key, err)
	}

	body, readErr := io.ReadAll(out.Body)
	closeErr := out.Body.Close()

	if readErr != nil {
		if closeErr != nil {
			return nil, errors.Join(
				fmt.Errorf("read object body: %w", readErr),
				fmt.Errorf("close object body: %w", closeErr),
			)
		}

		return nil, fmt.Errorf("read object body: %w", readErr)
	}

	if closeErr != nil {
		return nil, fmt.Errorf("close object body: %w", closeErr)
	}

	return body, nil
}

func (d *clientDriver) ensureDirMarker(name string) error {
	key := ensureTrailingSlash(d.objectKey(name))
	if key == "" {
		return nil
	}

	ctx := context.Background()

	_, err := d.s3.PutObject(ctx, &awss3.PutObjectInput{ //nolint:exhaustruct
		Bucket: aws.String(d.bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(nil),
	})
	if err != nil {
		return fmt.Errorf("create dir marker %q: %w", name, err)
	}

	return nil
}

func (d *clientDriver) deleteByPrefix(prefix string) error {
	prefix = ensureTrailingSlash(prefix)
	ctx := context.Background()

	paginator := awss3.NewListObjectsV2Paginator(d.s3, &awss3.ListObjectsV2Input{ //nolint:exhaustruct
		Bucket: aws.String(d.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list for delete: %w", err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		ids := make([]s3types.ObjectIdentifier, 0, len(page.Contents))
		for _, obj := range page.Contents {
			ids = append(ids, s3types.ObjectIdentifier{Key: obj.Key}) //nolint:exhaustruct
		}

		_, err = d.s3.DeleteObjects(ctx, &awss3.DeleteObjectsInput{ //nolint:exhaustruct
			Bucket: aws.String(d.bucket),
			Delete: &s3types.Delete{
				Objects: ids,
				Quiet:   aws.Bool(true),
			},
		})
		if err != nil {
			return fmt.Errorf("delete objects: %w", err)
		}
	}

	return nil
}

func (d *clientDriver) objectKey(name string) string {
	clean := cleanPath(name)
	if clean == "." || clean == "/" || clean == "" {
		if d.prefix != "" {
			return d.prefix
		}

		return ""
	}

	clean = strings.TrimPrefix(clean, "/")
	if d.prefix != "" {
		return fmt.Sprintf("%s/%s", d.prefix, clean)
	}

	return clean
}

func (d *clientDriver) directoryExists(key string) (bool, error) {
	prefix := ensureTrailingSlash(key)
	if prefix == "" {
		return true, nil
	}

	ctx := context.Background()

	out, err := d.s3.ListObjectsV2(ctx, &awss3.ListObjectsV2Input{ //nolint:exhaustruct
		Bucket:  aws.String(d.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(1),
	})
	if err != nil {
		return false, fmt.Errorf("list for dir %q: %w", key, err)
	}

	return len(out.Contents) > 0 || len(out.CommonPrefixes) > 0, nil
}

func ensureTrailingSlash(key string) string {
	if key == "" {
		return ""
	}

	if strings.HasSuffix(key, "/") {
		return key
	}

	return key + "/"
}

func cleanPath(pathValue string) string {
	if pathValue == "" {
		return "/"
	}

	if !strings.HasPrefix(pathValue, "/") {
		pathValue = "/" + pathValue
	}

	return path.Clean(pathValue)
}

func trimPrefix(value, prefix string) string {
	if prefix == "" {
		return value
	}

	return strings.TrimPrefix(value, prefix)
}

func isNotFound(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()

		return code == "NotFound" || code == "NoSuchKey"
	}

	var noSuchKey *s3types.NoSuchKey

	return errors.As(err, &noSuchKey)
}
