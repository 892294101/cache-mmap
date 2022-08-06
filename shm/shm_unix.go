// Copyright (c) 2020 Meng Huang (mhboy@outlook.com)
// This package is licensed under a MIT license that can be found in the LICENSE file.

// +build darwin linux dragonfly freebsd netbsd openbsd

package shm

import (
	"fmt"
	"github.com/892294101/cache-mmap/ftok"
	"math/rand"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

const (
	// IPC_CREAT creates if key is nonexistent
	IPC_CREAT = 01000

	// IPC_EXCL fails if key exists.
	IPC_EXCL = 02000

	// IPC_NOWAIT returns error no wait.
	IPC_NOWAIT = 04000

	// IPC_PRIVATE is private key
	IPC_PRIVATE = 00000

	// SEM_UNDO sets up adjust on exit entry
	SEM_UNDO = 010000

	// IPC_RMID removes identifier
	IPC_RMID = 0
	// IPC_SET sets ipc_perm options.
	IPC_SET = 1
	// IPC_STAT gets ipc_perm options.
	IPC_STAT = 2
)

const (
	// SYS_SHMGET is syscall SYS_SHMGET constant
	SYS_SHMGET = syscall.SYS_SHMGET
	// SYS_SHMAT is syscall SYS_SHMAT constant
	SYS_SHMAT = syscall.SYS_SHMAT
	// SYS_SHMDT is syscall SYS_SHMDT constant
	SYS_SHMDT = syscall.SYS_SHMDT
	// SYS_SHMCTL is syscall SYS_SHMCTL constant
	SYS_SHMCTL = syscall.SYS_SHMCTL
)

/*// Open returns the fd.
func Open(name string, oflag int, perm int) (int, error) {
	return syscall.Open("/dev/shm/"+name, oflag, uint32(perm))
}

// Unlink unlinks the name.
func Unlink(name string) error {
	return os.Remove("/dev/shm/" + name)
}
*/

type shm struct {
	lock      sync.RWMutex
	shmId     uintptr
	shmKey    int     // 对应： Shared Memory Segments shmid
	shmAddr   uintptr // 对应： Shared Memory Segments key
	shmFlags  int
	shmSize   int
	shmHandle []byte
}

// NewShm calls the shmget and shmat system call.
func NewShm(size int) (*shm, error) {
	ts := time.Now().UnixNano()
	tsID := rand.New(rand.NewSource(ts))
	id, err := ftok.Ftok("/tmp", uint8(tsID.Intn(99999)))
	if err != nil {
		fmt.Println(1)
		return nil, err
	}

	shm := new(shm)
	shm.shmSize = size
	shm.shmFlags = IPC_CREAT | 0640
	shm.shmId = uintptr(id)

	err = shm.createSHMGET()
	if err != nil {
		fmt.Println(2)
		return nil, err
	}
	err = shm.at()
	if err != nil {
		fmt.Println(3)
		return nil, err
	}

	return shm, nil
}

func (s *shm) GetShmId() int {
	return s.shmKey
}

func (s *shm) GetShmKey() string {
	return fmt.Sprintf("0x%x", s.shmId)
}

// Get calls the shmget system call.
func (s *shm) createSHMGET() error {
	s.lock.Lock()
	r1, _, errno := syscall.Syscall(SYS_SHMGET, uintptr(s.shmId), uintptr(validSize(int64(s.shmSize))), uintptr(s.shmFlags))
	skey := int(r1)
	if skey < 0 {
		return syscall.Errno(errno)
	}
	s.shmKey = skey
	s.lock.Unlock()
	return nil
}

// At calls the shmat system call.
func (s *shm) at() error {
	s.lock.Lock()
	shmaddr, _, errno := syscall.Syscall(SYS_SHMAT, uintptr(s.shmKey), 0, uintptr(s.shmFlags))
	if int(shmaddr) < 0 {
		s.Remove()
		return syscall.Errno(errno)
	}
	s.shmAddr = shmaddr

	var sl = struct {
		addr uintptr
		len  int
		cap  int
	}{s.shmAddr, s.shmSize, s.shmSize}
	b := *(*[]byte)(unsafe.Pointer(&sl))
	s.shmHandle = b
	s.lock.Unlock()
	return nil
}

// Dt calls the shmdt system call.
func (s *shm) Detach() error {
	s.lock.Lock()
	r1, _, errno := syscall.Syscall(SYS_SHMDT, uintptr(unsafe.Pointer(&s.shmHandle[0])), 0, 0)
	if int(r1) < 0 {
		return syscall.Errno(errno)
	}
	s.lock.Unlock()
	return nil
}

// Remove removes the shm with the given id.
func (s *shm) Remove() error {
	s.lock.Lock()
	r1, _, errno := syscall.Syscall(SYS_SHMCTL, uintptr(s.shmKey), IPC_RMID, 0)
	if int(r1) < 0 {
		return syscall.Errno(errno)
	}
	s.lock.Unlock()
	return nil
}
