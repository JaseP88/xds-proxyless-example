This module contains the xDS server build from [envoy go-control-plane](https://github.com/envoyproxy/go-control-plane).  
[Walkthrough](https://github.com/envoyproxy/go-control-plane/tree/main/internal/example)
  
Configurations resources are predefined for simplicity.  
Resources are also broken up to demo specific setups
- [single cluster](/cmd/greet/xds/internal/xds_resources/singlecluster/client_resource.go)
- [weighted cluster](/cmd/greet/xds/internal/xds_resources/weightedcluster/client_resource.go)
- [multi routed cluster](/cmd/greet/xds/internal/xds_resources/multiroutedcluster/client_resource.go)  
Switch by commenting and uncommenting accordingly.
```go
	res "github.com/JaseP88/xds-poc/cmd/greet/xds/internal/xds_resources/singlecluster"
	// res "github.com/JaseP88/xds-poc/cmd/greet/xds/internal/xds_resources/weightedcluster"
	// res "github.com/JaseP88/xds-poc/cmd/greet/xds/internal/xds_resources/multiroutedcluster"
```

To learn xDS configurations visit [LEARNING_README](/cmd/greet/xds/LEARNING_README/)

Callback handlers are invoked when xDS clients (gRPC client & server) connects.  
Find callback handlers [here](/cmd/greet/xds/internal/cb.go)