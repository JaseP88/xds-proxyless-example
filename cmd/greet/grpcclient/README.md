Any client connecting to xDS requires a bootstrapping json file.  
Furthermore, you will have to set an env var `GRPC_XDS_BOOTSTRAP=path/to/bootstrap.json` before starting the client.

### Ref
- [bootstrap doc](https://grpc.github.io/grpc/cpp/md_doc_grpc_xds_bootstrap_format.html)
### Sample
```json
{
    "xds_servers": [
        {
            "server_uri": "127.0.0.1:18000", // location of your xDS server
            "channel_creds": [
                {
                    "type": "insecure"
                }
            ],
            "server_features": [
                "xds_v3"
            ]
        }
    ],
    "node": {
        "id": "client123" // unique identifier for the client, this will be used in the internal xDS cache to maintain configuration for this client
    }
}
```
### gRPC client code setup
For this to work you will need to import `xds`
  
```go
import (
    _ "google.golang.org/grpc/xds" // To install the xds resolvers and balancers.
)
```
> info: you do not have to use it just import it
>
  
To create the connection to gRPC server target now will reference an LDS (Listener) name.  
```go
conn, err := grpc.NewClient("xds:///connect.me.to.grpcserver", grpc.WithTransportCredentials(creds))
```

The `xds:///` prefix here will denote gRPC resolvers to connect to xDS control plane first to get the endpoint information of your gRPC server.  

> info: there are other semantics gRPC supports besides xds:///
>

The string `connect.me.to.grpcserver` is the listener name that will start a journey from LDS->RDS->CDS->EDS (endpoint).  

```yaml
static_resources:
  listeners:
  - name: listener_0 ## The listener (LDS)
    address:
      socket_address: { address: 127.0.0.1, port_value: 10000 }
    filter_chains:
    - filters:
      - name: envoy.filters.network.http_connection_manager
        typed_config:
          "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
          stat_prefix: ingress_http
          codec_type: AUTO
          route_config:
            name: local_route ## The route (RDS)
            virtual_hosts:
            - name: local_service
              domains: ["*"]
              routes:
              - match: { prefix: "/" }
                route: { cluster: some_service } ## route reference cluster "some_service"
          http_filters:
          - name: envoy.filters.http.router
            typed_config:
              "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router
  clusters:
  - name: some_service ## The cluster (CDS)
    connect_timeout: 0.25s
    type: STATIC
    lb_policy: ROUND_ROBIN
    load_assignment:
      cluster_name: some_service  ## endpoint reference cluster "some_service"
      endpoints:  ## The endpoints (EDS) 
      - lb_endpoints:
        - endpoint:
            address:
              socket_address:
                address: 127.0.0.1
                port_value: 1234
```