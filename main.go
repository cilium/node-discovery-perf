// Copyright 2019 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	fakedatapath "github.com/cilium/cilium/pkg/datapath/fake"
	"github.com/cilium/cilium/pkg/kvstore"
	"github.com/cilium/cilium/pkg/mtu"
	nodemanager "github.com/cilium/cilium/pkg/node/manager"
	"github.com/cilium/cilium/pkg/nodediscovery"
	"github.com/cilium/cilium/pkg/option"

	"github.com/google/uuid"
	"gonum.org/v1/gonum/stat"
)

var waitTime = 1 * time.Second

type virtualNode struct {
	t   time.Duration        // time that took the node to receive updates about all other nodes
	mgr *nodemanager.Manager // manager for this virtual node
}

func main() {
	// initialCount is for measuring how long it takes for `initialCount` nodes to discover each other
	var initialCount = flag.Int("initial-count", 1, "Number of concurrent node discovery agents set up initially")
	// additionalCount is number of nodes to add after initial discovery is done
	// it's for measuring how much time does it take for all nodes that were initially
	// created to discover `additionalCount` new node(s)
	var additionalCount = flag.Int("additional-count", 0, "Number of nodes registered after initial nodes are set up")
	// is a number of total nodes for test minus nodes from current test
	// for example if running 3 nodeperf clients in 3 pods, this will be `2*initialCount`
	// (if `initialCount` is the same for all created clients)
	var externalCount = flag.Int("external-count", 0, "Number of nodes to expect from other nodeperf clients")
	var address = flag.String("etcd-address", "127.0.0.1:2379", "etcd address")
	var etcdConfig = flag.String("etcd-config", "", "etcd config file")
	flag.Parse()

	var err error

	if *etcdConfig == "" {
		fmt.Println("Setting address")
		err = kvstore.Setup("etcd", map[string]string{"etcd.address": *address}, nil)
	} else {
		fmt.Println("Setting config")
		err = kvstore.Setup("etcd", map[string]string{"etcd.config": *etcdConfig}, nil)
	}

	if err != nil {
		fmt.Println("error setting up kvstore:", err.Error())
		return
	}
	option.Config.IPv4ServiceRange = "auto"
	option.Config.IPv6ServiceRange = "auto"

	nodeCh := make(chan virtualNode)

	for i := 0; i < *initialCount; i++ {
		go registerAndWaitForOthers(nodeCh, *initialCount, *externalCount)
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
			go registerAndWaitForOthers(nodeCh, *initialCount+*additionalCount, *externalCount)
		}

		for _, m := range managers {
			go waitForCount(m, timeCh, *initialCount+*additionalCount+*externalCount)
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
		time.Sleep(waitTime)
	}
	t = time.Since(start)
}

func registerAndWaitForOthers(nodeChannel chan virtualNode, initialCount, externalCount int) {
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

	start := time.Now()
	nd.StartDiscovery(uid)
	<-nd.Registered

	for {
		nodes := nodeMngr.GetNodes()
		if len(nodes) >= initialCount+externalCount {
			break
		}
		time.Sleep(waitTime)
	}
	localNode.t = time.Since(start)
}
