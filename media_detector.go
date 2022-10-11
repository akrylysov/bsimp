package main

import "strings"

type StringSet map[string]struct{}

func NewStringSet(vs ...string) StringSet {
	ss := make(StringSet, len(vs))
	for _, v := range vs {
		ss.Add(v)
	}
	return ss
}

func (ss StringSet) Add(v string) {
	ss[v] = struct{}{}
}

func (ss StringSet) Contains(v string) bool {
	_, ok := ss[v]
	return ok
}

// TODO: this should probably live in the config.
var audioExtensions = NewStringSet("mp3", "m4a", "aac", "ogg", "oga", "flac")

// IsAudioFile returns whether the given file is an audio file.
func IsAudioFile(f *StorageFile) bool {
	_, ext := splitNameExt(strings.ToLower(f.Name()))
	return audioExtensions.Contains(ext)
}

type ScoredFile struct {
	*StorageFile
	Score int
}

var imageExtensions = NewStringSet("jpg", "jpeg", "png", "gif")
var coverNames = NewStringSet("cover", "front", "folder")

func splitNameExt(fullName string) (string, string) {
	idx := strings.LastIndexByte(fullName, '.')
	if idx == -1 {
		return fullName, ""
	}
	return fullName[:idx], fullName[idx+1:]
}

func scoreCover(f *StorageFile) int {
	name, ext := splitNameExt(strings.ToLower(f.Name()))
	if !imageExtensions.Contains(ext) {
		return -1
	}
	// Exact match.
	if coverNames.Contains(name) {
		return 2
	}
	// Partial match.
	for pattern := range coverNames {
		if strings.Contains(name, pattern) {
			return 1
		}
	}
	// Any image.
	return 0
}

// ScoreCovers returns a slice of image files scored as album covers.
// The highest score is more likely to be a cover image.
func ScoreCovers(files []*StorageFile) []ScoredFile {
	var scored []ScoredFile
	for _, f := range files {
		score := scoreCover(f)
		if score == -1 {
			continue
		}
		scored = append(scored, ScoredFile{
			StorageFile: f,
			Score:       score,
		})
	}
	return scored
}
