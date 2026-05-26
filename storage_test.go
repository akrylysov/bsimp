package main

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/johannesboyne/gofakes3"
	"github.com/johannesboyne/gofakes3/backend/s3mem"
	"github.com/stretchr/testify/assert"
)

func files(paths ...string) []*StorageFile {
	var files []*StorageFile
	for _, p := range paths {
		files = append(files, NewStorageFile(p, 1))
	}
	return files
}

func dirs(paths ...string) []*StorageDirectory {
	var dirs []*StorageDirectory
	for _, p := range paths {
		dirs = append(dirs, NewStorageDirectory(p))
	}
	return dirs
}

func TestStorageEntry_URLPath(t *testing.T) {
	testCases := []struct {
		p        string
		expected string
	}{
		{p: "", expected: ""},
		{p: "plain", expected: "plain"},
		{p: "a/b/c", expected: "a/b/c"},
		{p: "Album #1", expected: "Album%20%231"},
		{p: "Artist/Album #1/Track ?.mp3", expected: "Artist/Album%20%231/Track%20%3F.mp3"},
		{p: "50% off", expected: "50%25%20off"},
		{p: "a&b/c+d", expected: "a&b/c+d"},
		{p: "Психохирурги", expected: "%D0%9F%D1%81%D0%B8%D1%85%D0%BE%D1%85%D0%B8%D1%80%D1%83%D1%80%D0%B3%D0%B8"},
	}
	for _, tc := range testCases {
		t.Run(tc.p, func(t *testing.T) {
			assert.Equal(t, tc.expected, NewStorageDirectory(tc.p).URLPath())
			assert.Equal(t, tc.expected, NewStorageFile(tc.p, 0).URLPath())
		})
	}
}

func TestStorageDirectory_Parents(t *testing.T) {
	testCases := []struct {
		p        string
		expected []*StorageDirectory
	}{
		{
			p:        "",
			expected: dirs(),
		},
		{
			p:        "a",
			expected: dirs(""),
		},
		{
			p:        "a/b",
			expected: dirs("", "a"),
		},
		{
			p:        "a/b/c",
			expected: dirs("", "a", "a/b"),
		},
	}
	for _, tc := range testCases {
		dir := NewStorageDirectory(tc.p)
		assert.EqualValues(t, tc.expected, dir.Parents())
	}
}

func newTestS3Config() (S3Config, func()) {
	backend := s3mem.New()
	faker := gofakes3.New(backend)
	ts := httptest.NewServer(faker.Server())

	return S3Config{
		Region:         "test",
		Endpoint:       ts.URL,
		Bucket:         "test",
		ForcePathStyle: true,
		Credentials: &S3Credentials{
			ID:     "id1",
			Secret: "secret1",
		},
		RequestPresignExpiry: Duration(time.Minute),
	}, ts.Close
}

