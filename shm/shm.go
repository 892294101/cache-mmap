// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// Package shm provides a way to use System V shared memory.
package shm

import (
	"encoding/binary"
	"errors"
	"os"
	"strings"
	"syscall"
)

const (
	// O_RDONLY opens the file read-only.
	O_RDONLY int = syscall.O_RDONLY
	// O_WRONLY opens the file write-only.
	O_WRONLY int = syscall.O_WRONLY
	// O_RDWR opens the file read-write.
	O_RDWR int = syscall.O_RDWR

	// O_CREATE creates a new file if none exists.
	O_CREATE int = syscall.O_CREAT
)

var (
	// ErrUnmappedMemory is returned when a function is called on unmapped memory.
	ErrUnmappedMemory = errors.New("unmapped memory")
	// ErrIndexOutOfBound is returned when given offset lies beyond the mapped region.
	ErrIndexOutOfBound = errors.New("offset out of mapped region")
)

// validSize returns the valid size.
func validSize(size int64) int64 {
	pageSize := int64(os.Getpagesize())
	if size%pageSize == 0 {
		return size
	}
	return (size/pageSize + 1) * pageSize
}

func (s *shm) boundaryChecks(offset, numBytes int64) error {
	if s.shmHandle == nil {
		return ErrUnmappedMemory
	} else if offset+numBytes > int64(s.shmSize) || offset < 0 {
		return ErrIndexOutOfBound
	}
	return nil
}

func (s *shm) ReadAt(dest []byte, offset int64) (int, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if err := s.boundaryChecks(offset, int64(len(dest))); err != nil {
		return 0, err
	}
	return copy(dest, s.shmHandle[offset:]), nil
}

func (s *shm) WriteAt(src []byte, offset int64) (int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.boundaryChecks(offset, int64(len(src))); err != nil {
		return 0, err
	}
	return copy(s.shmHandle[offset:], src), nil
}

func (s *shm) WriteStringAt(src string, offset int64) (int, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.boundaryChecks(offset, int64(len(src))); err != nil {
		return 0, err
	}
	return copy(s.shmHandle[offset:], src), nil
}

func (s *shm) ReadStringAt(dest *strings.Builder, offset int64) (int, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	if err := s.boundaryChecks(offset, int64(dest.Len())); err != nil {
		return 0, err
	}
	dataLength := int64(s.shmSize) - offset
	emptySpace := int64(dest.Cap() - dest.Len())
	end := s.shmSize
	if dataLength > emptySpace {
		end = int(offset + emptySpace)
	}
	n, err := dest.Write(s.shmHandle[offset:end])
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ReadUint64At reads uint64 from offset.
func (s *shm) ReadUint64At(offset int64) (uint64, error) {
	s.lock.RLock()
	if err := s.boundaryChecks(offset, 8); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(s.shmHandle[offset : offset+8]), nil
}

// WriteUint64At writes num at offset.
func (s *shm) WriteUint64At(num uint64, offset int64) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if err := s.boundaryChecks(offset, 8); err != nil {
		return err
	}
	binary.LittleEndian.PutUint64(s.shmHandle[offset:offset+8], num)
	return nil
}
