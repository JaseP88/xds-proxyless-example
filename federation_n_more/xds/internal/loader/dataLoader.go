package xdscp

import (
	"slices"
        "strings"
        "strconv"
	"log/slog"

	cpRresource "xdsmod/internal/resources"

	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
)

var (
        domain 		= "*.xdspoc.com"
	federation 	= "xdstp://"
)


type RequestData struct {
	UpstreamNodeId	string
	UpstreamWeight	uint32
	ResourceName	string
	Host		string
	Port		uint32
	ServiceId	string
	BackupFlag	bool
	ListenerName	string
	Cluster	        string 
	ClsWeight	uint32
	Region		string
	Zone		string
	SubZone		string
	Priority	uint32
	LbWeight	uint32
	EndpointName	string
}



func MapFileDataToNodeConfig(nodeData *cpRresource.NodeData, configData *ConfigData, serviceName string) bool  {
	var serviceClusters []string
        var ret bool
	upsMap, _ := configData.Service[serviceName]
	for upsName, routeDetail := range upsMap{
			var routeClusterMap map[string]uint32 = map[string]uint32{}
			for _, clsDetail := range routeDetail.Clusters {
				serviceClusters = append(serviceClusters, getClusterName(clsDetail.Name))

				if clsDetail.Weight != 0 {
					routeClusterMap[getClusterName(clsDetail.Name)] = clsDetail.Weight
				} else {
					routeClusterMap[getClusterName(clsDetail.Name)] = 1
				}
					if isFederated(clsDetail.FederatedEdsName) {
						ret = addCluster(nodeData, getClusterName(clsDetail.Name), clsDetail.FederatedEdsName) || ret
					} else {

						ret = addCluster(nodeData, getClusterName(clsDetail.Name), getEndpointName(clsDetail.Name)) || ret
					}
			}
			listenerName := getListenerName(upsName, domain)
			routeName := getRouteName(routeDetail.Name, listenerName)
			ret = addListener(nodeData, listenerName, routeName) || ret
			ret = addRoute(nodeData, routeName, domain, routeClusterMap) || ret
	}

	for clusterName, localityarr := range configData.ClusterLoadAssignment {
		if !(slices.Contains(serviceClusters, getClusterName(clusterName))) {
			continue
		}

		for _, localityDetail := range localityarr {
			var lweight uint32
			if localityDetail.Weight == 0 {
				lweight = 1
			} else {
				lweight = localityDetail.Weight
			}

			for _, endpoint := range localityDetail.Endpoints {
				var uweight uint32
				if endpoint.Weight == 0 {
					uweight = 1
				} else { 
					uweight = endpoint.Weight
				}

				for _, port := range endpoint.Port {
					ret = addEndpoint(nodeData, getEndpointName(clusterName), endpoint.Host, localityDetail.Region, localityDetail.Zone, localityDetail.SubZone, port, uweight, lweight, localityDetail.Priority) || ret
				}
			}

		}
	}

	return ret
}

func GetUpstreamId(request RequestData) string {
	return request.ResourceName
}

func GetServiceId(request RequestData) string {
	return request.ServiceId
}


func CreateUpstreamData(request RequestData) *cpRresource.NodeData {
	var nodeRecord cpRresource.NodeData
	nodeRecord.NodeId = request.UpstreamNodeId
	nodeRecord.Listeners = map[string]*cpRresource.ListenerData{
		request.ResourceName: &cpRresource.ListenerData{ListenerName: request.ResourceName},
	}

	return &nodeRecord
}

func CreateServiceData(request RequestData) *cpRresource.NodeData {
	nodeRecord := &cpRresource.NodeData{NodeId: request.ServiceId, Listeners: map[string]*cpRresource.ListenerData{}, Routes: map[string]*cpRresource.RouteData{}, Clusters: map[string]*cpRresource.ClusterData{}, Endpoints: map[string]*cpRresource.EndpointData{}}

	listenerName := getListenerName(request.ListenerName, domain)
	addListener(nodeRecord, listenerName, getRouteName("", listenerName))
	addRoute(nodeRecord, getRouteName("", listenerName), domain, map[string]uint32{request.Cluster: uint32(request.ClsWeight)})
	addCluster(nodeRecord, getClusterName(request.Cluster), getEndpointName(request.Cluster))

	endpointName := request.Cluster
        if isFederated(request.EndpointName) {
                endpointName = request.EndpointName
        }
	addEndpoint(nodeRecord, getEndpointName(endpointName), request.Host, request.Region, request.Zone, request.SubZone, request.Port, request.UpstreamWeight, request.LbWeight, request.Priority)

	return nodeRecord
}

func UpdateServiceData(request RequestData, nodeRecord *cpRresource.NodeData) bool {
	var ret bool
	listenerName := getListenerName(request.ListenerName, domain)
	routeName := getRouteName("", listenerName)
	ret = addListener(nodeRecord, listenerName, routeName) 
	ret = addRoute(nodeRecord, routeName, domain, map[string]uint32{request.Cluster: uint32(request.ClsWeight)}) || ret
	ret = addCluster(nodeRecord, getClusterName(request.Cluster), getEndpointName(request.Cluster)) || ret

	if request.LbWeight == 0 {
		request.LbWeight = 1
	}
	if request.UpstreamWeight == 0 {
		request.UpstreamWeight = 1
	}

	endpointName := request.Cluster
	if isFederated(request.EndpointName) {
		endpointName = request.EndpointName
	}
	ret = addEndpoint(nodeRecord, getEndpointName(endpointName), request.Host, request.Region, request.Zone, request.SubZone, request.Port, request.UpstreamWeight, request.LbWeight, request.Priority) || ret

	return ret
}

