package main

import (
	"context"
	"errors"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type storageEntry struct {
	path string
}

// URLPath returns the URL-escaped path suitable for embedding as a URL in templates.
func (e *storageEntry) URLPath() string {
	return (&url.URL{Path: e.path}).EscapedPath()
}

func (e *storageEntry) Name() string {
	_, file := path.Split(e.path)
	return file
}

func (e *storageEntry) String() string {
	return e.Name()
}

type StorageDirectory struct {
	storageEntry
}

func NewStorageDirectory(p string) *StorageDirectory {
	return &StorageDirectory{
		storageEntry{
			path: p,
		},
	}
}

func ReverseSlice[T any](s []T) {
	i := 0
	j := len(s) - 1
	for i < j {
		s[i], s[j] = s[j], s[i]
		i += 1
		j -= 1
	}
}

// Parents return a slice of all parent directories from the root.
// E.g. it returns [/, /a, /a/b] for /a/b.
func (e *StorageDirectory) Parents() []*StorageDirectory {
	if e.path == "" {
		// The root directory doesn't have any parents.
		return nil
	}
	var dirs []*StorageDirectory
	p := e.path
	for idx := strings.LastIndexByte(p, '/'); idx != -1; idx = strings.LastIndexByte(p, '/') {
		p = p[:idx]
		dirs = append(dirs, NewStorageDirectory(p))
	}

	// Append root directory.
	dirs = append(dirs, NewStorageDirectory(""))

	ReverseSlice(dirs)

	return dirs
}

type StorageFile struct {
	storageEntry
	Size int64
}

func NewStorageFile(p string, size int64) *StorageFile {
	return &StorageFile{
		storageEntry: storageEntry{
			path: p,
		},
		Size: size,
	}
}

// FriendlyName returns a user-friendly file name. The implementation just returns the name without extension.
func (e *StorageFile) FriendlyName() string {
	name, _ := splitNameExt(e.Name())
	return name
}

type S3Storage struct {
	s3        *s3.Client
	presigner *s3.PresignClient
	cfg       S3Config
}

func NewS3Storage(cfg S3Config) *S3Storage {
	awsConfig := aws.Config{Region: cfg.Region}
	if cfg.Credentials != nil {
		awsConfig.Credentials = credentials.NewStaticCredentialsProvider(cfg.Credentials.ID, cfg.Credentials.Secret, cfg.Credentials.Token)
	}
	svc := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.ForcePathStyle
	})
	return &S3Storage{
		s3:        svc,
		presigner: s3.NewPresignClient(svc),
		cfg:       cfg,
	}
}

// prefix returns an S3 prefix from a public user-provided path.
// prefix can be the entire key.
func (store *S3Storage) prefix(p string) string {
	prefix := path.Join(store.cfg.BasePrefix, p)
	return prefix
}

// path returns a public path exposed to the user from an internal S3 key.
func (store *S3Storage) path(key string) string {
	return strings.TrimRight(
		strings.TrimPrefix(key, store.cfg.BasePrefix),
		Delimiter,
	)
}

// List returns slices of directories and files under the given path.
func (store *S3Storage) List(ctx context.Context, p string) ([]*StorageDirectory, []*StorageFile, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(store.cfg.Bucket),
		Delimiter: aws.String(Delimiter),
	}
	prefix := store.prefix(p)
	if prefix != "" {
		input.Prefix = aws.String(prefix + Delimiter)
	}

	paginator := s3.NewListObjectsV2Paginator(store.s3, input)

	var prefixes []types.CommonPrefix
	var objects []types.Object

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, nil, err
		}
		prefixes = append(prefixes, page.CommonPrefixes...)
		for _, object := range page.Contents {
			// Ignore empty objects used to emulate empty directories.
			if aws.ToInt64(object.Size) != 0 {
				objects = append(objects, object)
			}
		}
	}

	if len(prefixes) == 0 && len(objects) == 0 {
		return nil, nil, errors.New("directory doesn't exist")
	}

	var dirs []*StorageDirectory
	var files []*StorageFile

	for _, prefix := range prefixes {
		dirs = append(dirs, NewStorageDirectory(store.path(*prefix.Prefix)))
	}

	for _, object := range objects {
		files = append(files, NewStorageFile(store.path(*object.Key), aws.ToInt64(object.Size)))
	}

	return dirs, files, nil
}

// FileSize returns size of the file under the given path.
func (store *S3Storage) FileSize(ctx context.Context, p string) (int64, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(store.cfg.Bucket),
		Key:    aws.String(store.prefix(p)),
	}
	resp, err := store.s3.HeadObject(ctx, input)
	if err != nil {
		return 0, err
	}
	return aws.ToInt64(resp.ContentLength), nil
}

// FileContentURL returns a publicly accessible URL for the file under the given path.
func (store *S3Storage) FileContentURL(ctx context.Context, p string) (string, error) {
	size, err := store.FileSize(ctx, p)
	if err != nil {
		return "", err
	}
	if size == 0 {
		return "", errors.New("no content")
	}
	presignedURL, err := store.presigner.PresignGetObject(ctx,
		&s3.GetObjectInput{
			Bucket: aws.String(store.cfg.Bucket),
			Key:    aws.String(store.prefix(p)),
		},
		s3.WithPresignExpires(time.Duration(store.cfg.RequestPresignExpiry)))
	if err != nil {
		return "", err
	}
	return presignedURL.URL, nil
}
