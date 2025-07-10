package xdscp

import (
        "context"
        "fmt"
        "net"
        "os"
        "strings"
        "strconv"
	"log/slog"

	cpxdsStore  "xdsmod/internal/xdsstore"
	cpLoader    "xdsmod/internal/loader"

	discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
)

const (
        ServerListenerResourceNameTemplate = "grpc/server?xds.resource.listening_address="
)

type Processor struct {
	XDSstore	*cpxdsStore.XdsStore
}

func InitProcessor(ctx context.Context) *Processor {
	return &Processor{XDSstore: cpxdsStore.GetXdsStore(ctx)}
}

func (processor *Processor) LoadResourceFile(resourceFile string) {
	if len(resourceFile) != 0 {
                if _, err := os.Stat(resourceFile); err != nil {
                        slog.Error("Resource file loading Failed", "Error", err)
                } else {
                        fileData, err := cpLoader.ParseYaml(resourceFile)
                        if err != nil {
                                slog.Error("Failed load Resource file", "err", err)
                        } else {
                                processor.XDSstore.LoadFileResourceData(fileData)
                        }
                }
        } else {
                slog.Error("Missing Resource File to load")
        }
}

func (processor *Processor) ProcessDiscoveryRequest(req *discoveryservice.DiscoveryRequest) {

	if len(req.GetResourceNames()) != 1 {
		slog.Error("Invalid discovery request", "resource name", req.GetResourceNames())
		return
	}

	if processor.validateRequestData(req) {
		slog.Error("Discovery request validation failed")
		return 
	}

	reqData, err := processor.setRequestData(req) 
	if err != nil {
		slog.Error("Failed to parse discovery request")
		return 
	}

	processor.XDSstore.Update(reqData)
}

func (processor *Processor) setRequestData(req *discoveryservice.DiscoveryRequest) (cpLoader.RequestData, error) {

	host, port, err := processor.getAddress(req.GetResourceNames()[0])
        if err != nil {
		return cpLoader.RequestData{}, fmt.Errorf("Skip Registration - Failed to get resource name %s, error:%v", req.GetResourceNames()[0], err)
        }

	return cpLoader.RequestData{
		UpstreamNodeId: req.GetNode().GetId(),
		ResourceName: req.GetResourceNames()[0],
		Host:   host,
		Port:   port,
		BackupFlag: req.GetNode().GetMetadata().GetFields()["isbackup"].GetBoolValue(),
		ServiceId: req.GetNode().GetMetadata().GetFields()["service_id"].GetStringValue(),
		UpstreamWeight: uint32(req.GetNode().GetMetadata().GetFields()["upstream_weight"].GetNumberValue()),
		Cluster: req.GetNode().GetCluster(),
		ClsWeight: uint32(req.GetNode().GetMetadata().GetFields()["cluser_weight"].GetNumberValue()),
		LbWeight: uint32(req.GetNode().GetMetadata().GetFields()["locality_weight"].GetNumberValue()),
		Region: req.GetNode().GetLocality().GetRegion(),
		Zone: req.GetNode().GetLocality().GetZone(),
		SubZone: req.GetNode().GetLocality().GetSubZone(),
		Priority: uint32(req.GetNode().GetMetadata().GetFields()["priority"].GetNumberValue()),
		EndpointName:req.GetNode().GetMetadata().GetFields()["endpoint_name"].GetStringValue(),
		ListenerName: req.GetNode().GetMetadata().GetFields()["listener_name"].GetStringValue(),

	}, nil
}

func (processor *Processor) validateRequestData(req *discoveryservice.DiscoveryRequest) bool {

	var missingMandatoryField bool

	if isEmpty(req.GetNode().GetId()) {
		missingMandatoryField = true
		slog.Error("Discovery request missing node id")
	}

	if isEmpty(req.GetResourceNames()[0]) {
		missingMandatoryField = true
                slog.Error("Discovery request missing node resource name")
        }

	if isEmpty(req.GetNode().GetMetadata().GetFields()["service_id"].GetStringValue()) {
		missingMandatoryField = true
                slog.Error("Discovery request missing node service id")
        }

	return missingMandatoryField
}

func isEmpty(data string) bool{
	if len(strings.TrimSpace(data)) == 0 {
		return true
	}
	return false
}

func (processor *Processor) getAddress(uri string) (string, uint32, error) {
        address, found := strings.CutPrefix(uri, ServerListenerResourceNameTemplate)
        if !found {
                return "", 0, fmt.Errorf("Invalid upstream address: %s:%s", uri, ServerListenerResourceNameTemplate)
        }

        svrHost, svrPort, err := net.SplitHostPort(address)
        if err != nil {
                return "", 0, fmt.Errorf("failed to get address from upstream resource name: %v", err)
        }

        svPort, err := strconv.Atoi(svrPort)
        if err != nil {
                return "", 0, fmt.Errorf("Failed to convert port to int: %v", err)
        }

	return svrHost, uint32(svPort), err

}
