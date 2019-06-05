package main

import (
	"demo/cache"
	"demo/cluster"
	"demo/http"
	"flag"
	"fmt"
	"log"
)

func main() {
	fmt.Println("Cache service initializing ....")

	node := flag.String("node", "127.0.0.1", "node address")
	clus := flag.String("cluster", "", "cluster address")
	ttl := flag.Int("ttl", 10, "Time to live (sec)")
	capacity := flag.Int("capacity", 40, "LRU capacity")
	flag.Parse()

	tmp := "inmemory"
	typ := &tmp
	log.Println("Node address is :", *node)
	log.Println("Cluster address is :", *clus)
	log.Println("Now cache ttl is :", *ttl)
	log.Println("Now cache LRU capacity is :", *capacity)
	c := cache.New(*typ, *ttl, *capacity)

	n, e := cluster.New(*node, *clus)
	if e != nil {
		panic(e)
	}
	http.New(c, n).Listen()
}
