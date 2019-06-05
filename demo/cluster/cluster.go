package cluster

import (
	"github.com/hashicorp/memberlist"
	//"io/ioutil"
	"log"
	"stathat.com/c/consistent"
	"time"
)

type Node interface {
	ShouldProcess(key string) (string, bool)
	Members() []string
	GetAddr() string
}

type node struct {
	*consistent.Consistent
	addr string
}

func (n *node) GetAddr() string {
	return n.addr
}

func (n *node) ShouldProcess(key string) (string, bool) {
	addr, _ := n.Get(key)
	return addr, addr == n.addr
}

func New(addr, cluster string) (Node, error) {
	conf := memberlist.DefaultLANConfig()
	conf.Name = addr
	conf.BindAddr = addr
	//conf.LogOutput = ioutil.Discard
	list, err := memberlist.Create(conf)
	if err != nil {
		return nil, err
	}
	if cluster == "" {
		cluster = addr
	}
	clus := []string{cluster}
	_, err = list.Join(clus)
	if err != nil {
		return nil, err
	}

	circle := consistent.New()
	circle.NumberOfReplicas = 256

	go func() {
		for {
			members := list.Members()
			nodes := make([]string, len(members))
			for i, n := range members {
				nodes[i] = n.Name
				log.Println("node :", n.Name)
			}
			circle.Set(nodes)
			time.Sleep(time.Second * 5)
			log.Println("Checking the cluster ...")
		}
	}()

	return &node{circle, addr}, nil
}
