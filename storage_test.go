package main

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
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

	region := "test"
	return S3Config{
		Region:         &region,
		Endpoint:       &ts.URL,
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

	cfg, closeS3 := newTestS3Config()
	defer closeS3()
	s, err := NewS3Storage(cfg)
	asrt.NoError(err)

	put := func(path, content string) {
		t.Helper()
		_, err := s.s3.PutObject(&s3.PutObjectInput{
			Body:   strings.NewReader(content),
			Bucket: aws.String("test"),
			Key:    aws.String(path),
		})
		asrt.NoError(err)
	}

	// Bucket doesn't exist.
	_, _, err = s.List("")
	asrt.Error(err)

	// Bucket exists, but has no content.
	_, err = s.s3.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String("test"),
	})
	asrt.NoError(err)

	_, _, err = s.List("")
	asrt.Error(err)

	// Single file.
	put("file1.jpg", "1")
	put("empty", "") // Empty files should be ignored.
	dirs, files, err := s.List("")
	asrt.NoError(err)
	asrt.Empty(dirs)
	asrt.Len(files, 1)
	asrt.Equal("file1.jpg", files[0].path)

	// Single directory.
	put("dir1/file2.jpg", "12")
	put("dir1/empty", "")
	dirs, files, err = s.List("")
	asrt.NoError(err)
	asrt.Len(dirs, 1)
	asrt.Equal("dir1", dirs[0].path)
	asrt.Len(dirs[0].Parents(), 1)
	asrt.Equal("", dirs[0].Parents()[0].path)
	asrt.Len(files, 1)
	asrt.Equal("file1.jpg", files[0].path)

	// Two directories.
	put("dir2/file3.jpg", "123")
	dirs, files, err = s.List("")
	asrt.NoError(err)
	asrt.Len(dirs, 2)
	asrt.Equal("dir1", dirs[0].path)
	asrt.Equal("dir2", dirs[1].path)
	asrt.Len(files, 1)
	asrt.Equal("file1.jpg", files[0].path)

	// Nested directories.
	put("dir2/dir22/file4.jpg", "1234")
	dirs, files, err = s.List("")
	asrt.NoError(err)
	asrt.Len(dirs, 2)
	asrt.Equal("dir1", dirs[0].path)
	asrt.Equal("dir2", dirs[1].path)
	asrt.Len(files, 1)
	asrt.Equal("file1.jpg", files[0].path)

	dirs, files, err = s.List("dir1")
	asrt.NoError(err)
	asrt.Empty(dirs)
	asrt.Len(files, 1)
	asrt.Equal("dir1/file2.jpg", files[0].path)
	asrt.Equal("file2.jpg", files[0].Name())
	asrt.Equal("file2", files[0].FriendlyName())

	dirs, files, err = s.List("dir2")
	asrt.NoError(err)
	asrt.Len(dirs, 1)
	asrt.Equal("dir2/dir22", dirs[0].path)
	asrt.Equal("dir22", dirs[0].Name())
	asrt.Len(dirs[0].Parents(), 2)
	asrt.Equal("", dirs[0].Parents()[0].path)
	asrt.Equal("dir2", dirs[0].Parents()[1].path)
	asrt.Len(files, 1)
	asrt.Equal("dir2/file3.jpg", files[0].path)

	dirs, files, err = s.List("dir2/dir22")
	asrt.NoError(err)
	asrt.Empty(dirs)
	asrt.Len(files, 1)
	asrt.Equal("dir2/dir22/file4.jpg", files[0].path)

	// Prefix doexn't exist.
	_, _, err = s.List("dir3")
	asrt.Error(err)

	_, _, err = s.List("dir2/dir23")
	asrt.Error(err)

	// Content URL.
	url, err := s.FileContentURL("file1.jpg")
	asrt.NoError(err)
	asrt.NotEmpty(url)

	url, err = s.FileContentURL("dir2/dir22/file4.jpg")
	asrt.NoError(err)
	asrt.NotEmpty(url)

	url, err = s.FileContentURL("dir2/dir22/file5.jpg")
	asrt.Error(err)
	asrt.Empty(url)

	// File size.
	size, err := s.FileSize("file1.jpg")
	asrt.NoError(err)
	asrt.EqualValues(1, size)

	size, err = s.FileSize("dir2/dir22/file4.jpg")
	asrt.NoError(err)
	asrt.EqualValues(4, size)

	_, err = s.FileSize("dir2/dir22/file5.jpg")
	asrt.Error(err)

	// Base prefix dir1.
	s.cfg.BasePrefix = "dir1/"
	dirs, files, err = s.List("")
	asrt.NoError(err)
	asrt.Empty(dirs)
	asrt.Len(files, 1)
	asrt.Equal("file2.jpg", files[0].path)

	// Base prefix dir2.
	s.cfg.BasePrefix = "dir2/"
	dirs, files, err = s.List("")
	asrt.NoError(err)
	asrt.Len(dirs, 1)
	asrt.Equal("dir22", dirs[0].path)
	asrt.Len(dirs[0].Parents(), 1)
	asrt.Equal("", dirs[0].Parents()[0].path)
	asrt.Len(files, 1)
	asrt.Equal("file3.jpg", files[0].path)

	dirs, files, err = s.List("dir22")
	asrt.NoError(err)
	asrt.Empty(dirs)
	asrt.Len(files, 1)
	asrt.Equal("dir22/file4.jpg", files[0].path)

	// Base prefix doesn't exist.
	s.cfg.BasePrefix = "dir3/"
	_, _, err = s.List("")
	asrt.Error(err)
}
