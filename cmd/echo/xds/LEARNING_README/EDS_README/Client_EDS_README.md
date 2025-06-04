# Overview
EDS or Cluster Load Assignment is the last later of Discovery Request.  
It holds all information regarding Endpoint information for the Grpc Clients to discover.  
Such information like priority, load balancing weights, locality, addresses, etc.  

EDS has a relation ship with [CDS](/cmd/echo/xds/LEARNING_README/CDS_README/Client_CDS_README.md)

## References
[EDS RFC](https://github.com/grpc/proposal/blob/master/A27-xds-global-load-balancing.md#eds)  

### Grpc Client specific EDS
```go
func makeEndpoint() *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: "Cluster",
		Endpoints: []*endpoint.LocalityLbEndpoints{
			{
				Locality: &core.Locality{
					Region:  "Region1",
					Zone:    "local1",
					SubZone: "local1",
				},
				Priority: uint32(0), //0 is highest and is default
				LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 90},
				LbEndpoints: []*endpoint.LbEndpoint{
					{
						HealthStatus: core.HealthStatus_HEALTHY,
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: &core.Address{
									Address: &core.Address_SocketAddress{
										SocketAddress: &core.SocketAddress{
											Protocol: core.SocketAddress_TCP,
											Address:  "127.0.0.1",,
											PortSpecifier: &core.SocketAddress_PortValue{
												PortValue: 50051,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Locality: &core.Locality{
					Region:  "Region2",
					Zone:    "local2",
					SubZone: "local2",
				},
				Priority: uint32(0),
				LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 10},
				LbEndpoints: []*endpoint.LbEndpoint{
					{
						HealthStatus: core.HealthStatus_HEALTHY,
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: &core.Address{
									Address: &core.Address_SocketAddress{
										SocketAddress: &core.SocketAddress{
											Protocol: core.SocketAddress_TCP,
											Address:  "127.0.0.1",
											PortSpecifier: &core.SocketAddress_PortValue{
												PortValue: 50053,
											},
										},
									},
								},
							},
						},
					},
					{
						HealthStatus: core.HealthStatus_HEALTHY,
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: &core.Address{
									Address: &core.Address_SocketAddress{
										SocketAddress: &core.SocketAddress{
											Protocol: core.SocketAddress_TCP,
											Address:  UpstreamHost,
											PortSpecifier: &core.SocketAddress_PortValue{
												PortValue: uint32(50055),
											},
										},
									},
								},
							},
						},
					},
				},
			},
			{
				Locality: &core.Locality{
					Region:  "FailOver",
					Zone:    "local2",
					SubZone: "local2",
				},
				Priority: uint32(1), // Lower priority
				LbEndpoints: []*endpoint.LbEndpoint{
					{
						HealthStatus: core.HealthStatus_HEALTHY,
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: &core.Address{
									Address: &core.Address_SocketAddress{
										SocketAddress: &core.SocketAddress{
											Protocol: core.SocketAddress_TCP,
											Address:  "127.0.0.1",
											PortSpecifier: &core.SocketAddress_PortValue{
												PortValue: uint32(50057),
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}
```