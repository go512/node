package vault

import (
	"fmt"
	"sync"
)

var (
	lock  sync.RWMutex
	vault = make(map[string]any)
)

func ReadVault(key string) (value any) {
	cell := ReadCell(key)
	value, isValid := cell.GetValue()
	if !isValid {
		return nil
	}
	return
}

func ReadCell(key string) *RemoteCell {
	cell, found := vault[key]
	if !found {
		return nil
	}
	return cell.(*RemoteCell)
}

func register(cell *RemoteCell) {
	lock.Lock()
	defer lock.Unlock()

	if _, found := vault[cell.name]; found {
		panic(fmt.Sprintf("cell %s already exists", cell.name))
	}

	vault[cell.name] = cell
}
