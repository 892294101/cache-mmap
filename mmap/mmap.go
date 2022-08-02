// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// Package mmap provides a way to memory-map a file.
package mmap

import (
	"encoding/binary"
	"errors"
	"strings"
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
	data   map[*byte][]byte
	length int64
	key    *byte
	dirty  bool
}

func ProtFlags(p Prot) (prot int, flags int) {
	return protFlags(p)
}

func NewMmap(fd int, offset int64, length int, p Prot) (*File, error) {
	prot, flags := ProtFlags(p)
	return newMmap(fd, offset, length, prot, flags)
}

func newMmap(fd int, offset int64, length int, prot int, flags int) (*File, error) {
	m := new(File)
	m.data = make(map[*byte][]byte)
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
	if err := m.boundaryChecks(offset, int64(len(dest))); err != nil {
		return 0, err
	}
	return copy(dest, m.data[m.key][offset:]), nil
}

func (m *File) WriteAt(src []byte, offset int64) (int, error) {
	m.dirty = true
	if err := m.boundaryChecks(offset, int64(len(src))); err != nil {
		return 0, err
	}
	return copy(m.data[m.key][offset:], src), nil
}

func (m *File) WriteStringAt(src string, offset int64) (int, error) {
	m.dirty = true
	if err := m.boundaryChecks(offset, int64(len(src))); err != nil {
		return 0, err
	}
	m.dirty = true
	return copy(m.data[m.key][offset:], src), nil
}

func (m *File) ReadStringAt(dest *strings.Builder, offset int64) (int, error) {
	if err := m.boundaryChecks(offset, int64(dest.Len())); err != nil {
		return 0, err
	}

	dataLength := m.length - offset
	emptySpace := int64(dest.Cap() - dest.Len())
	end := m.length
	if dataLength > emptySpace {
		end = offset + emptySpace
	}

	n, _ := dest.Write(m.data[m.key][offset:end])
	return n, nil
}

// ReadUint64At reads uint64 from offset.
func (m *File) ReadUint64At(offset int64) uint64 {
	m.boundaryChecks(offset, 8)
	return binary.LittleEndian.Uint64(m.data[m.key][offset : offset+8])
}

// WriteUint64At writes num at offset.
func (m *File) WriteUint64At(num uint64, offset int64) {
	m.dirty = true
	m.boundaryChecks(offset, 8)
	m.dirty = true
	binary.LittleEndian.PutUint64(m.data[m.key][offset:offset+8], num)
}
