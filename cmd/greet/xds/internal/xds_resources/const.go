package resources

import (
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"google.golang.org/protobuf/encoding/prototext"
	"strings"
)

const (
	ClusterA            = "ClusterA"
	ClusterB            = "ClusterB"
	RouteName           = "local_route"
	GrpcClientListener  = "connect.me.to.grpcserver"
	GrpcServer1Listener = "example/resource/127.0.0.1:50051"
	GrpcServer2Listener = "example/resource/127.0.0.1:50053"
	GrpcServer3Listener = "example/resource/127.0.0.1:50055"
	GrpcServer4Listener = "example/resource/127.0.0.1:50057"
	UpstreamHost        = "127.0.0.1"
	UpstreamPort_50051       = 50051
	UpstreamPort_50053       = 50053
	UpstreamPort_50055       = 50055
	UpstreamPort_50057       = 50057
)

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
