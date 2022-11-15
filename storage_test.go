package main

import (
	"testing"

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
