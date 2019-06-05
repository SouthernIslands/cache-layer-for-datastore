package cache

import "container/list"

type Cache interface {
	Set(string, []byte) error
	Get(string) ([]byte, error)
	Del(string) error
	NewScanner() Scanner
	GetStat() Stat
	GetMap() map[string]*entry
	GetList() list.List
}
