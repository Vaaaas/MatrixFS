package sysTool

import(
	"sync"
)

type SafeMap struct {
	lock     *sync.RWMutex
	InnerMap map[interface{}]interface{}
}

// NewSafeMap 建立新的对象
func NewSafeMap() *SafeMap {
	return &SafeMap{
		lock:     new(sync.RWMutex),
		InnerMap: make(map[interface{}]interface{}),
	}
}

// Get from maps return the k's value
func (safeMap *SafeMap) Get(id interface{}) interface{} {
	safeMap.lock.RLock()
	defer safeMap.lock.RUnlock()
	if value, ok := safeMap.InnerMap[id]; ok {
		return value
	}
	return nil
}

// Set Maps the given key and value. Returns false
// if the key is already in the map and changes nothing.
func (safeMap *SafeMap) Set(id interface{}, node interface{}) bool {
	safeMap.lock.Lock()
	defer safeMap.lock.Unlock()
	safeMap.InnerMap[id] = node
	//if val, ok := safeMap.InnerMap[id]; !ok {
	//	safeMap.InnerMap[id] = node
	//} else if val != node {
	//	safeMap.InnerMap[id] = node
	//} else {
	//	return false
	//}
	return true
}

// Delete the given key and value.
func (safeMap *SafeMap) Delete(id interface{}) {
	safeMap.lock.Lock()
	defer safeMap.lock.Unlock()
	delete(safeMap.InnerMap, id)
}

// Items returns all items in safemap.
func (safeMap *SafeMap) Items() map[interface{}]interface{} {
	safeMap.lock.RLock()
	defer safeMap.lock.RUnlock()
	r := make(map[interface{}]interface{})
	for k, v := range safeMap.InnerMap {
		r[k] = v
	}
	return r
}

// Count returns the number of items within the map.
func (safeMap *SafeMap) Count() int {
	safeMap.lock.RLock()
	defer safeMap.lock.RUnlock()
	return len(safeMap.InnerMap)
}

// Check Returns true if k is exist in the map.
func (allNodes *SafeMap) Check(id interface{}) bool {
	allNodes.lock.RLock()
	defer allNodes.lock.RUnlock()
	_, ok := allNodes.InnerMap[id]
	return ok
}