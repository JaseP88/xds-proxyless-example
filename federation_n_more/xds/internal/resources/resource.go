package xdscp

import (
        "fmt"
        "os"
        "reflect"

        "log/slog"
        cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
        core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
        endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
        listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
        lv2 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
        route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
        router "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/router/v3"
        hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
        "github.com/envoyproxy/go-control-plane/pkg/cache/types"
        resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
        "github.com/envoyproxy/go-control-plane/pkg/wellknown"
        "github.com/golang/protobuf/ptypes"
        "google.golang.org/protobuf/types/known/wrapperspb"

        http "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
        "google.golang.org/protobuf/proto"
        "google.golang.org/protobuf/types/known/anypb"
        durationpb "google.golang.org/protobuf/types/known/durationpb"
)

var RouterHTTPFilter = HTTPFilter("router", &router.Router{})

type NodeData struct {
	NodeId		       string
        Listeners              map[string]*ListenerData
        Routes                 map[string]*RouteData
        Clusters               map[string]*ClusterData
        Endpoints              map[string]*EndpointData
}

type ListenerData struct {
        ListenerName		string
        RouteName		string
}

type RouteData struct {
        RouteName       	string
        WeightedCluster 	bool
        Cluster         	map[string]uint32
        Domain          	string
}

type ClusterData struct {
        ClusterName		string
        EndpointServiceName	string
}

type EndpointData struct {
        EndpointServiceName    string
        Locality               map[string]*LocalityData
}

type LocalityData struct {
        Weight                 uint32
        Backends               map[string]*BackendData
        Priority               uint32
        Locations              *LocationData
}

type BackendData struct {
	Host			string	
        Ports                  	[]uint32
	AdditionalPorts        	[]uint32
        HealthStatus           	core.HealthStatus
        Weight                 	uint32
        Metadata               	map[string]any
}

type LocationData struct {
	Name			string
        Region                 	string
        Zone                   	string
        SubZone                	string
}

type UpdateOptions struct {
        NodeID string
        Endpoints []*endpoint.ClusterLoadAssignment
        Clusters  []*cluster.Cluster
        Routes    []*route.RouteConfiguration
        Listeners []*listener.Listener
}

func marshalAny(m proto.Message) *anypb.Any {
        a, err := anypb.New(m)
        if err != nil {
                panic(fmt.Sprintf("anypb.New(%+v) failed: %v", m, err))
        }
        return a
}

func resourceSlice(i any) []types.Resource {
        v := reflect.ValueOf(i)
        rs := make([]types.Resource, v.Len())
        for i := 0; i < v.Len(); i++ {
                rs[i] = v.Index(i).Interface().(types.Resource)
        }
        return rs
}

func HTTPFilter(name string, config proto.Message) *http.HttpFilter {
        return &http.HttpFilter{
                Name: name,
                ConfigType: &http.HttpFilter_TypedConfig{
                        TypedConfig: marshalAny(config),
                },
        }
}

func makeConfigSource() *core.ConfigSource {
        source := &core.ConfigSource{}
        source.ResourceApiVersion = core.ApiVersion_V3
        source.ConfigSourceSpecifier = &core.ConfigSource_Ads{
                Ads: &core.AggregatedConfigSource{},
        }

        return source
}



