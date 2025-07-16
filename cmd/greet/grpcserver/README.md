gRPC servers can be clients of xDS control plane.  Thus the setup to interact with xDS will be similar.

### Ref
- [gRPC client bootstrap setup](/cmd/greet/grpcclient/README.md)  
- [bootstrap doc](https://grpc.github.io/grpc/cpp/md_doc_grpc_xds_bootstrap_format.html)

### gRPC server code setup
Import `xds`
  
```go
import (
    _ "google.golang.org/grpc/xds"
)
```
  
Then use xds to create the xDS server.  
```go
greeterService, err := xds.NewGRPCServer(grpc.Creds(creds))
```   

> info: To simulate multiple instances of your backend services for the demo, multiple bootstrap jsons are provided.  Ensure each gRPC servers have the specific paths to these bootstrap files in the env var GRPC_XDS_BOOTSTRAP when starting.
>
