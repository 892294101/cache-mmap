package cache

import "time"

const (
	_defaultCacheMaxSize    = 128 * (1 << 22) // 512 M
	_defaultCacheMinSize    = 128 * (1 << 17) // 16 M
	_defaultCacheSize       = 128 * (1 << 20) // 128 M
	_defaultFlushPeriodTime = time.Second
	_defaultCacheFile       = "./tmp_cache.dat"
)

type CacheConfig struct {
	CacheSize       int
	Dir             string
	FlushPeriodTime time.Duration
}

type Options func(*CacheConfig) error

func NewConfig() *CacheConfig {
	return &CacheConfig{
		CacheSize:       _defaultCacheSize,
		FlushPeriodTime: _defaultFlushPeriodTime,
		Dir:             _defaultCacheFile,
	}
}

func SetCacheSize(size int) Options {
	return func(c *CacheConfig) error {
		if size < _defaultCacheMaxSize && size > _defaultCacheMinSize {
			c.CacheSize = size
		}
		return nil
	}
}

func SetDirFile(file string) Options {
	return func(c *CacheConfig) error {
		c.Dir = file
		return nil
	}
}
