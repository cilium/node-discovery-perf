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

type virtualNode struct {
	t   time.Duration        // time that took the node to receive updates about all other nodes
	mgr *nodemanager.Manager // manager for this virtual node
}

func main() {
	var initialCount = flag.Int("initial-count", 1, "Number of concurrent node discovery agents set up initially")
	var additionalCount = flag.Int("additional-count", 0, "Number of nodes registered after initial nodes are set up")
	var address = flag.String("etcd-address", "127.0.0.1:2379", "etcd address")
	flag.Parse()

	err := kvstore.Setup("etcd", map[string]string{"etcd.address": *address})
	if err != nil {
		fmt.Println("error setting up kvstore:", err.Error())
		return
	}
	option.Config.IPv4ServiceRange = "auto"
	option.Config.IPv6ServiceRange = "auto"

	nodeCh := make(chan virtualNode)

	for i := 0; i < *initialCount; i++ {
		go registerAndWaitForOthers(nodeCh, *initialCount)
	}

	times := make([]float64, 0, *initialCount)
	managers := make([]*nodemanager.Manager, 0, *initialCount)

	for i := 0; i < *initialCount; i++ {
		n := <-nodeCh
		times = append(times, n.t.Seconds())
		managers = append(managers, n.mgr)
	}

	m, s := stat.MeanStdDev(times, nil)
	fmt.Printf("Mean discovery time: %fs, variance: %fs\n", m, s)

	if *additionalCount > 0 {
		timeCh := make(chan time.Duration)

		for i := 0; i < *additionalCount; i++ {
			go registerAndWaitForOthers(nodeCh, *initialCount+*additionalCount)
		}

		for _, m := range managers {
			go waitForCount(m, timeCh, *initialCount+*additionalCount)
		}
		times := make([]float64, 0, *initialCount)
		for i := 0; i < *initialCount; i++ {
			times = append(times, (<-timeCh).Seconds())
		}
		m, s = stat.MeanStdDev(times, nil)
		fmt.Printf("After adding %d nodes: Mean discovery time: %fs, variance: %fs\n", *additionalCount, m, s)
	}
}

func waitForCount(manager *nodemanager.Manager, timeCh chan time.Duration, count int) {
	var t time.Duration
	defer func() {
		timeCh <- t
	}()
	start := time.Now()

	for {
		nodes := manager.GetNodes()
		if len(nodes) >= count {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	t = time.Since(start)
}

func registerAndWaitForOthers(nodeChannel chan virtualNode, n int) {
	var localNode virtualNode
	defer func() {
		nodeChannel <- localNode
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
	localNode.mgr = nodeMngr
	nd := nodediscovery.NewNodeDiscovery(nodeMngr, mtuConfig)
	node.SetName(uid)

	start := time.Now()
	nd.StartDiscovery()
	<-nd.Registered

	for {
		nodes := nodeMngr.GetNodes()
		if len(nodes) >= n {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	localNode.t = time.Since(start)
}