func defaultServerListenerCommon(listenerName, host string, port uint32, routeName string, inlineRouteConfig bool) *listener.Listener {
        var hcm *http.HttpConnectionManager
        ptypeRouterConfig, err := ptypes.MarshalAny(&router.Router{})

        if inlineRouteConfig {
                hcm = &http.HttpConnectionManager{
                        RouteSpecifier: &http.HttpConnectionManager_RouteConfig{
                                RouteConfig: &route.RouteConfiguration{
                                        Name: routeName,
                                        VirtualHosts: []*route.VirtualHost{{
                                                Domains: []string{"*"},
                                 
                                                Routes: []*route.Route{{
                                                        Match: &route.RouteMatch{
                                                                PathSpecifier: &route.RouteMatch_Prefix{Prefix: "/"},
                                                        },
                                                        Action: &route.Route_NonForwardingAction{},
                                                }}}}},
                        },
                        HttpFilters: []*http.HttpFilter{RouterHTTPFilter},
                }
        } else {
                hcm = &http.HttpConnectionManager{
                        RouteSpecifier: &http.HttpConnectionManager_Rds{
                                Rds: &http.Rds{
                                        ConfigSource:    makeConfigSource(),
                                        RouteConfigName: routeName,
                                },
                        },
                        HttpFilters: []*http.HttpFilter{{
                            Name:       wellknown.Router,
                            ConfigType: &http.HttpFilter_TypedConfig{TypedConfig: ptypeRouterConfig},
                        }},
                }
        }

        pbst, err := ptypes.MarshalAny(hcm)
        if err != nil {
                panic(err)
        }

        var ts *core.TransportSocket

        return &listener.Listener{
                Name: listenerName,
                Address: &core.Address{
                        Address: &core.Address_SocketAddress{
                                SocketAddress: &core.SocketAddress{
                                        Address: host,
                                        PortSpecifier: &core.SocketAddress_PortValue{
                                                PortValue: port,
                                        },
                                },
                        },
                },
                FilterChains: []*listener.FilterChain{
                        {
                                Name: "v4-wildcard",
                                FilterChainMatch: &listener.FilterChainMatch{
                                        PrefixRanges: []*core.CidrRange{
                                                {
                                                        AddressPrefix: "0.0.0.0",
                                                        PrefixLen: &wrapperspb.UInt32Value{
                                                                Value: uint32(0),
                                                        },
                                                },
                                        },
                                        SourceType: listener.FilterChainMatch_SAME_IP_OR_LOOPBACK,
                                        SourcePrefixRanges: []*core.CidrRange{
                                                {
                                                        AddressPrefix: "0.0.0.0",  
                                                        PrefixLen: &wrapperspb.UInt32Value{
                                                                Value: uint32(0),
                                                        },
                                                },
                                        },
                                },
                                Filters: []*listener.Filter{
                                        {
                                                Name:       "filter-1",
                                                ConfigType: &listener.Filter_TypedConfig{TypedConfig: marshalAny(hcm)},
                                        },
                                },
                                TransportSocket: ts,
                        },
                        {
                                Name: "v6-wildcard",
                                FilterChainMatch: &listener.FilterChainMatch{
                                        PrefixRanges: []*core.CidrRange{
                                                {
                                                        AddressPrefix: "::",
                                                        PrefixLen: &wrapperspb.UInt32Value{
                                                                Value: uint32(0),
                                                        },
                                                },
                                        },
                                        SourceType: listener.FilterChainMatch_SAME_IP_OR_LOOPBACK,
                                        SourcePrefixRanges: []*core.CidrRange{
                                                {
                                                        AddressPrefix: "::",
                                                        PrefixLen: &wrapperspb.UInt32Value{
                                                                Value: uint32(0),
                                                        },
                                                },
                                        },
                                },
                                Filters: []*listener.Filter{
                                        {
                                                Name:       "filter-1",
                                                ConfigType: &listener.Filter_TypedConfig{TypedConfig: marshalAny(hcm)},
                                        },
                                },
                                TransportSocket: ts,
                        },
                        {
                                Name: "wellknown-filter",
                                Filters: []*listener.Filter{{
                                        Name: wellknown.HTTPConnectionManager,
                                        ConfigType: &listener.Filter_TypedConfig{
                                                TypedConfig: pbst,
                                        },
                               }},
                        },

                },
        }
}

func makeHttpListener(listenerName, routeConfigName string) *listener.Listener {
        ptypeRouterConfig, err := ptypes.MarshalAny(&router.Router{})
        if err != nil {
                slog.Error("Failed to Marshal router: %v\n", err)
                os.Exit(1)
        }

        manager := &hcm.HttpConnectionManager{
                CodecType: hcm.HttpConnectionManager_AUTO,
                RouteSpecifier: &hcm.HttpConnectionManager_Rds{
                        Rds: &hcm.Rds{
                                ConfigSource:    makeConfigSource(),
                                RouteConfigName: routeConfigName,
                        },
                },
                HttpFilters: []*hcm.HttpFilter{{
                        Name:       wellknown.Router,
                        ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: ptypeRouterConfig},
                }},
        }

        ptypeMgr, err := ptypes.MarshalAny(manager)
        if err != nil {
                slog.Error("Failed to Marshal HttpConnectionManager: %v\n", err)
                os.Exit(1)
        }

        return &listener.Listener{
                Name: listenerName,
                ApiListener: &lv2.ApiListener{
                        ApiListener: ptypeMgr,
                },
        }

}

func makeRoute(routeData *RouteData) *route.RouteConfiguration {
        var rts []*route.Route
	var wclusters []*route.WeightedCluster_ClusterWeight

	if len(routeData.Cluster) <= 0 {
		slog.Error("Failed to make route, missing cluster name")
		return nil
	}
	hostname, err := os.Hostname()
	if err != nil {
		slog.Error("Failed to get hostname, while building route", routeData.RouteName)
		hostname = "Unknown"
	}

       if len(routeData.Cluster) > 1 {
               for clusterName, weight := range routeData.Cluster {
		      wclusters = append(wclusters, &route.WeightedCluster_ClusterWeight{Name: clusterName, Weight: &wrapperspb.UInt32Value{Value: uint32(weight)}})
	       }

	       rts = append(rts, &route.Route{
                        Match: &route.RouteMatch{
                                PathSpecifier: &route.RouteMatch_Prefix{
                                        Prefix: "",
                                },
                        },
                        Action: &route.Route_Route{
                                Route: &route.RouteAction{
					ClusterSpecifier: &route.RouteAction_WeightedClusters{
						WeightedClusters: &route.WeightedCluster{
							Clusters: wclusters,
						},
					 },
                                },
                        },
                })
        } else {
                for clusterName, _ := range routeData.Cluster {
			rts = append(rts, &route.Route{
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "",
					},
				},
				Action: &route.Route_Route{
					Route: &route.RouteAction{
						ClusterSpecifier: &route.RouteAction_Cluster{
							Cluster: clusterName,
						},
					},
				},
			})
			break
		}
        }



        return &route.RouteConfiguration{
                Name:             routeData.RouteName,
                ValidateClusters: &wrapperspb.BoolValue{Value: true},
                VirtualHosts: []*route.VirtualHost{{
			Name:    hostname,
			Domains: []string{routeData.Domain},
                        Routes:  rts,
                }},
        }
}

