package xdscp

import (
        "context"
        "fmt"
        "sync/atomic"

        cpRresource "xdsmod/internal/resources"
	cpLoader    "xdsmod/internal/loader"

        "log/slog"
        "github.com/envoyproxy/go-control-plane/pkg/cache/types"
        cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
        resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

type NodeStore struct {
	nodeData		*cpRresource.NodeData
	version			int32
}

type XdsStore struct {
        ctx                 context.Context                
        Cache               cachev3.SnapshotCache          // Resource snapshot.
	nodeStore           map[string]*NodeStore	
        
}

func (xdsCache *XdsStore) createUpstreamResources(request cpLoader.RequestData) bool {
        if _, found := xdsCache.nodeStore[request.UpstreamNodeId]; found {
                slog.Info("Skip Upstream registration, nodeID exist", " nodeID", request.UpstreamNodeId)
                return false
        }
	upstreamresource := cpLoader.CreateUpstreamData(request)
	xdsCache.nodeStore[request.UpstreamNodeId] = &NodeStore{nodeData: upstreamresource, version: 0}
        updatedResources := cpRresource.DefaultUpstreamResources(cpLoader.GetUpstreamId(request), request.Host, request.Port) //Restrict to one upstream entry
	xdsCache.UpdateCache(updatedResources, request.UpstreamNodeId)
	return true
}

func (xdsCache *XdsStore) updateServiceResources(request cpLoader.RequestData)  {
	if _, found := xdsCache.nodeStore[cpLoader.GetServiceId(request)]; found {
                if cpLoader.UpdateServiceData(request, xdsCache.nodeStore[cpLoader.GetServiceId(request)].nodeData) {
			atomic.AddInt32(&xdsCache.nodeStore[cpLoader.GetServiceId(request)].version, 1)
		} else{
			slog.Info("Skip Service update, no resource change.", "ServiceId", cpLoader.GetServiceId(request))
			return 
		}
	} else {
		slog.Info("New Service, add resources", "ServiceId", cpLoader.GetServiceId(request))
		serviceNodeData := cpLoader.CreateServiceData(request)
		xdsCache.nodeStore[cpLoader.GetServiceId(request)] = &NodeStore{nodeData: serviceNodeData, version: 0}
	}

	updatedResources := cpRresource.CreateResources(xdsCache.nodeStore[cpLoader.GetServiceId(request)].nodeData)
	xdsCache.UpdateCache(updatedResources, cpLoader.GetServiceId(request))       
        return 
}

func (xdsCache *XdsStore) LoadFileResourceData(confData *cpLoader.ConfigData) {
        for serviceName, _ := range confData.Service {
                if _, found := xdsCache.nodeStore[serviceName]; found {
			slog.Info("FileLoad - Found Service entry, check update required", "ServiceId", serviceName)
			if cpLoader.MapFileDataToNodeConfig(xdsCache.nodeStore[serviceName].nodeData, confData, serviceName) {
				atomic.AddInt32(&xdsCache.nodeStore[serviceName].version, 1)
			} else {
				slog.Info("FileLoad - Skip Service update, no resource change.", "ServiceId", serviceName)
				continue
			}
		} else {
			nodeRecord := &cpRresource.NodeData{NodeId: serviceName, Listeners: map[string]*cpRresource.ListenerData{}, Routes: map[string]*cpRresource.RouteData{}, Clusters: map[string]*cpRresource.ClusterData{}, Endpoints: map[string]*cpRresource.EndpointData{}}
			slog.Info("FileLoad - New Service, add resources", "ServiceId", serviceName)
			cpLoader.MapFileDataToNodeConfig(nodeRecord, confData, serviceName)
			xdsCache.nodeStore[serviceName] = &NodeStore{nodeData: nodeRecord, version: 0}
		}

		updatedResources := cpRresource.CreateResources(xdsCache.nodeStore[serviceName].nodeData)
		xdsCache.UpdateCache(updatedResources, serviceName)
	}
}

func (xdsCache *XdsStore) UpdateCache(resources map[resource.Type] []types.Resource, nodeId string) error {
        snapshot, err := cachev3.NewSnapshot(fmt.Sprint(xdsCache.nodeStore[nodeId].version), resources)
        if err != nil {
                return fmt.Errorf("Failed to create new snapshot cache: %v", err)

        }

        if err := xdsCache.Cache.SetSnapshot(xdsCache.ctx, nodeId, snapshot); err != nil {
                return fmt.Errorf("Failed to update resource snapshot in management server: %v", err)
        }

        slog.Info("Updated snapshot cache with resource snapshot.", "NodeId", nodeId)
	slog.Info("==", "Listener:",  snapshot.GetResources(resource.ListenerType))
	slog.Info("==", "Route   :", snapshot.GetResources(resource.RouteType))
	slog.Info("==", "Cluster :", snapshot.GetResources(resource.ClusterType))
	slog.Info("==", "Endpoint:", snapshot.GetResources(resource.EndpointType))
        return nil
}

func (xdsCache *XdsStore) Update(requestData cpLoader.RequestData) {
	xdsCache.createUpstreamResources(requestData)
        xdsCache.updateServiceResources(requestData)
}

func GetXdsStore(ctx context.Context) *XdsStore{
        return &XdsStore{
                ctx:			ctx,
                Cache:			cachev3.NewSnapshotCache(false, cachev3.IDHash{}, nil),
		nodeStore:		make(map[string] *NodeStore),
        }
}

func contains(arr []uint32, val []uint32) bool {
        for _, p := range val{
            ret := false
            for _, v := range arr {
                if v == p {
                    ret = true
                    break
                }
            }
            if !ret {
               return false
            }
        }
        return true
}
