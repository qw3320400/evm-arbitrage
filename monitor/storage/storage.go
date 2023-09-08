package storage

import (
	"sync"
	"time"
)

const (
	StoreKeyUniswapv2Pairs = "Uniswapv2Pairs"
)

var (
	AllDatasStorage = map[string]*DatasStorage{
		StoreKeyUniswapv2Pairs: {},
	}
)

func GetStorage(key string) *DatasStorage {
	return AllDatasStorage[key]
}

type DataUpdate interface {
	NeedUpdate(interface{}) bool
	Expired(int64) bool
}

type DatasStorage struct {
	datas map[interface{}]interface{}
	lock  sync.RWMutex
}

func (s *DatasStorage) Load(key interface{}) interface{} {
	if s == nil {
		return nil
	}
	if s.datas == nil {
		s.lock.Lock()
		s.datas = map[interface{}]interface{}{}
		s.lock.Unlock()
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.datas[key]
}

func (s *DatasStorage) Store(keys []interface{}, datas []interface{}) {
	if s == nil || len(datas) == 0 || len(keys) != len(datas) {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if s.datas == nil {
		s.datas = map[interface{}]interface{}{}
	}
	for i, data := range datas {
		if old, ok := s.datas[keys[i]]; ok && !old.(DataUpdate).NeedUpdate(data) {
			continue
		}
		s.datas[keys[i]] = data
	}
}

func (s *DatasStorage) LoadAll() map[interface{}]interface{} {
	if s == nil {
		return map[interface{}]interface{}{}
	}
	if s.datas == nil {
		s.lock.Lock()
		s.datas = map[interface{}]interface{}{}
		s.lock.Unlock()
	}
	s.lock.RLock()
	defer s.lock.RUnlock()
	all := map[interface{}]interface{}{}
	for key, one := range s.datas {
		if one.(DataUpdate).Expired(time.Now().Unix()) {
			delete(s.datas, key)
			continue
		}
		all[key] = one
	}
	return all
}
