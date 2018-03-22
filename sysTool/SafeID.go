package sysTool

import (
	"sync"
)

type SafeID struct {
	lock      *sync.RWMutex
	IDCounter uint
}

func NewSafeID() *SafeID {
	return &SafeID{
		lock:      new(sync.RWMutex),
		IDCounter: 0,
	}
}

func (id *SafeID) GetSafeID() uint {
	id.lock.RLock()
	defer id.lock.RUnlock()
	return id.IDCounter
}

func (id *SafeID) PlusSafeID() {
	id.lock.RLock()
	defer id.lock.RUnlock()
	id.IDCounter++
}
