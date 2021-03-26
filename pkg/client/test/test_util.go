// Copyright 2019-present Open Networking Foundation.
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

package test

import (
	"context"
	"fmt"
	"github.com/lucasbfernandes/go-client/pkg/client/primitive"
	netutil "github.com/lucasbfernandes/go-client/pkg/client/util/net"
	"github.com/atomix/go-framework/pkg/atomix/counter"
	"github.com/atomix/go-framework/pkg/atomix/election"
	"github.com/atomix/go-framework/pkg/atomix/indexedmap"
	"github.com/atomix/go-framework/pkg/atomix/leader"
	"github.com/atomix/go-framework/pkg/atomix/list"
	"github.com/atomix/go-framework/pkg/atomix/lock"
	"github.com/atomix/go-framework/pkg/atomix/log"
	"github.com/atomix/go-framework/pkg/atomix/map"
	atomixprimitive "github.com/atomix/go-framework/pkg/atomix/primitive"
	"github.com/atomix/go-framework/pkg/atomix/set"
	"github.com/atomix/go-framework/pkg/atomix/value"
	"github.com/atomix/go-local/pkg/atomix/local"
	"net"
)

const basePort = 5000

// StartTestPartitions starts the given number of local partitions and returns client connections for them
func StartTestPartitions(numPartitions int) ([]primitive.Partition, []chan struct{}) {
	partitions := make([]primitive.Partition, numPartitions)
	chans := make([]chan struct{}, numPartitions)
	for i := 0; i < numPartitions; i++ {
		partitionID := i + 1
		address, ch := startTestPartition(partitionID)
		partitions[i] = primitive.Partition{
			ID:      partitionID,
			Address: address,
		}
		chans[i] = ch
	}
	return partitions, chans
}

// startTestPartition starts a single local partition
func startTestPartition(partitionID int) (netutil.Address, chan struct{}) {
	for port := basePort; port < basePort+100; port++ {
		address := netutil.Address(fmt.Sprintf("localhost:%d", port))
		lis, err := net.Listen("tcp", string(address))
		if err != nil {
			continue
		}
		node := local.NewNode(lis, []atomixprimitive.PartitionID{atomixprimitive.PartitionID(partitionID)})
		counter.RegisterPrimitive(node)
		election.RegisterPrimitive(node)
		indexedmap.RegisterPrimitive(node)
		lock.RegisterPrimitive(node)
		log.RegisterPrimitive(node)
		leader.RegisterPrimitive(node)
		list.RegisterPrimitive(node)
		_map.RegisterPrimitive(node)
		set.RegisterPrimitive(node)
		value.RegisterPrimitive(node)
		node.Start()

		ch := make(chan struct{})
		go func() {
			<-ch
			node.Stop()
		}()
		return address, ch
	}
	panic("cannot find open port")
}

// OpenSessions opens sessions for the given partitions
func OpenSessions(partitions []primitive.Partition, opts ...primitive.SessionOption) ([]*primitive.Session, error) {
	sessions := make([]*primitive.Session, len(partitions))
	for i, partition := range partitions {
		session, err := primitive.NewSession(context.TODO(), partition, opts...)
		if err != nil {
			return nil, err
		}
		sessions[i] = session
	}
	return sessions, nil
}

// CloseSessions closes the given sessions
func CloseSessions(sessions []*primitive.Session) {
	for _, session := range sessions {
		_ = session.Close()
	}
}

// StopTestPartitions stops the given test partition channels
func StopTestPartitions(chans []chan struct{}) {
	for _, ch := range chans {
		close(ch)
	}
}
