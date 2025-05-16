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

package resources

import (
	"strings"
	"time"

	"google.golang.org/protobuf/encoding/prototext"
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
	// tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"

	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

const (
	ClusterName         = "backend_cluster"
	RouteName           = "local_route"
	GrpcClientListener  = "connect.me.to.grpcserver"
	GrpcServer1Listener = "example/resource/127.0.0.1:50051"
	GrpcServer2Listener = "example/resource/127.0.0.1:50053"
	GrpcServer3Listener = "example/resource/127.0.0.1:50055"
	GrpcServer4Listener = "example/resource/127.0.0.1:50057"
	UpstreamHost        = "127.0.0.1"
	UpstreamPortA       = 50051
	UpstreamPortB       = 50053
)

func makeCluster() *cluster.Cluster {
	// tlsManager := &tls.UpstreamTlsContext{
	// 	CommonTlsContext: &tls.CommonTlsContext{
	// 		ValidationContextType: &tls.CommonTlsContext_CombinedValidationContext{
	// 			CombinedValidationContext: &tls.CommonTlsContext_CombinedCertificateValidationContext{
	// 				DefaultValidationContext: &tls.CertificateValidationContext{
	// 					CaCertificateProviderInstance: &tls.CertificateProviderPluginInstance{
	// 						InstanceName: "my_custom_cert_provider",
	// 					},
	// 				},
	// 			},
	// 		},
	// 		TlsCertificateProviderInstance: &tls.CertificateProviderPluginInstance{
	// 			InstanceName: "my_custom_cert_provider",
	// 		},
	// 	},
	// }

	// tlsManagerAsAny, err := anypb.New(tlsManager)
	// if err != nil {
	// 	panic(err)
	// }

	return &cluster.Cluster{
		Name: ClusterName,
		// TransportSocket: &core.TransportSocket{
		// 	Name: "envoy.transport_sockets.tls", // this is required, A29: if a transport_socket name is not envoy.transport_sockets.tls i.e. something we don't recognize, gRPC will NACK an LDS update
		// 	ConfigType: &core.TransportSocket_TypedConfig{
		// 		TypedConfig: tlsManagerAsAny,
		// 	},
		// },
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
				Priority:            uint32(0), //0 is highest and is default
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
				Priority:            uint32(0),
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

func DebugSnapshot(snapshot *cache.Snapshot) string {
	sb := strings.Builder{}

	for t, val := range snapshot.Resources {
		name, _ := cache.GetResponseTypeURL(types.ResponseType(t))
		sb.WriteString(name)
		sb.WriteString("\nVersion: ")
		sb.WriteString(val.Version)
		sb.WriteString("\n===============\n")
		for _, v := range val.Items {
			sb.WriteString(prototext.Format(v.Resource))
			sb.WriteString("----------\n")
		}

		sb.WriteString("\n\n")
	}

	return sb.String()
}

func GenerateSnapshotClientSnapshot(version string, weightA uint32, weightB uint32) *cache.Snapshot {
	snap, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  {makeCluster()},
			resource.EndpointType: {makeEndpoint(weightA, weightB)},
			resource.RouteType:    {makeClientRoute()},
			resource.ListenerType: {makeClientListener()},
		},
	)
	return snap
}
