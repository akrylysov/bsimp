package utils

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSplitNameExt(t *testing.T) {
	tests := []struct {
		fullName string
		name     string
		ext      string
	}{
		{
			fullName: "image.png",
			name:     "image",
			ext:      "png",
		},
		{
			fullName: "document.pdf",
			name:     "document",
			ext:      "pdf",
		},
		{
			fullName: "archive.tar.gz",
			name:     "archive.tar",
			ext:      "gz",
		},
		{
			fullName: "no_extension",
			name:     "no_extension",
			ext:      "",
		},
		{
			fullName: "file.",
			name:     "file",
			ext:      "",
		},
		{
			fullName: ".hidden",
			name:     "",
			ext:      "hidden",
		},
		{
			fullName: "",
			name:     "",
			ext:      "",
		},
		{
			fullName: "data.csv",
			name:     "data",
			ext:      "csv",
		},
		{
			fullName: "app.config.json",
			name:     "app.config",
			ext:      "json",
		},
	}
	asrt := assert.New(t)
	for _, test := range tests {
		name, ext := SplitNameExt(test.fullName)
		asrt.Equal(name, test.name)
		asrt.Equal(ext, test.ext)
	}
}
