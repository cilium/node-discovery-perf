package main

import (
	"flag"
	"fmt"
	"sync"

	fakedatapath "github.com/cilium/cilium/pkg/datapath/fake"
	"github.com/cilium/cilium/pkg/kvstore"
	"github.com/cilium/cilium/pkg/mtu"
	"github.com/cilium/cilium/pkg/node"
	nodemanager "github.com/cilium/cilium/pkg/node/manager"
	"github.com/cilium/cilium/pkg/nodediscovery"
	"github.com/cilium/cilium/pkg/option"

	"github.com/google/uuid"
)

func main() {
	var count = flag.Int("count", 1, "Number of concurrent node discovery agents")
	var address = flag.String("etcd-address", "127.0.0.1:2379", "etcd address")
	flag.Parse()

	err := kvstore.Setup("etcd", map[string]string{"etcd.address": *address})
	if err != nil {
		fmt.Println("error setting up kvstore:", err.Error())
		return
	}
	option.Config.IPv4ServiceRange = "auto"
	option.Config.IPv6ServiceRange = "auto"

	dp := fakedatapath.NewDatapath()
	nodeMngr, err := nodemanager.NewManager("all", dp.Node())
	if err != nil {
		fmt.Println("error creating nodemanager:", err.Error())
		return
	}

	mtuConfig := mtu.NewConfiguration(false, 1500)

	var wg sync.WaitGroup
	for i := 0; i < *count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			nd := nodediscovery.NewNodeDiscovery(nodeMngr, mtuConfig)
			id, err := uuid.NewRandom()
			if err != nil {
				fmt.Println("error generating uuid:", err.Error())
				return
			}
			node.SetName(id.String())
			nd.StartDiscovery()
			<-nd.Registered
		}()
	}
	wg.Wait()
}
