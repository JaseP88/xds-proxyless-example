This package contains an xDS server that has configurations for `grpc clients` and `grpc servers`.  
Current configuration is a *single* cluster route with TLS contexts & a weighted balance of 90% to a grpc server and 10% to another, and one more configuration for priority / failover grpc server.  


# Learning xDS
## Client configurations
### REFS
- https://github.com/grpc/proposal/blob/master/A27-xds-global-load-balancing.md

### client specific LDS
```go
func makeClientListener() *listener.Listener {
	routerConfig, _ := anypb.New(&router.Router{})

	httpConnectionManager := &hcm.HttpConnectionManager{
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource: &core.ConfigSource{
					ConfigSourceSpecifier: &core.ConfigSource_Ads{ // tells control plane to use RDS in ADS
						Ads: &core.AggregatedConfigSource{},
					},
				},
				RouteConfigName: RouteName,
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name:       "http-router",
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
		}},
	}

	httpConnectionManagerAsAny, err := anypb.New(httpConnectionManager)
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: GrpcClientListener,
		ApiListener: &listener.ApiListener{
			ApiListener: httpConnectionManagerAsAny, // grpc only support APIListener
		},
		FilterChains: []*listener.FilterChain{{
			Name: "filter-chain",
			Filters: []*listener.Filter{{
				Name:       wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{TypedConfig: httpConnectionManagerAsAny},
			}},
		}},
	}
}
```

### client specific RDS single cluster routing
```go
func makeClientRoute() *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: RouteName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    "VH",
			Domains: []string{"*"},
			Routes: []*route.Route{{
				Name: "http-router",
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: ClusterName,
						},
					},
				},
			}},
		}},
	}
}
```

### client specific RDS multi cluster routing
```go
func makeClientRoute() *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: RouteName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    "VH",
			Domains: []string{"*"},
			Routes: []*route.Route{{
				Name: "http-router",
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Path{
						Path: "/domain.MyService/UnaryReq",  // full name path of grpc service proto, /package.grpcservice/rpc
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: ClusterA,
						},
					},
				},
			},
			{
				Name: "http-router",
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Path{
						Path: "/domain2.AnotherService/DifferentRPC", // full name path of grpc service proto
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: ClusterB,
						},
					},
				},
			},
		},
		}},
	}
}
```

```go
func GenerateSnapshotClientSnapshot(version string, weightA uint32, weightB uint32) *cache.Snapshot {
	snap, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  {makeClusterA(), makeClusterB()}, // 2 resources
			resource.EndpointType: {makeEndpoint(weightA, weightB), makeEndpointB(weightA, weightB)}, // 2 resources
			resource.RouteType:    {makeClientRoute()},
			resource.ListenerType: {makeClientListener()},
		},
	)
	return snap
}
```

### client specific cds
```go
func makeCluster() *cluster.Cluster {
	tlsManager := &tls.UpstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			ValidationContextType: &tls.CommonTlsContext_CombinedValidationContext{
				CombinedValidationContext: &tls.CommonTlsContext_CombinedCertificateValidationContext{
					DefaultValidationContext: &tls.CertificateValidationContext{
						CaCertificateProviderInstance: &tls.CertificateProviderPluginInstance{
							InstanceName: "my_custom_cert_provider", // this should match what is in the clients bootstrap json
						},
					},
				},
			},
			TlsCertificateProviderInstance: &tls.CertificateProviderPluginInstance{
				InstanceName: "my_custom_cert_provider", // this should match what is in the clients bootstrap json
			},
		},
	}

	tlsManagerAsAny, err := anypb.New(tlsManager)
	if err != nil {
		panic(err)
	}

	return &cluster.Cluster{
		Name:                 ClusterName,
		TransportSocket: &core.TransportSocket{
			Name: "envoy.transport_sockets.tls", // this is required, A29: if a transport_socket name is not envoy.transport_sockets.tls i.e. something we don't recognize, gRPC will NACK an LDS update
			ConfigType: &core.TransportSocket_TypedConfig{
				TypedConfig: tlsManagerAsAny,
			},
		},
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},  // tells control plane to use EDS
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_Ads{  // tells control plane to do through ADS
					Ads: &core.AggregatedConfigSource{},
				},
			},
		},
		LbPolicy:        cluster.Cluster_ROUND_ROBIN,
		DnsLookupFamily: cluster.Cluster_V4_ONLY,
	}
}
```