func makeCluster(clusterName, edsServiceName string) *cluster.Cluster {
        return &cluster.Cluster{
                Name:                 clusterName,
                ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_EDS},
                LbPolicy:             cluster.Cluster_ROUND_ROBIN,
                HealthChecks: []*core.HealthCheck{&core.HealthCheck{Timeout: &durationpb.Duration{Seconds: 5} , Interval: &durationpb.Duration{Seconds: 10}}},
                EdsClusterConfig: &cluster.Cluster_EdsClusterConfig{
                                                                       EdsConfig: makeConfigSource(),
                                                                       ServiceName: edsServiceName,
                                                                   },
        }
}

func makeEndpoint(edsData *EndpointData) *endpoint.ClusterLoadAssignment {

	var endpoints []*endpoint.LocalityLbEndpoints

	for _, locality := range edsData.Locality {
		var lbEndpoints []*endpoint.LbEndpoint
		for _, b := range locality.Backends {
			additionalAddresses := make([]*endpoint.Endpoint_AdditionalAddress, len(b.AdditionalPorts))
			for i, p := range b.AdditionalPorts {
				additionalAddresses[i] = &endpoint.Endpoint_AdditionalAddress{
					Address: &core.Address{Address: &core.Address_SocketAddress{
						SocketAddress: &core.SocketAddress{
							Protocol:      core.SocketAddress_TCP,
							Address:       b.Host,
							PortSpecifier: &core.SocketAddress_PortValue{PortValue: p},
						}},
					},
				}
			}
		    	for _, p := range b.Ports {
				lbEndpoints = append(lbEndpoints, &endpoint.LbEndpoint{
					HostIdentifier: &endpoint.LbEndpoint_Endpoint{Endpoint: &endpoint.Endpoint{
						Address: &core.Address{Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Protocol:      core.SocketAddress_TCP,
								Address:       b.Host,
								//PortSpecifier: &core.SocketAddress_PortValue{PortValue: b.Ports[0]},
								PortSpecifier: &core.SocketAddress_PortValue{PortValue: p},
							},
						}},
						AdditionalAddresses: additionalAddresses,
					}},
					HealthStatus:        b.HealthStatus,
					LoadBalancingWeight: &wrapperspb.UInt32Value{Value: b.Weight},
				})
			}
		}

		location := locality.Locations

		endpoints = append(endpoints, &endpoint.LocalityLbEndpoints{
			 Locality:            &core.Locality{Region: location.Region, Zone: location.Zone, SubZone: location.SubZone},
			 LbEndpoints:         lbEndpoints,
			 LoadBalancingWeight: &wrapperspb.UInt32Value{Value: locality.Weight},
			 Priority:            locality.Priority,
		 })
	 }

	 cla := &endpoint.ClusterLoadAssignment{
		 ClusterName: edsData.EndpointServiceName,
		 Endpoints:   endpoints,
	 }
	 return cla
}

func DefaultUpstreamResources(lName, host string, port uint32) map[resource.Type] []types.Resource {

        return getResources(UpdateOptions{
                                Listeners: []*listener.Listener{defaultServerListenerCommon(lName, host, port, "no-route", true)},
                            })
}

func getResources(opts UpdateOptions) map[resource.Type] []types.Resource {

        return map[resource.Type][]types.Resource{
                   resource.ListenerType: resourceSlice(opts.Listeners),
                   resource.RouteType:    resourceSlice(opts.Routes),
                   resource.ClusterType:  resourceSlice(opts.Clusters),
                   resource.EndpointType: resourceSlice(opts.Endpoints),
        }
}


func CreateResources(resourceData *NodeData) map[resource.Type] []types.Resource {
	var listeners []*listener.Listener
        var routes []*route.RouteConfiguration
	var clusters []*cluster.Cluster
	var endpoints []*endpoint.ClusterLoadAssignment

	for _, alistener := range resourceData.Listeners {
		listeners = append(listeners, makeHttpListener(alistener.ListenerName, alistener.RouteName))
	}

	for _, aroute := range resourceData.Routes {
	        routes = append(routes, makeRoute(aroute))
	}

	for _, acluster := range resourceData.Clusters {
		clusters = append(clusters, makeCluster(acluster.ClusterName, acluster.EndpointServiceName))
	}
	
	for _, endpoint := range resourceData.Endpoints {
		endpoints = append(endpoints, makeEndpoint(endpoint))
	}
	return getResources(UpdateOptions{
		Listeners: listeners,
		Routes   : routes,
		Clusters : clusters,
		Endpoints: endpoints,
	})
}