func getListenerName(name, domain string) string {
	if len(name) == 0 {
		return "default-listener." + domain[1:]
	}
	if isFederated(name) {
		return name
	}
	return name + domain[1:]
}

func getRouteName(name, upstreamName string) string {
        if len(name) == 0 {
		return "default-route-" + upstreamName
        }
	if isFederated(name) {
                return name
        }
	return name + "-" + upstreamName
}

func getClusterName(name string) string {
        if len(name) == 0 {
                return "default-cluster"
        }
	if isFederated(name) {
                return name
        }
        return name
}

func getEndpointName(name string) string {
	if len(name) == 0 {
		return "default-endpoint" + getClusterName("")
	}
	if isFederated(name) {
                return name
        }
	return "endpoint-" + getClusterName(name)
}

func getLocalityName(region, zone, subzone, priority string) string {
        return "Eds-" + getRegion(region) + "-" + getZone(zone) + "-" + getSubZone(subzone) + "-" + priority
}

func getRegion(name string) string {
	if len(name) == 0 {
		name = "d-region"
	}
	return name 
}

func getZone(name string) string {
        if len(name) == 0 {
                name = "d-subZone"
        }
        return name
}

func getSubZone(name string) string {
        if len(name) == 0 {
                name = "d-subZone"
        }
        return name
}

func addEndpoint(nodeRecord *cpRresource.NodeData, edsName, hostName, region, zone, subZone string, port, uweight, lweight, priority uint32) bool {
	localityName := getLocalityName(region, zone, subZone, strconv.Itoa(int(priority)))

	if endpoint, ok := nodeRecord.Endpoints[edsName]; ok {
		if locality, ok := endpoint.Locality[localityName]; ok {
			if bitem, ok := locality.Backends[hostName]; ok {
				if slices.Contains(bitem.Ports, port) {
					slog.Debug("Skip update, Upstream exist", "EDS", edsName, "Host", hostName, "Port", port)
					return false
				} else {
					bitem.Ports = append(bitem.Ports, port)
				}
			} else {
				locality.Backends[hostName] = &cpRresource.BackendData{Host: hostName, Ports: []uint32{port}, HealthStatus: core.HealthStatus(0), Weight: uweight}
			}
		} else {
			endpoint.Locality[localityName] = &cpRresource.LocalityData{Weight: lweight, Backends: map[string]*cpRresource.BackendData{hostName: &cpRresource.BackendData{Host: hostName, Ports: []uint32{port}, HealthStatus: core.HealthStatus(0), Weight: uweight}}, Priority: priority, Locations: &cpRresource.LocationData{Name: localityName, Region: getRegion(region), Zone: getZone(zone), SubZone: getSubZone(subZone)}}
		}
	} else {
		nodeRecord.Endpoints[edsName] = &cpRresource.EndpointData{EndpointServiceName: edsName, Locality: map[string]*cpRresource.LocalityData{localityName: &cpRresource.LocalityData{Weight: lweight, Backends: map[string]*cpRresource.BackendData{hostName: &cpRresource.BackendData{Host: hostName, Ports: []uint32{port}, HealthStatus: core.HealthStatus(0), Weight: uweight}}, Priority: priority, Locations:&cpRresource.LocationData{Name: localityName, Region: getRegion(region), Zone: getZone(zone), SubZone: getSubZone(subZone)}}}}
	}
	return true
}

func addCluster(nodeRecord *cpRresource.NodeData, clusterName, edsName string) bool {
	if _, exist := nodeRecord.Clusters[clusterName]; exist {
		slog.Debug("Skip cluster creation, cluster exist", "cluster:", clusterName)
		return false
	}
	nodeRecord.Clusters[clusterName] = &cpRresource.ClusterData{ClusterName: clusterName, EndpointServiceName: edsName}
	return true
}

func addRoute(nodeRecord *cpRresource.NodeData, routeName, domain string, cluster map[string]uint32) bool {
	if _, exist := nodeRecord.Routes[routeName]; exist {
		for reqClsName, reqClsWeight := range cluster {
			if clsWeight, ok := nodeRecord.Routes[routeName].Cluster[reqClsName]; ok {
				if clsWeight == reqClsWeight {
					slog.Debug("Skip route creation/update", "Route:", routeName)
					return false
				}
			}
			nodeRecord.Routes[routeName].Cluster[reqClsName] = reqClsWeight
		}

		return true
	}
	nodeRecord.Routes[routeName] = &cpRresource.RouteData{RouteName: routeName, Cluster: cluster, Domain: domain}
	return true
}

func addListener(nodeRecord *cpRresource.NodeData, listenerName, routeName string) bool {
	if _, exist := nodeRecord.Listeners[listenerName]; exist {
                slog.Debug("Skip listerner creation, listener exist", "Listener", listenerName, "Route", routeName)
		return false
        }
	nodeRecord.Listeners[listenerName] = &cpRresource.ListenerData{ListenerName: listenerName, RouteName: routeName}
        return true
}

func addNodeId(nodeRecord *cpRresource.NodeData, nodeName string) {
	nodeRecord.NodeId = nodeName
}

func formFederationName(resourceName, authority string) string {
	return federation + authority + "/" + resourceName
}

func isFederated(name string) bool {
	return strings.HasPrefix(name, federation)
}

