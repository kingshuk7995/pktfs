package utils

import (
	"path/filepath"
	"strings"
	"sync"
)

// FileLock wraps a RWMutex for a single path
type FileLock struct {
	mu sync.RWMutex
}

// LockManager manages per-path locks
type LockManager struct {
	mu    sync.Mutex
	locks map[string]*FileLock
	root  string
}

// NewLockManager creates a new manager
func NewLockManager(root string) *LockManager {
	return &LockManager{
		locks: make(map[string]*FileLock),
		root:  filepath.Clean(root),
	}
}

// normalize ensures:
// 1. path is absolute within root
// 2. no traversal outside root
func (lm *LockManager) normalize(baseDir, path string) (string, error) {
	// join base + path
	full := filepath.Join(baseDir, path)
	clean := filepath.Clean(full)

	// enforce root sandbox if provided
	if lm.root != "" {
		rel, err := filepath.Rel(lm.root, clean)
		if err != nil || strings.HasPrefix(rel, "..") {
			return "", ErrInvalidPath
		}
	}

	return clean, nil
}

// get returns (or creates) the lock for a path
func (lm *LockManager) get(path string) *FileLock {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lock, ok := lm.locks[path]
	if !ok {
		lock = &FileLock{}
		lm.locks[path] = lock
	}
	return lock
}

// RLock acquires a read lock and returns an unlock func
func (lm *LockManager) RLock(baseDir, path string) (func(), error) {
	p, err := lm.normalize(baseDir, path)
	if err != nil {
		return nil, err
	}

	lock := lm.get(p)
	lock.mu.RLock()

	return func() {
		lock.mu.RUnlock()
	}, nil
}

// Lock acquires a write lock and returns an unlock func
func (lm *LockManager) Lock(baseDir, path string) (func(), error) {
	p, err := lm.normalize(baseDir, path)
	if err != nil {
		return nil, err
	}

	lock := lm.get(p)
	lock.mu.Lock()

	return func() {
		lock.mu.Unlock()
	}, nil
}

// TryLock attempts a write lock without blocking
func (lm *LockManager) TryLock(baseDir, path string) (func(), bool, error) {
	p, err := lm.normalize(baseDir, path)
	if err != nil {
		return nil, false, err
	}

	lock := lm.get(p)

	locked := lock.mu.TryLock()
	if !locked {
		return nil, false, nil
	}

	return func() {
		lock.mu.Unlock()
	}, true, nil
}

// TryRLock attempts a read lock without blocking
func (lm *LockManager) TryRLock(baseDir, path string) (func(), bool, error) {
	p, err := lm.normalize(baseDir, path)
	if err != nil {
		return nil, false, err
	}

	lock := lm.get(p)

	locked := lock.mu.TryRLock()
	if !locked {
		return nil, false, nil
	}

	return func() {
		lock.mu.RUnlock()
	}, true, nil
}

// Optional: cleanup method (manual, simple version)
func (lm *LockManager) Clear() {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.locks = make(map[string]*FileLock)
}

// --- Errors ---

var ErrInvalidPath = &PathError{"invalid path"}

type PathError struct {
	msg string
}

func (e *PathError) Error() string {
	return e.msg
}
