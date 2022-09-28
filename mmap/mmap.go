// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// Package mmap provides a way to memory-map a file.
package mmap

import (
	"encoding/binary"
	"errors"
	"os"
	"strings"
	"sync"
)

// Prot is the protection flag.
type Prot int

const (
	// READ represents the read prot
	READ Prot = 1 << iota
	// WRITE represents the write prot
	WRITE
	// COPY represents the copy prot
	COPY
	// EXEC represents the exec prot
	EXEC
)

var (
	// ErrUnmappedMemory is returned when a function is called on unmapped memory.
	ErrUnmappedMemory = errors.New("unmapped memory")
	// ErrIndexOutOfBound is returned when given offset lies beyond the mapped region.
	ErrIndexOutOfBound = errors.New("offset out of mapped region")
)

type File struct {
	rawFile *os.File
	lock    sync.RWMutex
	data    map[*byte][]byte
	length  int64
	key     *byte
	dirty   bool
}

func ProtFlags(p Prot) (prot int, flags int) {
	return protFlags(p)
}

func validSize(size int64) int64 {
	pageSize := int64(os.Getpagesize())
	if size%pageSize == 0 {
		return size
	}
	return (size/pageSize + 1) * pageSize
}

func NewMmap(f string, flag int, size int64) (*File, error) {
	file, err := os.OpenFile(f, flag, 0775)
	if err != nil {
		return nil, err
	}

	if err := file.Truncate(validSize(size)); err != nil {
		return nil, err
	}

	b, err := openMmap(file, int(file.Fd()), 0, int(validSize(size)), READ|WRITE)
	if err != nil {
		return nil, err
	}
	if err := b.FLock(); err != nil {
		return nil, err
	}
	return b, nil
}

func openMmap(rawFile *os.File, fd int, offset int64, length int, p Prot) (*File, error) {
	prot, flags := ProtFlags(p)
	return newMmap(rawFile, fd, offset, length, prot, flags)
}

func newMmap(rawFile *os.File, fd int, offset int64, length int, prot int, flags int) (*File, error) {
	m := new(File)
	m.data = make(map[*byte][]byte)
	m.rawFile = rawFile
	return m.mmap(fd, offset, length, prot, flags)
}

func (m *File) boundaryChecks(offset, numBytes int64) error {
	if m.data[m.key] == nil {
		return ErrUnmappedMemory
	} else if offset+numBytes > m.length || offset < 0 {
		return ErrIndexOutOfBound
	}
	return nil
}

func (m *File) ReadAt(dest []byte, offset int64) (int, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if err := m.boundaryChecks(offset, int64(len(dest))); err != nil {
		return 0, err
	}
	return copy(dest, m.data[m.key][offset:]), nil
}

func (m *File) WriteAt(src []byte, offset int64) (int, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.dirty = true
	if err := m.boundaryChecks(offset, int64(len(src))); err != nil {
		return 0, err
	}
	return copy(m.data[m.key][offset:], src), nil
}

func (m *File) WriteStringAt(src string, offset int64) (int, error) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.dirty = true
	if err := m.boundaryChecks(offset, int64(len(src))); err != nil {
		return 0, err
	}
	m.dirty = true
	return copy(m.data[m.key][offset:], src), nil
}

func (m *File) ReadStringAt(dest *strings.Builder, offset int64) (int, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if err := m.boundaryChecks(offset, int64(dest.Len())); err != nil {
		return 0, err
	}

	dataLength := m.length - offset
	emptySpace := int64(dest.Cap() - dest.Len())
	end := m.length
	if dataLength > emptySpace {
		end = offset + emptySpace
	}

	n, err := dest.Write(m.data[m.key][offset:end])
	if err != nil {
		return 0, err
	}
	return n, nil
}

// ReadUint64At reads uint64 from offset.
func (m *File) ReadUint64At(offset int64) (uint64, error) {
	m.lock.RLock()
	defer m.lock.RUnlock()
	if err := m.boundaryChecks(offset, 8); err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(m.data[m.key][offset : offset+8]), nil
}

// WriteUint64At writes num at offset.
func (m *File) WriteUint64At(num uint64, offset int64) error {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.dirty = true
	if err := m.boundaryChecks(offset, 8); err != nil {
		return err
	}
	m.dirty = true
	binary.LittleEndian.PutUint64(m.data[m.key][offset:offset+8], num)
	return nil
}
