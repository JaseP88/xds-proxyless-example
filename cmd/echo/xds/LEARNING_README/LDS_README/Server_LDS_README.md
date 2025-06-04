[Server xDS Configurations](https://github.com/grpc/proposal/blob/master/A36-xds-for-servers.md).  
For the server configuration for LDS is a bit different.  
There is a TLS Downstream Context that is configured for TLS or mTLS.  

### Grpc Server bootstrap
![grpc server bootstrap](/cmd/echo/xds/LEARNING_README/LDS_README/server_bootstrap.png "server bootstrap")

### Grpc Server specific LDS 
```go
func makeServerListener() *listener.Listener {
	routerConfig, _ := anypb.New(&router.Router{})

	httpConnectionManager := &hcm.HttpConnectionManager{
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{ // use RDS
			Rds: &hcm.Rds{
				ConfigSource: &core.ConfigSource{
					ConfigSourceSpecifier: &core.ConfigSource_Ads{
						Ads: &core.AggregatedConfigSource{},
					},
				},
				RouteConfigName: "local_route",
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
		Name: "this.is.the.lds",
		Address: &core.Address{ // A36 To be useful, the xDS-returned Listener must have an address that matches the listening address provided.
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  "127.0.0.1",
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: 50051,
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
