package main

import (
	"demobenchmark/cacheClient"
	"flag"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type statistic struct {
	count int
	time  time.Duration
}

type result struct {
	getCount    int
	missCount   int
	setCount    int
	statBuckets []statistic
}

func (r *result) addStatistic(bucket int, stat statistic) {
	if bucket > len(r.statBuckets)-1 {
		newStatBuckets := make([]statistic, bucket+1)
		copy(newStatBuckets, r.statBuckets)
		r.statBuckets = newStatBuckets
	}
	s := r.statBuckets[bucket]
	s.count += stat.count
	s.time += stat.time
	r.statBuckets[bucket] = s
}

func (r *result) addDuration(d time.Duration, typ string) {
	bucket := int(d / time.Millisecond)
	r.addStatistic(bucket, statistic{1, d})
	if typ == "get" {
		r.getCount++
	} else if typ == "set" {
		r.setCount++
	} else {
		r.missCount++
	}
}

func (r *result) addResult(src *result) {
	for i, s := range src.statBuckets {
		r.addStatistic(i, s)
	}
	r.getCount += src.getCount
	r.missCount += src.missCount
	r.setCount += src.setCount

}

func run(client cacheClient.Client, c *cacheClient.Cmd, r *result) {
	start := time.Now()
	client.Run(c)
	duration := time.Now().Sub(start)
	resultType := c.Name
	if resultType == "get" {
		if c.Res == false {
			resultType = "miss"
		}
	}
	r.addDuration(duration, resultType)
}

func operate(id, count int, ch chan *result) {
	client := cacheClient.New(server)
	r := &result{0, 0, 0, make([]statistic, 0)}
	for i := 0; i < count; i++ {
		var tmp int

		tmp = cacheClient.GenerateIndex(1, keyspace+1,int(math.Sqrt(float64(keyspace/2))) , keyspace/2)
		//tmp = rand.Intn(keyspace)+1


		var value cacheClient.BookData
		value = cacheClient.GenerateValue(tmp)

		checkX := value.ISBN[len(value.ISBN)-1]
		if checkX == 'X' {
			//log.Println(value.ISBN, " has X")
			res := value.ISBN[:len(value.ISBN)-1] + "10"
			value.ISBN = res
		}

		key := value.ISBN
		name := operation

		if operation == "mixed" {
			if rand.Intn(10) < getratio {
				name = "get"
			} else {
				name = "set"
			}
		}

		c := &cacheClient.Cmd{name, key, value, nil, false}
		run(client, c, r)
	}

	ch <- r
}

var server, operation string
var total, threads, keyspace, getratio int

func init() {
	flag.StringVar(&server, "h", "localhost", "cache server address")
	flag.IntVar(&total, "n", 10000, "total number of requests")
	flag.IntVar(&threads, "c", 100, "number of parallel connections")
	flag.StringVar(&operation, "op", "get", "test get, could be get/set/mixed")
	flag.IntVar(&keyspace, "r", 200, "size of keyspace, use random book ISBN from 1 to keyspace")
	flag.IntVar(&getratio, "g", 7, "get ratio is")
	flag.Parse()

	fmt.Println("server is", server)
	fmt.Println("total request if ", total)
	fmt.Println("number of parallel connections is", threads)
	fmt.Println("operation is", operation)
	fmt.Println("keyspace is", keyspace)

	rand.Seed(time.Now().UnixNano())
}

func main() {
	ch := make(chan *result, threads)
	res := &result{0, 0, 0, make([]statistic, 0)}
	start := time.Now()
	if total < threads {
		threads = total
	}
	for i := 0; i < threads; i++ {
		go operate(i, total/threads, ch)
	}

	for i := 0; i < threads; i++ {
		res.addResult(<-ch)
	}

	d := time.Now().Sub(start)
	totalCount := res.getCount + res.missCount + res.setCount

	fmt.Printf("%d records get\n", res.getCount)
	fmt.Printf("%d records miss\n", res.missCount)
	fmt.Printf("%d records set\n", res.setCount)
	fmt.Printf("%f seconds total\n", d.Seconds())

	statCountSum := 0
	statTimeSum := time.Duration(0)
	for b, s := range res.statBuckets {
		if s.count == 0 {
			continue
		}
		statCountSum += s.count
		statTimeSum += s.time
		fmt.Printf("%d%% requests < %d ms\n", statCountSum*100/totalCount, b+1)
	}
	fmt.Printf("%d usec average for each request\n", int64(statTimeSum/time.Microsecond)/int64(statCountSum))
	fmt.Printf("rps is %f\n", float64(totalCount)/float64(d.Seconds()))
}
