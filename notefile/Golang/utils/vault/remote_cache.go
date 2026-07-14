package vault

import (
	"context"
	"log"
	"sync"
	"time"

	// "github.com/google/go-cmp/cmp"
	"golang.org/x/sync/singleflight"
)

type RemoteCell struct {
	sync.RWMutex
	sf                  singleflight.Group // 合并并发加载请求
	value               any
	name                string
	createdAt           time.Time
	updatedAt           time.Time
	loaded              bool
	isValid             bool
	valueLoader         ValueLoader
	valueLoadingTimeout time.Duration
	refreshInterval     time.Duration // zero means no refresh
}

type ValueLoader func(ctx context.Context) (value any, isValid bool, err error)

func InitRemoteCell(name string, valueLoader ValueLoader, timeout time.Duration, refreshInterval time.Duration) {
	now := time.Now()
	c := &RemoteCell{
		name:                name,
		createdAt:           now,
		valueLoader:         valueLoader,
		valueLoadingTimeout: timeout,
		refreshInterval:     refreshInterval,
	}

	log.Printf("init remote cell: %v", name)

	register(c)
	go preloadAndRefresh(c)
	return
}

func InitStaticCell(name string, value any) {
	now := time.Now()
	c := &RemoteCell{
		name:      name,
		value:     value,
		isValid:   true,
		loaded:    true,
		createdAt: now,
		updatedAt: now,
	}

	log.Printf("init static cell: %v", name)

	register(c)
	return
}

func preloadAndRefresh(c *RemoteCell) {
	if err := c.loadAndStore(false); err != nil {
		log.Printf("initial load failed: %v, %v", c.name, err)
	}
	if c.refreshInterval > 0 {
		loopRefresh(c)
	}
}

func loopRefresh(c *RemoteCell) {
	for {
		interval := c.refreshInterval
		time.Sleep(interval)
		Refresh(c)
	}
}

func Refresh(c *RemoteCell) {
	if err := c.loadAndStore(true); err != nil {
		log.Printf("refresh failed: %v, %v", c.name, err)
	}
}

func (c *RemoteCell) loadValue() (value any, isValid bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.valueLoadingTimeout)
	defer cancel()
	return c.valueLoader(ctx)
}

func (c *RemoteCell) loadAndStore(force bool) error {
	_, err, _ := c.sf.Do(c.name, func() (any, error) {
		// double-check：非强制且已加载则跳过
		if !force {
			c.RLock()
			loaded := c.loaded
			c.RUnlock()
			if loaded {
				return nil, nil
			}
		}

		value, isValid, err := c.loadValue()
		if err != nil {
			return nil, err
		}

		c.Lock()
		// if !c.loaded || !cmp.Equal(value, c.value) {
		if !c.loaded {
			c.value = value
			c.isValid = isValid
			c.updatedAt = time.Now()
			if c.loaded {
				log.Printf("cell refreshed: %v", c.name)
			} else {
				log.Printf("cell loaded: %v", c.name)
			}
		} else {
			c.isValid = isValid
		}
		c.loaded = true
		c.Unlock()

		return nil, nil
	})
	return err
}

func (c *RemoteCell) ensureLoaded() error {
	c.RLock()
	loaded := c.loaded
	c.RUnlock()
	if loaded {
		return nil
	}
	return c.loadAndStore(false)
}

func RefreshCell(name string) {
	c := ReadCell(name)
	Refresh(c)
}

func (c *RemoteCell) SetValue(value any) {
	c.Lock()
	defer c.Unlock()

	c.value = value
	c.updatedAt = time.Now()
	c.isValid = true
	c.loaded = true
}

func (c *RemoteCell) GetValue() (value any, ok bool) {
	if err := c.ensureLoaded(); err != nil {
		panic(err)
	}

	c.RLock()
	defer c.RUnlock()

	if c.loaded && c.isValid {
		return c.value, true
	}
	return nil, false
}
