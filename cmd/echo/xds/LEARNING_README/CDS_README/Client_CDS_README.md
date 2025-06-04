# Overview
Cluster is the 3rd layer or config Discovery Requests that is linked and referenced by [RDS](/cmd/echo/xds/LEARNING_README/RDS_README/Client_RDS_README.md) and [EDS](/cmd/echo/xds/LEARNING_README/EDS_README/Client_EDS_README.md).  
The significant configuration for cluster resource is the UpstreamTlsContext (backends) TLS cert setup for the data plane specifically for mTLS.  

## References
[CDS RFC](https://github.com/grpc/proposal/blob/master/A27-xds-global-load-balancing.md#cds)  
[Security RFC](https://github.com/grpc/proposal/blob/master/A29-xds-tls-security.md)  
[Circuit Breaker](https://github.com/grpc/proposal/blob/master/A32-xds-circuit-breaking.md)  

### Client TLS cert configuration
![grpc client bootstrap](/cmd/echo/xds/LEARNING_README/CDS_README/client_bootstrap.png "client bootstrap")


### Grpc Client specific cds
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
		Name: "cluster",
		//  Circuit breaker (A32), only max request currently
		//  https://github.com/grpc/proposal/blob/master/A32-xds-circuit-breaking.md 
		CircuitBreakers: &cluster.CircuitBreakers{
			Thresholds: []*cluster.CircuitBreakers_Thresholds{{
				MaxRequests: &wrapperspb.UInt32Value{Value: 1000},
			}},
		},
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