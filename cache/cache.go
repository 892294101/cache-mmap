package cache

type name struct {
}

/*
type Cache struct {
	cacheSize uint64              // 缓存大小
	handle    mmap.MMap           // map写入器
	config    *config.CacheConfig // 配置信息
	bw        bytesWriter         //字节写入器
	br        bytesReader         //字节读取器
}

func (c Cache) Close() error {
	if err := c.handle.Unmap(); err != nil {
		return err
	}
	return nil
}

func NewCache(opts ...config.Options) (*Cache, error) {

	//初始化Cache配置
	conf := config.NewConfig()
	for _, opt := range opts {
		if err := opt(conf); err != nil {
			return nil, err
		}
	}
	cc := new(Cache)
	cc.config = conf

	f, err := os.OpenFile(cc.config.Dir, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return nil, err
	}

	if err := f.Truncate(int64(cc.config.CacheSize)); err != nil {
		return nil, err
	}

	mm, err := mmap.Map(f, mmap.RDWR, 0)
	if err != nil {
		return nil, err
	}
	cc.handle = mm

	return cc, nil
}*/
