package storage

import "os"

type UniversalStorage struct {
	S3    *S3Storage
	Local *LocalStorage
}

func (store *UniversalStorage) List(p string) ([]*StorageDirectory, []*StorageFile, error) {
	if store.S3 != nil {
		return store.S3.List(p)
	}
	if store.Local != nil {
		return store.Local.List(p)
	}
	return nil, nil, nil
}

func (store *UniversalStorage) FileContentURL(p string) (string, error) {
	if store.S3 != nil {
		return store.S3.FileContentURL(p)
	}
	if store.Local != nil {
		return store.Local.FileContentURL(p)
	}
	return "", nil
}

func (store *UniversalStorage) OpenFile(p string) (*os.File, error) {
	if store.Local != nil {
		return store.Local.OpenFile(p)
	}
	return nil, nil
}
