package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/892294101/cache-mmap/mmap"
	"os"
)

type name struct {
	admin string
}

func main() {
	/*c, err := cache.NewCache(cache.SetCacheSize(1000), cache.SetDirFile(`d:\tmp.dat`))
	defer c.Close()
	if err != nil {
		fmt.Println(err)
	}*/

	b, err := mmap.NewMmap("./test.dat", os.O_CREATE|os.O_RDWR, 128*(1<<17))
	if err != nil {
		fmt.Println("NewMmap", err)
		os.Exit(1)
	}
	defer func() {
		if err := b.Close(); err != nil {
			fmt.Println("unmap error: ", err)
		}
	}()

	buf := &bytes.Buffer{}
	obj := name{"asldjfalksdflasdjlfasjd;fasd;fasdfsadfasdf"}
	if err := binary.Write(buf, binary.LittleEndian, obj); err != nil {
		fmt.Println("binary:", err)
	}

	x, err := b.WriteAt(buf.Bytes(), 0)
	if err != nil {
		b.Flush()
		os.Exit(1)
	}

	fmt.Println("index: ", x)

	b.Flush()

	/*shm, err := shm.NewShm(8192)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer shm.Remove()
	defer shm.Detach()

	shm.WriteStringAt("收快递费机撒独立开发", 0)

	var des strings.Builder
	des.Grow(len("收快递费机撒独立开发"))

	shm.ReadStringAt(&des, 0)
	fmt.Println(des.String())

	fmt.Printf("key: %s id: %d \n", shm.GetShmKey(), shm.GetShmId())

	time.Sleep(time.Second * 5)
	*/
	/*mem, err := shm.NewSystemVMem(4325435, 10000)
	if err != nil {
		fmt.Println(mem)
	}*/

}
