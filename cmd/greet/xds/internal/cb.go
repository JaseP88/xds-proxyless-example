package internal

import (
	"context"
	"log"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discovery "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	"github.com/envoyproxy/go-control-plane/pkg/server/v3"
)


type MyCallbacks struct {}

var _ server.Callbacks = &MyCallbacks{}

func (cb *MyCallbacks) Report() {
	log.Println("Report")
}

func (cb *MyCallbacks) OnStreamOpen(_ context.Context, id int64, typ string) error {
	// First callback handler to be invoked upon the ADS stream being opened.
	log.Println("OnStreamOpen")
	return nil
}

func (cb *MyCallbacks) OnStreamClosed(id int64, node *core.Node) {
	log.Println("OnStreamClosed")
}

func (cb *MyCallbacks) OnDeltaStreamOpen(_ context.Context, id int64, typ string) error {	
	log.Println("OnDeltaStreamOpen")
	return nil
}

func (cb *MyCallbacks) OnDeltaStreamClosed(id int64, node *core.Node) {
	log.Println("OnDeltaStreamClosed")
}

func (cb *MyCallbacks) OnStreamRequest(id int64, req *discovery.DiscoveryRequest) error {
	// Second callback handler to be invoked after OnStreamRequest.
	// Callback is invoked multiple times totaling the number of resources that is configured in the xDS server. (LDS, RDS, CDS, EDS, etc.)
	log.Println("OnStreamRequest")
	return nil
}

func (cb *MyCallbacks) OnStreamResponse(ctx context.Context, id int64, req *discovery.DiscoveryRequest, res *discovery.DiscoveryResponse) {
	log.Println("OnStreamResponse")
}

func (cb *MyCallbacks) OnStreamDeltaResponse(id int64, req *discovery.DeltaDiscoveryRequest, res *discovery.DeltaDiscoveryResponse) {
	log.Println("OnStreamDeltaResponse")
}

func (cb *MyCallbacks) OnStreamDeltaRequest(int64, *discovery.DeltaDiscoveryRequest) error {
	log.Println("OnStreamDeltaRequest")
	return nil
}

func (cb *MyCallbacks) OnFetchRequest(context.Context, *discovery.DiscoveryRequest) error {
	log.Println("OnFetchRequest")
	return nil
}

func (cb *MyCallbacks) OnFetchResponse(*discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {
	log.Println("OnFetchResponse")
}
