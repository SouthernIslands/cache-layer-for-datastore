package cache

type Scanner interface {
	Scan() bool
	Key() string
	Value() []byte
	Close()
}

type pair struct {
	k string
	v []byte
}

type CacheScanner struct {
	pair
	pairChan  chan *pair
	closeChan chan struct{}
}

func (s *CacheScanner) Scan() bool {
	p, ok := <-s.pairChan
	if ok {
		s.k, s.v = p.k, p.v
	}
	return ok
}

func (s *CacheScanner) Key() string {
	return s.k
}

func (s *CacheScanner) Value() []byte {
	return s.v
}

func (s *CacheScanner) Close() {
	close(s.closeChan)
}

func (c *inMemoryCache) NewScanner() Scanner {
	pairChan := make(chan *pair)
	closeChan := make(chan struct{})

	go func() {
		defer close(pairChan)
		c.mutex.RLock()
		for k, v := range c.table {
			c.mutex.RUnlock()
			select {
			case <-closeChan:
				return
			case pairChan <- &pair{k, v.value}:
			}
			c.mutex.RLock()
		}
		c.mutex.RUnlock()
	}()

	return &CacheScanner{pair{}, pairChan, closeChan}
}
