package main

import (
	"fmt"
	"github.com/892294101/cache-mmap/mmap"
	"os"
	"strconv"
)

func main() {
	/*c, err := cache.NewCache(cache.SetCacheSize(1000), cache.SetDirFile(`d:\tmp.dat`))
	defer c.Close()
	if err != nil {
		fmt.Println(err)
	}*/

	file, err := os.OpenFile("./test.dat", os.O_CREATE|os.O_RDWR, 0775)
	if err != nil {
		panic(err)
	}
	file.Truncate(128 * (1 << 22))

	defer file.Close()

	b, err := mmap.NewMmap(int(file.Fd()), 0, 128*(1<<22), mmap.READ|mmap.WRITE)
	if err != nil {
		fmt.Println("NewMmap", err)
		os.Exit(1)
	}

	var n int64
	for i := 0; i < 100000000; i++ {
		x, err := b.WriteAt([]byte(strconv.Itoa(i)+" "), n)
		if err != nil {
			b.Flush()
			fmt.Println("Msync", err, n, i)
			os.Exit(1)
		}
		n += int64(x)
	}

	fmt.Println("index: ", n)
	b.Flush()

}
