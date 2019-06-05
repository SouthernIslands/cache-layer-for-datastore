package cache

import "log"

func New(typ string, ttl int, capacity int) Cache {
	var c Cache
	if typ == "inmemory" {
		c = newInMemoryCache(capacity, ttl)
	}
	if c == nil {
		panic("unknown cache type " + typ)
	}
	log.Println(typ, "is ready to serve")
	return c
}
