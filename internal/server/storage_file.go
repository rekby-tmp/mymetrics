package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/rekby-tmp/mymetrics/internal/common"
	"os"
	"sync/atomic"
	"time"
)

type FileStorage struct {
	filePath string

	storage MemStorage

	closed        atomic.Bool
	storeInterval time.Duration
	storeTimer    *time.Timer
}

func NewFileStorage(path string, storeInterval time.Duration) *FileStorage {
	res := &FileStorage{
		storage:       *NewMemStorage(),
		filePath:      path,
		storeInterval: storeInterval,
	}

	return res
}

func (s *FileStorage) Close() error {
	s.closed.Store(true)
	return nil
}

func (s *FileStorage) Get(name string, metricType common.MetricType) (val any, err error) {
	if s.closed.Load() {
		return nil, errors.New("failed to get value from closed storage")
	}

	return s.storage.Get(name, metricType)
}

func (s *FileStorage) List() (map[common.MetricType][]string, error) {
	if s.closed.Load() {
		return nil, errors.New("failed to get list values from closed storage")
	}

	return s.storage.List()
}

func (s *FileStorage) Store(name string, metricType common.MetricType, val any) (err error) {
	_, err = s.StoreAndGet(name, metricType, val)
	return err
}

func (s *FileStorage) StoreAndGet(name string, metricType common.MetricType, val any) (newVal any, err error) {
	s.storage.m.Lock()
	defer s.storage.m.Unlock()

	if s.closed.Load() {
		return nil, errors.New("failed to store value to closed storage")
	}

	newVal, err = s.storage.storeAndGet(name, metricType, val)
	if err != nil {
		return nil, err
	}

	if s.storeInterval == 0 {
		if err = s.flush(); err != nil {
			return nil, err
		}
	} else if s.storeTimer == nil {
		s.storeTimer = time.AfterFunc(s.storeInterval, func() {
			_ = s.Flush()
		})
	}

	return newVal, nil
}

func (s *FileStorage) LoadFromFile(path string) error {
	var stored fileStructForStore
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open storage file for read: %w", err)
	}
	err = json.NewDecoder(f).Decode(&stored)
	if err != nil {
		return fmt.Errorf("failed to unmarshal storage file: %w", err)
	}

	s.storage.m.Lock()
	defer s.storage.m.Unlock()

	s.storage.counter = stored.Counters
	if s.storage.counter == nil {
		s.storage.counter = map[string]int64{}
	}
	s.storage.gauge = stored.Gauges
	if s.storage.gauge == nil {
		s.storage.gauge = map[string]float64{}
	}
	return nil
}

func (s *FileStorage) Flush() error {
	s.storage.m.Lock()
	defer s.storage.m.Unlock()

	if s.closed.Load() {
		return errors.New("flush on closed storage")
	}

	return s.flush()
}

func (s *FileStorage) flush() error {
	content, err := s.marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal internal state: %w", err)
	}

	tmpFile := s.filePath + ".tmp"
	err = os.WriteFile(tmpFile, content, 0600)
	if err != nil {
		return fmt.Errorf("failed to write file %q: %w", s.filePath, err)
	}
	err = os.Rename(tmpFile, s.filePath)
	if err != nil {
		return fmt.Errorf("file rename tmp file while save storage: %w", err)
	}
	return nil
}

func (s *FileStorage) marshal() ([]byte, error) {
	var store fileStructForStore

	store.Counters = s.storage.counter
	store.Gauges = s.storage.gauge

	return json.Marshal(store)
}

type fileStructForStore struct {
	Counters map[string]int64
	Gauges   map[string]float64
}
