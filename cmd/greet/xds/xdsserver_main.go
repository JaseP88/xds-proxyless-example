//   Copyright Steve Sloka 2021
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/JaseP88/xds-poc/cmd/greet/xds/internal"
	"github.com/JaseP88/xds-poc/cmd/greet/xds/internal/xds_resources"
	res "github.com/JaseP88/xds-poc/cmd/greet/xds/internal/xds_resources/singlecluster"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	// res "github.com/JaseP88/xds-poc/cmd/greet/xds/internal/xds_resources/weightedcluster"
	// res "github.com/JaseP88/xds-poc/cmd/greet/xds/internal/xds_resources/multiroutedcluster"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

var (
	l       internal.Logger
	port    uint
	address string
	ads     bool
	tls     bool
	simple  bool
)

func init() {
	l = internal.Logger{Debug: false}
	// address
	flag.StringVar(&address, "a", "127.0.0.1", "address")
	// The port that this xDS server listens on
	flag.UintVar(&port, "p", 18000, "xDS management server port")
	// ads
	flag.BoolVar(&ads, "ads", false, "ads")
	// tls
	flag.BoolVar(&tls, "tls", true, "tls")
	// simple
	flag.BoolVar(&simple, "simple", false, "whether to use simple or complex setup, with or without grpc server resources")
}

func main() {
	flag.Parse()

	// Create a cache
	cache := cache.NewSnapshotCache(ads, cache.IDHash{}, l)

	// Create the snapshot that we'll serve to Envoy
	register(cache)

	ctx := context.Background()
	cb := &internal.MyCallbacks{}
	srv := server.NewServer(ctx, cache, cb)

	// Run the xDS server
	go internal.RunServer(srv, address, port, tls)

	// Console Prompts for weighted lb
	reader := bufio.NewReader(os.Stdin)
	versionNum := 2
	for {
		log.Println("Press Enter to update weights")
		reader.ReadString('\n')

		targetWeightA, _ := reader.ReadString('\n')
		twA, _ := strconv.Atoi(strings.Trim(targetWeightA, "\n"))
		log.Printf("weight for Locality 1 with address %s:%d changed to %d", address, 50051, twA)

		targetWeightB, _ := reader.ReadString('\n')
		twB, _ := strconv.Atoi(strings.Trim(targetWeightB, "\n"))
		log.Printf("weight for serviceB with address %s:%d changed to %d", address, 50053, twB)

		newSnap := res.GenerateSnapshotClientSnapshot(strconv.Itoa(versionNum), uint32(twA), uint32(twB))
		versionNum++

		if err := newSnap.Consistent(); err != nil {
			l.Errorf("error generating new snapshot not: %+v\n%+v", newSnap, err)
			os.Exit(1)
		}

		if err := cache.SetSnapshot(context.Background(), "client123", newSnap); err != nil {
			l.Errorf("snapshot error %q for %+v", err, newSnap)
			os.Exit(1)
		}
	}
}

func register(cache cache.SnapshotCache) {
	clientSnapshot := res.GenerateSnapshotClientSnapshot("", 90, 10)
	registerSnapshot(clientSnapshot, cache, "client123")
	if !simple {
		server1Snapshot := resources.GenerateSnapshotServerSnapshot("", 50051)
		registerSnapshot(server1Snapshot, cache, "server1")
		server2Snapshot := resources.GenerateSnapshotServerSnapshot("", 50053)
		registerSnapshot(server2Snapshot, cache, "server2")
		server3Snapshot := resources.GenerateSnapshotServerSnapshot("", 50055)
		registerSnapshot(server3Snapshot, cache, "server3")
		server4Snapshot := resources.GenerateSnapshotServerSnapshot("", 50057)
		registerSnapshot(server4Snapshot, cache, "server4")
	}
}

func registerSnapshot(snap *cache.Snapshot, cache cache.SnapshotCache, id string) {
	ds := resources.DebugSnapshot(snap)
	l.Infof("%s", ds)
	if err := snap.Consistent(); err != nil {
		l.Errorf("snapshot inconsistency: %+v\n%+v", snap, err)
		os.Exit(1)
	}
	l.Infof("will serve snapshot %+v for nodeId:%s", snap, id)

	// Add the snapshot to the cache
	if err := cache.SetSnapshot(context.Background(), id, snap); err != nil {
		l.Errorf("snapshot error %q for %+v", err, snap)
		os.Exit(1)
	}
}