### client specific EDS
```go
func makeEndpoint(weightA uint32, weightB uint32) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: ClusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{
			{
				Locality: &core.Locality{
					Region:  "Region1",
					Zone:    "local1",
					SubZone: "local1",
				},
				Priority: uint32(0), //0 is highest and is default
				LoadBalancingWeight: &wrapperspb.UInt32Value{Value: weightA},
				LbEndpoints: []*endpoint.LbEndpoint{
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
												PortValue: UpstreamPortA,
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
				LoadBalancingWeight: &wrapperspb.UInt32Value{Value: weightB},
				LbEndpoints: []*endpoint.LbEndpoint{
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
												PortValue: UpstreamPortB,
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
				Priority: uint32(1),
				LbEndpoints: []*endpoint.LbEndpoint{
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


### server specific LDS 
```go
func makeServerListener(port uint32) *listener.Listener {
	routerConfig, _ := anypb.New(&router.Router{})

	httpConnectionManager := &hcm.HttpConnectionManager{
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{ // use RDS
			Rds: &hcm.Rds{
				ConfigSource: &core.ConfigSource{
					ConfigSourceSpecifier: &core.ConfigSource_Ads{
						Ads: &core.AggregatedConfigSource{},
					},
				},
				RouteConfigName: RouteName,
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name:       "http-router",
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
		}},
	}

	httpConnectionManagerAsAny, err := anypb.New(httpConnectionManager)
	if err != nil {
		panic(err)
	}

	tlsManager := &tls.DownstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			ValidationContextType: &tls.CommonTlsContext_CombinedValidationContext{
				CombinedValidationContext: &tls.CommonTlsContext_CombinedCertificateValidationContext{
					DefaultValidationContext: &tls.CertificateValidationContext{
						CaCertificateProviderInstance: &tls.CertificateProviderPluginInstance{
							InstanceName: "my_custom_cert_provider", // must match what is in the bootstrap json
						},
					},
				},
			},
			TlsCertificateProviderInstance: &tls.CertificateProviderPluginInstance{
				InstanceName: "my_custom_cert_provider",
			},
		},
	}

	tlsManagerAsAny, err := anypb.New(tlsManager)
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: listenerName,
		// A39: if the HttpConnectionManager proto is inside an HTTP API Listener, it will look only at filters registered for the gRPC client, whereas if it is inside a TCP Listener, it will look only at filters registered for the gRPC server.
		// ApiListener: &listener.ApiListener{
		// 	ApiListener: httpConnectionManagerAsAny,
		// },
		Address: &core.Address{ // A36 To be useful, the xDS-returned Listener must have an address that matches the listening address provided.
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  UpstreamHost,
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: port,
					},
				},
			},
		},
		FilterChains: []*listener.FilterChain{{
			Name: "filter-chain",
			TransportSocket: &core.TransportSocket{
				Name: "envoy.transport_sockets.tls", // this is required, A29: if a transport_socket name is not envoy.transport_sockets.tls i.e. something we don't recognize, gRPC will NACK an LDS update
				ConfigType: &core.TransportSocket_TypedConfig{
					TypedConfig: tlsManagerAsAny,
				},
			},
			Filters: []*listener.Filter{{
				Name:       wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{TypedConfig: httpConnectionManagerAsAny},
			}},
		}},
	}
}
```

### server specific RDS
```go
func makeServerRoute() *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: RouteName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    "VH",
			Domains: []string{"*"},
			Routes: []*route.Route{{
				Name: "http-router",
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &route.Route_NonForwardingAction{
					NonForwardingAction: &route.NonForwardingAction{
						// A36: Route.non_forwarding_action is expected for all Routes used on server-side and Route.route continues to be expected for all Routes used on client-side
					},
				},
			}},
		}},
	}
}
```