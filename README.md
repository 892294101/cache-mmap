# cache-mmap
 
n := new(name)
	n.admin = "ABCDEFG"
	data := unsafe.Pointer(&n)

	dataRe := (*[100]byte)(data)

	x, err = b.WriteAt((*dataRe)[0:], 10)
	if err != nil {
		b.Flush()
		fmt.Println("Msync2", x)
		os.Exit(1)
	}
	fmt.Println("index2: ", x)

	des := make([]byte, 100)
	_, err = b.ReadAt(des, 10)
	dataB := unsafe.Pointer(&des)

	dataRes := (*name)(dataB)
	fmt.Println("dataRes.admin", dataRes.admin)