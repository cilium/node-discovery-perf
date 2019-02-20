package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	fakedatapath "github.com/cilium/cilium/pkg/datapath/fake"
	"github.com/cilium/cilium/pkg/kvstore"
	"github.com/cilium/cilium/pkg/mtu"
	"github.com/cilium/cilium/pkg/node"
	nodemanager "github.com/cilium/cilium/pkg/node/manager"
	"github.com/cilium/cilium/pkg/nodediscovery"
	"github.com/cilium/cilium/pkg/option"

	"github.com/google/uuid"
	"gonum.org/v1/gonum/stat"
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

	timesCh := make(chan float64)

	for i := 0; i < *count; i++ {
		go func() {
			var t time.Duration
			defer func() {
				timesCh <- t.Seconds()
			}()
			id, err := uuid.NewRandom()
			if err != nil {
				fmt.Println("error generating uuid:", err.Error())
				return
			}

			uid := strings.Replace(id.String(), "-", "", -1)

			mtuConfig := mtu.NewConfiguration(false, 1500)
			dp := fakedatapath.NewDatapath()
			nodeMngr, err := nodemanager.NewManager(uid, dp.Node())
			if err != nil {
				fmt.Println("error creating nodemanager:", err.Error())
				return
			}
			nd := nodediscovery.NewNodeDiscovery(nodeMngr, mtuConfig)
			node.SetName(uid)

			start := time.Now()
			nd.StartDiscovery()
			<-nd.Registered

			for {
				nodes := nodeMngr.GetNodes()
				if len(nodes) >= *count {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			t = time.Since(start)
		}()
	}

	times := make([]float64, *count)
	for i := 0; i < *count; i++ {
		times = append(times, <-timesCh)
	}

	m, s := stat.MeanStdDev(times, nil)
	fmt.Printf("Mean discovery time: %fs, variance: %fs\n", m, s)
}
