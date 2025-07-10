package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"net"
	"strings"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	server "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
	"log/slog"

	inprocessor	"xdsmod/internal/processor"
)

var (
	listenerPort = flag.Int("port", 8001, "Control plane port")
)

type ManagementServerOptions struct {
	OnStreamOpen func(context.Context, int64, string) error

	OnStreamClosed func(int64, *core.Node)

	OnStreamRequest func(int64, *discoveryservice.DiscoveryRequest) error

	OnStreamResponse func(context.Context, int64, *discoveryservice.DiscoveryRequest, *discoveryservice.DiscoveryResponse)
}

func RunManagementServer(ctx context.Context, opts ManagementServerOptions) *inprocessor.Processor {
	processor := inprocessor.InitProcessor(ctx)
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *listenerPort))
	if err != nil {
		slog.Error("failed to listen", "err", err)
	}

	callbacks := server.CallbackFuncs{
		StreamOpenFunc:     opts.OnStreamOpen,
		StreamClosedFunc:   opts.OnStreamClosed,
		StreamRequestFunc:  opts.OnStreamRequest,
		StreamResponseFunc: opts.OnStreamResponse,
	}
	xs := server.NewServer(ctx, processor.XDSstore.Cache, callbacks)
	gs := grpc.NewServer()

	discoveryservice.RegisterAggregatedDiscoveryServiceServer(gs, xs)

	slog.Info("management server listening")
	go gs.Serve(lis)

	slog.Info("xDS management server serving.", "address", lis.Addr().String())
	processor.LoadResourceFile(os.Getenv("XDS_RESOURCE_FILE"))

	return processor 
}

type discoveryRequest struct {
	resourceName []string
	clientNodeId string
	listenerName string
	serverNodeId string
        xdstp        string
        region       string
        zone         string
        weight       string
        priority     string
}

func main() {
        flag.Parse()
	ctx := context.Background()

	slog.Info("Init xDS server(CP)")
	ldsRequestCh := make(chan *discoveryservice.DiscoveryRequest)

	processor := RunManagementServer(ctx, ManagementServerOptions{
		OnStreamRequest: func(id int64, req *discoveryservice.DiscoveryRequest) error {
			if len(req.GetResourceNames()) >= 1 && (req.GetTypeUrl() == "type.googleapis.com/envoy.config.listener.v3.Listener") && (strings.HasPrefix(req.GetResourceNames()[0], "grpc/server?xds.resource.listening_address=")) {
                             
				select {
			        case ldsRequestCh <- req:
					slog.Info("Received discovery request from", "server:", req.GetResourceNames()[0])
				default:
				}
			}
			return nil
		},
	})

	for {
		select {
		case req := <-ldsRequestCh:
		       processor.ProcessDiscoveryRequest(req)
		}
	}
}

