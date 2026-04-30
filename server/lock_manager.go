package server

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
)

var ErrInvalidPath = errors.New("invalid path")

type fileLock struct {
	mu sync.RWMutex
}

type LockManager struct {
	mu    sync.Mutex
	locks map[string]*fileLock
	root  string
}

func NewLockManager(root string) *LockManager {
	return &LockManager{
		locks: make(map[string]*fileLock),
		root:  filepath.Clean(root),
	}
}

func (lm *LockManager) normalize(path string) (string, error) {
	clean := filepath.Clean(path)

	if lm.root != "" {
		rel, err := filepath.Rel(lm.root, clean)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return "", ErrInvalidPath
		}
	}

	return clean, nil
}

func (lm *LockManager) get(path string) *fileLock {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lock, ok := lm.locks[path]
	if !ok {
		lock = &fileLock{}
		lm.locks[path] = lock
	}
	return lock
}

func (lm *LockManager) RLock(path string) (func(), error) {
	p, err := lm.normalize(path)
	if err != nil {
		return nil, err
	}

	lock := lm.get(p)
	lock.mu.RLock()

	return func() {
		lock.mu.RUnlock()
	}, nil
}

func (lm *LockManager) Lock(path string) (func(), error) {
	p, err := lm.normalize(path)
	if err != nil {
		return nil, err
	}

	lock := lm.get(p)
	lock.mu.Lock()

	return func() {
		lock.mu.Unlock()
	}, nil
}

func (lm *LockManager) Clear() {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.locks = make(map[string]*fileLock)
}
