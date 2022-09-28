// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// +build windows

package mmap

import (
	"os"
	"syscall"
	"unsafe"
)

const (
	PAGE_READONLY          = syscall.PAGE_READONLY
	PAGE_READWRITE         = syscall.PAGE_READWRITE
	PAGE_WRITECOPY         = syscall.PAGE_WRITECOPY
	PAGE_EXECUTE_READ      = syscall.PAGE_EXECUTE_READ
	PAGE_EXECUTE_READWRITE = syscall.PAGE_EXECUTE_READWRITE
	PAGE_EXECUTE_WRITECOPY = syscall.PAGE_EXECUTE_WRITECOPY

	FILE_MAP_COPY    = syscall.FILE_MAP_COPY
	FILE_MAP_WRITE   = syscall.FILE_MAP_WRITE
	FILE_MAP_READ    = syscall.FILE_MAP_READ
	FILE_MAP_EXECUTE = syscall.FILE_MAP_EXECUTE
)

// Offset returns the valid offset.
func Offset(offset int64) int64 {
	pageSize := int64(os.Getpagesize() * 16)
	return offset / pageSize * pageSize
}

func protFlags(p Prot) (prot int, flags int) {
	prot = PAGE_READONLY
	flags = FILE_MAP_READ
	if p&WRITE != 0 {
		prot = PAGE_READWRITE
		flags = FILE_MAP_WRITE
	}
	if p&COPY != 0 {
		prot = PAGE_WRITECOPY
		flags = FILE_MAP_COPY
	}
	if p&EXEC != 0 {
		prot <<= 4
		flags |= FILE_MAP_EXECUTE
	}
	return
}

func (m *File) mmap(fd int, offset int64, length int, prot int, flags int) (*File, error) {
	if length <= 0 {
		return nil, syscall.EINVAL
	}
	handle, err := syscall.CreateFileMapping(syscall.Handle(fd), nil, uint32(prot), uint32((offset+int64(length))>>32), uint32((offset+int64(length))&0xFFFFFFFF), nil)
	if err != nil {
		return nil, err
	}

	addr, err := syscall.MapViewOfFile(handle, uint32(flags), uint32(offset>>32), uint32(offset&0xFFFFFFFF), uintptr(length))
	if err != nil {
		return nil, err
	}
	err = syscall.CloseHandle(handle)
	if err != nil {
		return nil, err
	}

	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{addr, length, length}
	b := *(*[]byte)(unsafe.Pointer(&sl))
	p := &b[cap(b)-1]
	m.data[p] = b
	m.key = p
	m.length = int64(length)
	return m, nil
}

func (m *File) Flush() (err error) {
	if !m.dirty {
		return nil
	}
	b := m.data[m.key]
	slice := (*struct {
		addr uintptr
		len  int
		cap  int
	})(unsafe.Pointer(&b))
	p := &b[cap(b)-1]
	data := m.data[p]

	if data == nil || &b[0] != &data[0] {
		return syscall.EINVAL
	}
	m.dirty = false
	err = syscall.FlushViewOfFile(slice.addr, uintptr(slice.len))
	return err
}

func (m *File) fLock() error {
	// Windowns 不支持文件锁
	/*if err := syscall.Flock(int(m.rawFile.Fd()), syscall.LOCK_SH|syscall.LOCK_NB); err != nil {
		return errors.New(fmt.Sprintf("add checkpoint exclusive lock failed: %v", err))
	}*/
	return nil
}

func (m *File) fUnLock() error {
	// // Windowns 不支持文件锁
	// return syscall.Flock(int(m.rawFile.Fd()), syscall.LOCK_UN)
	return nil
}

func (m *File) unmap() (err error) {
	data := m.data[m.key]
	if len(data) == 0 || len(data) != cap(data) {
		return syscall.EINVAL
	}
	p := &data[cap(data)-1]
	b := m.data[p]
	if b == nil || &b[0] != &data[0] {
		return syscall.EINVAL
	}
	err = syscall.UnmapViewOfFile(uintptr(unsafe.Pointer(&b[0])))
	if err != nil {
		return err
	}
	delete(m.data, p)

	return nil
}
