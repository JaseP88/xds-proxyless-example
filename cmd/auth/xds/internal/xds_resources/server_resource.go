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
	"google.golang.org/protobuf/types/known/anypb"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
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

func makeServerListener(port uint32) *listener.Listener {
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

	tlsManager := &tls.DownstreamTlsContext{
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

	var listenerName string
	switch port {
	case 50051:
		listenerName = GrpcServer1Listener
	case 50053:
		listenerName = GrpcServer2Listener
	case 50055:
		listenerName = GrpcServer3Listener
	case 50057:
		listenerName = GrpcServer4Listener
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

func GenerateSnapshotServerSnapshot(version string, port uint32) *cache.Snapshot {
	snap, _ := cache.NewSnapshot(version,
		map[resource.Type][]types.Resource{
			resource.ClusterType:  {},
			resource.EndpointType: {},
			resource.RouteType:    {makeServerRoute()},
			resource.ListenerType: {makeServerListener(port)},
		},
	)
	return snap
}