package main

import (
	"errors"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type storageEntry struct {
	path string
}

func (e *storageEntry) Path() string {
	return e.path
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
	s3  *s3.S3
	cfg S3Config
}

func NewS3Storage(cfg S3Config) (*S3Storage, error) {
	awsConfig := aws.Config{
		Region:   cfg.Region,
		Endpoint: cfg.Endpoint,
	}
	if cfg.Credentials != nil {
		awsConfig.Credentials = credentials.NewStaticCredentials(cfg.Credentials.ID, cfg.Credentials.Secret, cfg.Credentials.Token)
	}
	if cfg.ForcePathStyle {
		awsConfig.S3ForcePathStyle = aws.Bool(true)
	}
	sess, err := session.NewSession(&awsConfig)
	if err != nil {
		return nil, err
	}
	store := S3Storage{
		s3:  s3.New(sess),
		cfg: cfg,
	}
	return &store, nil
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
func (store *S3Storage) List(p string) ([]*StorageDirectory, []*StorageFile, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:    aws.String(store.cfg.Bucket),
		Delimiter: aws.String(Delimiter),
	}
	prefix := store.prefix(p)
	if prefix != "" {
		input.Prefix = aws.String(prefix + Delimiter)
	}

	var prefixes []*s3.CommonPrefix
	var objects []*s3.Object
	err := store.s3.ListObjectsV2Pages(input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		prefixes = append(prefixes, page.CommonPrefixes...)
		for _, object := range page.Contents {
			// Ignore empty objects used to emulate empty directories.
			if *object.Size != 0 {
				objects = append(objects, object)
			}
		}
		return true
	})
	if err != nil {
		return nil, nil, err
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
		files = append(files, NewStorageFile(store.path(*object.Key), *object.Size))
	}

	return dirs, files, nil
}

// FileSize returns size of the file under the given path.
func (store *S3Storage) FileSize(p string) (int64, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(store.cfg.Bucket),
		Key:    aws.String(store.prefix(p)),
	}
	resp, err := store.s3.HeadObject(input)
	if err != nil {
		return 0, err
	}
	return *resp.ContentLength, nil
}

// FileContentURL returns a publicly accessible URL for the file under the given path.
func (store *S3Storage) FileContentURL(p string) (string, error) {
	size, err := store.FileSize(p)
	if err != nil {
		return "", err
	}
	if size == 0 {
		return "", errors.New("no content")
	}
	req, _ := store.s3.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(store.cfg.Bucket),
		Key:    aws.String(store.prefix(p)),
	})
	return req.Presign(time.Duration(store.cfg.RequestPresignExpiry))
}
