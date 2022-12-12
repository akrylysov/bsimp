package main

import (
	"sort"
)

type MediaListing struct {
	CurrentDirectory *StorageDirectory
	Directories      []*StorageDirectory
	Files            []*StorageFile
	Cover            *StorageFile
	AudioTracks      []*StorageFile
}

type MediaLibrary struct {
	store *S3Storage
}

func NewMediaLibrary(store *S3Storage) *MediaLibrary {
	return &MediaLibrary{
		store: store,
	}
}

func (ml *MediaLibrary) findCover(files []*StorageFile) *StorageFile {
	candidates := ScoreCovers(files)
	if len(candidates) == 0 {
		// TODO: look at subdirectories such as covers, scans and artwork.
		return nil
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})
	return candidates[0].StorageFile
}

// List returns directory listing under the provided path.
func (ml *MediaLibrary) List(p string) (*MediaListing, error) {
	dirs, files, err := ml.store.List(p)
	if err != nil {
		return nil, err
	}

	// Find album cover.
	cover := ml.findCover(files)

	// Find audio tracks and separate all other files.
	var tracks []*StorageFile
	var otherFiles []*StorageFile
	for _, f := range files {
		if IsAudioFile(f) {
			tracks = append(tracks, f)
		} else if cover == nil || f.Path() != cover.Path() {
			otherFiles = append(otherFiles, f)
		}
	}

	listing := &MediaListing{
		CurrentDirectory: NewStorageDirectory(p),
		Directories:      dirs,
		Files:            otherFiles,
		Cover:            cover,
		AudioTracks:      tracks,
	}
	return listing, nil
}

// ContentURL returns a public URL to a file under the given path.
func (ml *MediaLibrary) ContentURL(p string) (string, error) {
	return ml.store.FileContentURL(p)
}
