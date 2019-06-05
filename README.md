# cache-layer-for-datastore
This project is implement in Go and has been tested at Ubuntu 18.04 with Go 1.24

Build and run Demo:
1. go build demo
2. ./demo -node {localhostaddress}
More options:
-node       local host ip address
-cluster    cluster host ip address
-ttl        time to live(sec) for cache entry
-capacity   LRU size


Build Benchmark and target it to Demo address
1. go build demobenchmark
2. ./demobenchmark -h {hostaddress}
More options:
-h  localhost 
-n  total number of requests 
-c  number of parallel connections
-op test operations, could be  get\set\mixed
-r  size of keyspace
-g  get ratio in all requests