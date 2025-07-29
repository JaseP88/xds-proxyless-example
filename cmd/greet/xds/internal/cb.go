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
	log.Printf("OnStreamOpen")
	return nil
}

func (cb *MyCallbacks) OnStreamClosed(id int64, node *core.Node) {
	log.Printf("OnStreamClosed: streamId=%d node=%v", id, node)
}

func (cb *MyCallbacks) OnDeltaStreamOpen(_ context.Context, id int64, typ string) error {
	log.Printf("OnDeltaStreamOpen: streamId=%d type=%v", id, typ)
	return nil
}

func (cb *MyCallbacks) OnDeltaStreamClosed(id int64, node *core.Node) {
	log.Printf("OnDeltaStreamClosed: streamId=%d node=%v", id, node)
}

func (cb *MyCallbacks) OnStreamRequest(id int64, req *discovery.DiscoveryRequest) error {
	// Second callback handler to be invoked after OnStreamRequest.
	// Callback is invoked multiple times totaling the number of resources that is configured in the xDS server. (LDS, RDS, CDS, EDS, etc.)
	log.Printf("OnStreamRequest: streamId=%d request=%v", id, req)
	return nil
}

func (cb *MyCallbacks) OnStreamResponse(ctx context.Context, id int64, req *discovery.DiscoveryRequest, res *discovery.DiscoveryResponse) {
	log.Printf("OnStreamResponse: streamId=%d request=%v response=%v", id, req, res)
}

func (cb *MyCallbacks) OnStreamDeltaResponse(id int64, req *discovery.DeltaDiscoveryRequest, res *discovery.DeltaDiscoveryResponse) {
	log.Printf("OnStreamDeltaResponse: streamId=%d request=%v response=%v", id, req, res)
}

func (cb *MyCallbacks) OnStreamDeltaRequest(id int64, req *discovery.DeltaDiscoveryRequest) error {
	log.Printf("OnStreamDeltaRequest: streamId=%d request=%v", id, req)
	return nil
}

func (cb *MyCallbacks) OnFetchRequest(context.Context, *discovery.DiscoveryRequest) error {
	log.Println("OnFetchRequest")
	return nil
}

func (cb *MyCallbacks) OnFetchResponse(*discovery.DiscoveryRequest, *discovery.DiscoveryResponse) {
	log.Println("OnFetchResponse")
}
