This package contains an xDS server that has configurations for `grpc clients` and `grpc servers`.  
Current configuration is a *single* cluster route with TLS contexts & a weighted balance of 90% to a grpc server and 10% to another, and one more configuration for priority / failover grpc server.  


# Learning xDS

### REFS
- https://github.com/grpc/proposal/blob/master/A27-xds-global-load-balancing.md
- https://grpc.github.io/grpc/core/md_doc_grpc_xds_features.html
- https://www.youtube.com/watch?v=Z3X6kD_1SFo


# mTLS between grpc client and servers and xDS
### bootstrap.json
```json
{
  "xds_servers": [
    {
      "server_uri": "127.0.0.1:18000",
      "channel_creds": [
        {
          "type": "tls",
          "config": {
            "ca_certificate_file": "../../../scripts/certs/imposter/imposterCA.pem",
            "certificate_file": "../../../scripts/certs/imposter/imposter_grpcclient.pem",
            "private_key_file": "../../../scripts/certs/imposter/imposter_grpcclient_key.pem"
          }
        }
      ],
      "server_features": [
        "xds_v3"
      ]
    }
  ],
  "node": {
    "id": "server123",
    "cluster": "backend_cluster"
  },
  "server_listener_resource_name_template": "example/resource/%s",
  "certificate_providers": {
    "my_custom_cert_provider": {
      "plugin_name": "file_watcher",
      "config": {
        "certificate_file": "../../../scripts/certs/normal/grpcserver.pem",
        "private_key_file": "../../../scripts/certs/normal/grpcserver_key.pem",
        "ca_certificate_file": "../../../scripts/certs/normal/ca.pem",
        "refresh_internal": "600s"
      }
    }
  }
}
```

client code in Golang
```go
import xdscreds "google.golang.org/grpc/credentials/xds"

creds, err := xdscreds.NewClientCredentials(xdscreds.ClientOptions{FallbackCreds: insecure.NewCredentials()})
...
conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(creds))
```

server code in Golang
```go
import xdscreds "google.golang.org/grpc/credentials/xds"

creds, err := xdscreds.NewServerCredentials(xdscreds.ServerOptions{FallbackCreds: insecure.NewCredentials()})
...
srv, err := xds.NewGRPCServer(grpc.Creds(creds))
```

xDS code in Golang
```go
func loadTLSCredentials() (credentials.TransportCredentials, error) {
    // Load server's certificate and private key
    serverCert, err := tls.LoadX509KeyPair("../../../scripts/certs/normal/xdsserver.pem", "../../../scripts/certs/normal/xdsserver_key.pem")
    if err != nil {
        return nil, err
    }

    // Create the credentials and return it
    config := &tls.Config{
        Certificates: []tls.Certificate{serverCert},
        ClientAuth:   tls.NoClientCert,
    }

    return credentials.NewTLS(config), nil
}

func main() {
    ...
    creds, err := loadTLSCredentials()
    xDS, err := grpc.NewServer(grpc.Creds(creds))
    ...
}
```

### Insecure
```json
{
    "type": "insecure"
}
```

### mTLS
```json
{
    "type": "tls",
    "config": {
        "ca_certificate_file": "../../../scripts/certs/normal/ca.pem",
        "certificate_file": "../../../scripts/certs/normal/grpcserver.pem",
        "private_key_file": "../../../scripts/certs/normal/grpcserver_key.pem"
    }
}
```

### Imposter
```json
{
    "type": "tls",
    "config": {
        "ca_certificate_file": "../../../scripts/certs/imposter/imposterCA.pem",
        "certificate_file": "../../../scripts/certs/imposter/imposter_grpcserver.pem",
        "private_key_file": "../../../scripts/certs/imposter/imposter_grpcserver_key.pem"
    }
}
```