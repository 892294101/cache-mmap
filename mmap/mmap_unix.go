package mmap

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"syscall"
	"unsafe"
)

const (
	PROT_READ  = syscall.PROT_READ
	PROT_WRITE = syscall.PROT_WRITE
	PROT_EXEC  = syscall.PROT_EXEC

	MAP_SHARED  = syscall.MAP_SHARED
	MAP_PRIVATE = syscall.MAP_PRIVATE
	MAP_COPY    = MAP_PRIVATE
)

// Offset returns the valid offset.
func Offset(offset int64) int64 {
	pageSize := int64(os.Getpagesize())
	return offset / pageSize * pageSize
}

func protFlags(p Prot) (prot int, flags int) {
	prot = PROT_READ
	flags = MAP_SHARED
	if p&WRITE != 0 {
		prot |= PROT_WRITE
	}
	if p&COPY != 0 {
		flags = MAP_COPY
	}
	if p&EXEC != 0 {
		prot |= PROT_EXEC
	}
	return
}

func (m *File) mmap(fd int, offset int64, length int, prot int, flags int) (*File, error) {
	data, err := syscall.Mmap(fd, offset, length, prot, flags)
	if err != nil {
		return nil, err
	}
	key := byte(rand.Uint32())
	m.key = &key
	m.data[m.key] = data
	m.length = int64(length)
	return m, nil
}

func (m *File) Flush() error {
	if !m.dirty {
		return nil
	}
	_, _, err := syscall.Syscall(syscall.SYS_MSYNC, uintptr(unsafe.Pointer(&m.data[m.key][0])), uintptr(m.length), uintptr(syscall.MS_SYNC))
	if err != 0 {
		return err
	}
	m.dirty = false
	return nil
}

func (m *File) fLock() error {
	if err := syscall.Flock(int(m.rawFile.Fd()), syscall.LOCK_SH|syscall.LOCK_NB); err != nil {
		return errors.New(fmt.Sprintf("add checkpoint exclusive lock failed: %v", err))
	}
	return nil
}

func (m *File) fUnLock() error {
	return syscall.Flock(int(m.rawFile.Fd()), syscall.LOCK_UN)
}

func (m *File) unmap() (err error) {
	return syscall.Munmap(m.data[m.key])
}
