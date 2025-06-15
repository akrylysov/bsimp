package storage

import (
	"github.com/ginqi7/bsimp/internal/config"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	cfg config.S3Config
}

func NewLocalStorage(cfg config.S3Config) (*LocalStorage, error) {
	store := LocalStorage{
		cfg: cfg,
	}
	return &store, nil
}

func (store *LocalStorage) List(p string) ([]*StorageDirectory, []*StorageFile, error) {

	var dirs []*StorageDirectory
	var files []*StorageFile
	basePath := *store.cfg.Endpoint
	entries, _ := os.ReadDir(filepath.Join(basePath, p))

	for _, entry := range entries {
		// fmt.Println(getPath(p, entry.Name()))
		if entry.IsDir() {
			dirs = append(dirs, NewStorageDirectory(getPath(p, entry.Name())))
		} else {
			files = append(files, NewStorageFile(getPath(p, entry.Name()), 10))
		}
	}
	return dirs, files, nil
}

func getPath(parent string, p string) string {
	if len(parent) == 0 {
		return p
	} else {
		return parent + "/" + p
	}
}

func (store *LocalStorage) FileContentURL(p string) (string, error) {
	return "/audio/" + p, nil
}

func (store *LocalStorage) OpenFile(p string) (*os.File, error) {
	basePath := *store.cfg.Endpoint
	return os.Open(filepath.Join(basePath, p))
}