func TestS3Storage(t *testing.T) {
	asrt := assert.New(t)

	ctx := context.Background()
	cfg, closeS3 := newTestS3Config()
	defer closeS3()
	s := NewS3Storage(cfg)

	put := func(path, content string) {
		t.Helper()
		_, err := s.s3.PutObject(ctx, &s3.PutObjectInput{
			Body:   strings.NewReader(content),
			Bucket: aws.String("test"),
			Key:    aws.String(path),
		})
		asrt.NoError(err)
	}

	// Bucket doesn't exist.
	_, _, err := s.List(ctx, "")
	asrt.Error(err)

	// Bucket exists, but has no content.
	_, err = s.s3.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String("test"),
	})
	asrt.NoError(err)

	_, _, err = s.List(ctx, "")
	asrt.Error(err)

	// Single file.
	put("file1.jpg", "1")
	put("empty", "") // Empty files should be ignored.
	dirs, files, err := s.List(ctx, "")
	asrt.NoError(err)
	asrt.Empty(dirs)
	asrt.Len(files, 1)
	asrt.Equal("file1.jpg", files[0].path)

	// Single directory.
	put("dir1/file2.jpg", "12")
	put("dir1/empty", "")
	dirs, files, err = s.List(ctx, "")
	asrt.NoError(err)
	asrt.Len(dirs, 1)
	asrt.Equal("dir1", dirs[0].path)
	asrt.Len(dirs[0].Parents(), 1)
	asrt.Equal("", dirs[0].Parents()[0].path)
	asrt.Len(files, 1)
	asrt.Equal("file1.jpg", files[0].path)

	// Two directories.
	put("dir2/file3.jpg", "123")
	dirs, files, err = s.List(ctx, "")
	asrt.NoError(err)
	asrt.Len(dirs, 2)
	asrt.Equal("dir1", dirs[0].path)
	asrt.Equal("dir2", dirs[1].path)
	asrt.Len(files, 1)
	asrt.Equal("file1.jpg", files[0].path)

	// Nested directories.
	put("dir2/dir22/file4.jpg", "1234")
	dirs, files, err = s.List(ctx, "")
	asrt.NoError(err)
	asrt.Len(dirs, 2)
	asrt.Equal("dir1", dirs[0].path)
	asrt.Equal("dir2", dirs[1].path)
	asrt.Len(files, 1)
	asrt.Equal("file1.jpg", files[0].path)

	dirs, files, err = s.List(ctx, "dir1")
	asrt.NoError(err)
	asrt.Empty(dirs)
	asrt.Len(files, 1)
	asrt.Equal("dir1/file2.jpg", files[0].path)
	asrt.Equal("file2.jpg", files[0].Name())
	asrt.Equal("file2", files[0].FriendlyName())

	dirs, files, err = s.List(ctx, "dir2")
	asrt.NoError(err)
	asrt.Len(dirs, 1)
	asrt.Equal("dir2/dir22", dirs[0].path)
	asrt.Equal("dir22", dirs[0].Name())
	asrt.Len(dirs[0].Parents(), 2)
	asrt.Equal("", dirs[0].Parents()[0].path)
	asrt.Equal("dir2", dirs[0].Parents()[1].path)
	asrt.Len(files, 1)
	asrt.Equal("dir2/file3.jpg", files[0].path)

	dirs, files, err = s.List(ctx, "dir2/dir22")
	asrt.NoError(err)
	asrt.Empty(dirs)
	asrt.Len(files, 1)
	asrt.Equal("dir2/dir22/file4.jpg", files[0].path)

	// Prefix doexn't exist.
	_, _, err = s.List(ctx, "dir3")
	asrt.Error(err)

	_, _, err = s.List(ctx, "dir2/dir23")
	asrt.Error(err)

	// Content URL.
	url, err := s.FileContentURL(ctx, "file1.jpg")
	asrt.NoError(err)
	asrt.NotEmpty(url)

	url, err = s.FileContentURL(ctx, "dir2/dir22/file4.jpg")
	asrt.NoError(err)
	asrt.NotEmpty(url)

	url, err = s.FileContentURL(ctx, "dir2/dir22/file5.jpg")
	asrt.Error(err)
	asrt.Empty(url)

	// File size.
	size, err := s.FileSize(ctx, "file1.jpg")
	asrt.NoError(err)
	asrt.EqualValues(1, size)

	size, err = s.FileSize(ctx, "dir2/dir22/file4.jpg")
	asrt.NoError(err)
	asrt.EqualValues(4, size)

	_, err = s.FileSize(ctx, "dir2/dir22/file5.jpg")
	asrt.Error(err)

	// Base prefix dir1.
	s.cfg.BasePrefix = "dir1/"
	dirs, files, err = s.List(ctx, "")
	asrt.NoError(err)
	asrt.Empty(dirs)
	asrt.Len(files, 1)
	asrt.Equal("file2.jpg", files[0].path)

	// Base prefix dir2.
	s.cfg.BasePrefix = "dir2/"
	dirs, files, err = s.List(ctx, "")
	asrt.NoError(err)
	asrt.Len(dirs, 1)
	asrt.Equal("dir22", dirs[0].path)
	asrt.Len(dirs[0].Parents(), 1)
	asrt.Equal("", dirs[0].Parents()[0].path)
	asrt.Len(files, 1)
	asrt.Equal("file3.jpg", files[0].path)

	dirs, files, err = s.List(ctx, "dir22")
	asrt.NoError(err)
	asrt.Empty(dirs)
	asrt.Len(files, 1)
	asrt.Equal("dir22/file4.jpg", files[0].path)

	// Base prefix doesn't exist.
	s.cfg.BasePrefix = "dir3/"
	_, _, err = s.List(ctx, "")
	asrt.Error(err)
}
