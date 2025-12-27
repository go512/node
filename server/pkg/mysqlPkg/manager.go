package mysqlPkg

import (
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/singleflight"
	"reflect"
	"sync"
)

type Manager struct {
	single  *singleflight.Group
	clients sync.Map
	configs sync.Map
}

func NewManager(configs ManagerConfig) *Manager {
	mgr := &Manager{
		single: new(singleflight.Group),
	}

	mgr.Load(configs)

	return mgr
}

func (mgr *Manager) Config(name string) (*Config, error) {
	if mgr == nil || len(name) == 0 {
		return nil, fmt.Errorf("name is empty")
	}

	iface, ok := mgr.configs.Load(name)
	if !ok {
		return nil, fmt.Errorf("config not found")
	}

	config, ok := iface.(*Config)
	if !ok {
		return nil, fmt.Errorf("config is not *Config")
	}

	return config, nil
}

func (mgr *Manager) Add(name string, config *Config) {
	if mgr == nil || len(name) == 0 || config == nil {
		return
	}

	config.FillWithDefault()

	oldConfig, err := mgr.Config(name)
	if err != nil {
		mgr.configs.Store(name, config)

		return
	}

	if reflect.DeepEqual(oldConfig, config) {
		return
	}

	// store new config
	mgr.configs.Store(name, config)

	// remove old client
	mgr.clients.Delete(name)
}

func (mgr *Manager) Del(name string) {
	if mgr == nil || len(name) == 0 {
		return
	}

	mgr.clients.Delete(name)
	mgr.configs.Delete(name)
}

func (mgr *Manager) Load(configs ManagerConfig) {
	if mgr == nil {
		return
	}

	for name, config := range configs {
		mgr.Add(name, config)
	}
}

func (mgr *Manager) Reload(configs ManagerConfig) error {
	if mgr == nil {
		return nil
	}

	gerr := errgroup.Group{}
	for name, config := range configs {
		gerr.Go(func() error {
			client, err := mgr.GetClient(name)
			if err != nil {
				if err == errors.New("no config found") {
					err = nil
				}
				return err
			}
			return client.Reload(config)
		})
	}
	return gerr.Wait()
}
