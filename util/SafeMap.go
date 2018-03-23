package util

import (
	"sync"
)

type SafeMap struct {
	lock     *sync.RWMutex
	InnerMap map[interface{}]interface{}
}

//NewSafeMap 建立新的对象
func NewSafeMap() *SafeMap {
	return &SafeMap{
		lock:     new(sync.RWMutex),
		InnerMap: make(map[interface{}]interface{}),
	}
}

//Get 利用ID获取对象
func (safeMap *SafeMap) Get(id interface{}) interface{} {
	safeMap.lock.RLock()
	defer safeMap.lock.RUnlock()
	if value, ok := safeMap.InnerMap[id]; ok {
		return value
	}
	return nil
}

//Set 将ID对应值设置为指定对象
func (safeMap *SafeMap) Set(id interface{}, node interface{}) bool {
	safeMap.lock.Lock()
	defer safeMap.lock.Unlock()
	safeMap.InnerMap[id] = node
	return true
}

//Delete ID及对应的对象
func (safeMap *SafeMap) Delete(id interface{}) {
	safeMap.lock.Lock()
	defer safeMap.lock.Unlock()
	delete(safeMap.InnerMap, id)
}

//Item 返回map的副本
func (safeMap *SafeMap) Items() map[interface{}]interface{} {
	safeMap.lock.RLock()
	defer safeMap.lock.RUnlock()
	r := make(map[interface{}]interface{})
	for k, v := range safeMap.InnerMap {
		r[k] = v
	}
	return r
}

//Count 返回map里元素个数
func (safeMap *SafeMap) Count() int {
	safeMap.lock.RLock()
	defer safeMap.lock.RUnlock()
	return len(safeMap.InnerMap)
}