// Copyright 2020 Envoyproxy Authors
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

package wc_resources

import (
	"time"

	"github.com/JaseP88/xds-poc/cmd/greet/xds/internal/xds_resources"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

func makeClientListener() *listener.Listener {
	routerConfig, _ := anypb.New(&router.Router{})

	httpConnectionManager := &hcm.HttpConnectionManager{
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource: &core.ConfigSource{
					ConfigSourceSpecifier: &core.ConfigSource_Ads{
						Ads: &core.AggregatedConfigSource{},
					},
				},
				RouteConfigName: resources.RouteName,
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
		Name: resources.GrpcClientListener,
		ApiListener: &listener.ApiListener{
			ApiListener: httpConnectionManagerAsAny,
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


func makeClientRoute() *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: resources.RouteName,
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
						ClusterSpecifier: &route.RouteAction_WeightedClusters{
							WeightedClusters: &route.WeightedCluster{
								Clusters: []*route.WeightedCluster_ClusterWeight{
									{
										Name:   resources.ClusterA,
										Weight: &wrapperspb.UInt32Value{Value: uint32(70)},
									},
									{
										Name:   resources.ClusterB,
										Weight: &wrapperspb.UInt32Value{Value: uint32(30)},
									},
								},
							},
						},
					},
				},
			}},
		}},
	}
}

func makeClusterA() *cluster.Cluster {
	tlsManager := &tls.UpstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			ValidationContextType: &tls.CommonTlsContext_CombinedValidationContext{
				CombinedValidationContext: &tls.CommonTlsContext_CombinedCertificateValidationContext{
					DefaultValidationContext: &tls.CertificateValidationContext{
						CaCertificateProviderInstance: &tls.CertificateProviderPluginInstance{
							InstanceName: "my_custom_cert_provider",
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

	return &cluster.Cluster{
		Name: resources.ClusterA,
		TransportSocket: &core.TransportSocket{
			Name: "envoy.transport_sockets.tls", // this is required, A29: if a transport_socket name is not envoy.transport_sockets.tls i.e. something we don't recognize, gRPC will NACK an LDS update
			ConfigType: &core.TransportSocket_TypedConfig{
				TypedConfig: tlsManagerAsAny,
			},
		},
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_Ads{
					Ads: &core.AggregatedConfigSource{},
				},
			},
		},
		LbPolicy:        cluster.Cluster_ROUND_ROBIN,
		DnsLookupFamily: cluster.Cluster_V4_ONLY,
	}
}

func makeClusterB() *cluster.Cluster {
	tlsManager := &tls.UpstreamTlsContext{
		CommonTlsContext: &tls.CommonTlsContext{
			ValidationContextType: &tls.CommonTlsContext_CombinedValidationContext{
				CombinedValidationContext: &tls.CommonTlsContext_CombinedCertificateValidationContext{
					DefaultValidationContext: &tls.CertificateValidationContext{
						CaCertificateProviderInstance: &tls.CertificateProviderPluginInstance{
							InstanceName: "my_custom_cert_provider",
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

	return &cluster.Cluster{
		Name: resources.ClusterB,
		TransportSocket: &core.TransportSocket{
			Name: "envoy.transport_sockets.tls", // this is required, A29: if a transport_socket name is not envoy.transport_sockets.tls i.e. something we don't recognize, gRPC will NACK an LDS update
			ConfigType: &core.TransportSocket_TypedConfig{
				TypedConfig: tlsManagerAsAny,
			},
		},
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
		EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
			EdsConfig: &core.ConfigSource{
				ConfigSourceSpecifier: &core.ConfigSource_Ads{
					Ads: &core.AggregatedConfigSource{},
				},
			},
		},
		LbPolicy:        cluster.Cluster_ROUND_ROBIN,
		DnsLookupFamily: cluster.Cluster_V4_ONLY,
	}
}

func makeEndpointA() *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: resources.ClusterA,
		Endpoints: []*endpoint.LocalityLbEndpoints{
			{
				Locality: &core.Locality{
					Region:  "Region1",
					Zone:    "local1",
					SubZone: "local1",
				},
				Priority: uint32(0), //0 is highest and is default
				LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 100},
				LbEndpoints: []*endpoint.LbEndpoint{
					{
						HealthStatus: core.HealthStatus_HEALTHY,
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: &core.Address{
									Address: &core.Address_SocketAddress{
										SocketAddress: &core.SocketAddress{
											Protocol: core.SocketAddress_TCP,
											Address:  resources.UpstreamHost,
											PortSpecifier: &core.SocketAddress_PortValue{
												PortValue: resources.UpstreamPort_50051,
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
											Address:  resources.UpstreamHost,
											PortSpecifier: &core.SocketAddress_PortValue{
												PortValue: resources.UpstreamPort_50053,
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

func makeEndpointB() *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: resources.ClusterB,
		Endpoints: []*endpoint.LocalityLbEndpoints{
			{
				Locality: &core.Locality{
					Region:  "Region2",
					Zone:    "local2",
					SubZone: "local2",
				},
				Priority:            uint32(0),
				LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 100},
				LbEndpoints: []*endpoint.LbEndpoint{
					{
						HealthStatus: core.HealthStatus_HEALTHY,
						HostIdentifier: &endpoint.LbEndpoint_Endpoint{
							Endpoint: &endpoint.Endpoint{
								Address: &core.Address{
									Address: &core.Address_SocketAddress{
										SocketAddress: &core.SocketAddress{
											Protocol: core.SocketAddress_TCP,
											Address:  resources.UpstreamHost,
											PortSpecifier: &core.SocketAddress_PortValue{
												PortValue: resources.UpstreamPort_50055,
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
											Address:  resources.UpstreamHost,
											PortSpecifier: &core.SocketAddress_PortValue{
												PortValue: resources.UpstreamPort_50057,
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

func GenerateSnapshotClientSnapshot(version string, weightA uint32, weightB uint32) *cache.Snapshot {
	snap, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  {makeClusterA(), makeClusterB()},
			resource.EndpointType: {makeEndpointA(), makeEndpointB()},
			resource.RouteType:    {makeClientRoute()},
			resource.ListenerType: {makeClientListener()},
		},
	)
	return snap
}
