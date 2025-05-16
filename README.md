# XDS Control Plane

### insecugit
This branch is with tls off


## Overview 
This project implements a functional xDS Control Plane using [proxyless GRPC](https://grpc.github.io/grpc/core/md_doc_grpc_xds_features.html) instead of sidecars with active weighted endpoint load-balancing.  
The codebase is predominantly sourced from Envoy's [go-control-plane](https://github.com/envoyproxy/go-control-plane/tree/main) library.  
This project was to help me learn about xDS, control plane, service meshes, along with golang.  
Given at this current moment there are not a lot of working examples of a control plane / xDS server with proxyless grpc in the open community, I hope this example can help some of those fortunate folks who stumble across this repo.  
Secured mTLS is also implemented across the data-plane (grpc clients to servers) and the control plane (grpc clients and servers to xDS server)  
Cheers

## Run
### generate certs
```sh
cd pathto/scripts/certs/normal
sh generatecert.sh
```

### xDS Server
To run the xDS server.  
`note: xDS server runs on localhost:18000`
```sh
cd pathto/cmd/auth/xds
go run .
```

### (2) grpc servers 
To run grpc servers.  
`note: Servers run locally with separate ports *50051* & *50053*. Also there is a health port with port value port+1`
```sh
cd pathto/cmd/auth/grpcserver
export GRPC_XDS_BOOTSTRAP=pathto/cmd/auth/grpcserver/server1_bootstrap.json
go run . -p=50051 -n="server_A"
# in another terminal 
export GRPC_XDS_BOOTSTRAP=pathto/cmd/auth/grpcserver/server2_bootstrap.json
go run . -p=50053 -n="server_B"
```
`note: Initial weighted LB policy is 90% to server_A and 10% to server_B`

### grpc client
To run a client.
```sh
cd pathto/cmd/auth/grpcclient
export GRPC_XDS_BOOTSTRAP=pathto/cmd/auth/grpcclient/bootstrap.json
go run .
```
`note: Initial transactions to send is 10.  You can modify this when running your client with -tc flag`

### What to look for
- Grpc client and servers upon start up will receive resources from the xDS server.
- Look at command prompts from xDS server. Follow the direction and input in different weights where 
  the sum of the 2 weights equals 100. (ie: A=50, B=50)
  `note: increase initial transaction count via flag -tc to have time for this activity`
- Observe server traffic

### Extended  
### setup  
1 xDS  
1 grpc client  
1 grpc server in clusterA  
2 grpc server in clusterB  
1 grpc server in Failover cluster  
  
### scenarios
This outlines the traffic split by percentage across all the grpc servers   
  
***(1) initial state***  
|  ClusterA   |  ClusterB_1 |  ClusterB_2 |  Cluster Failover  |
|  ---------  |  ---------  |  ---------  |  ----------------  |
|  90         |  5          |  5          |  0                 |

***(2) change ClusterA weight to 0, change ClusterB weight to 100***  
|  ClusterA   |  ClusterB_1 |  ClusterB_2 |  Cluster Failover  |
|  ---------  |  ---------  |  ---------  |  ----------------  |
|  0          |  50         |  50         |  0                 |

***(3) remove ClusterB_2***  
|  ClusterA   |  ClusterB_1 |  X          |  Cluster Failover  |
|  ---------  |  ---------  |  ---------  |  ----------------  |
|  0          |  100        |  0          |  0                 |

***(4) remove ClusterB_1***  
|  ClusterA   |  X          |  X          |  Cluster Failover  |
|  ---------  |  ---------  |  ---------  |  ----------------  |
|  100        |  0          |  0          |  0                 |

***(5) remove ClusterA***  
|  X          |  X          |  X          |  Cluster Failover  |
|  ---------  |  ---------  |  ---------  |  ----------------  |
|  0          |  0          |  0          |  100               |

***(6) remove XDS***  
|  X          |  X          |  X          |  Cluster Failover  |
|  ---------  |  ---------  |  ---------  |  ----------------  |
|  0          |  0          |  0          |  100               |

***(7) restart ClusterA***  
|  ClusterA   |  X          |  X          |  Cluster Failover  |
|  ---------  |  ---------  |  ---------  |  ----------------  |
|  100        |  0          |  0          |  0               |

***(8) restart ClusterB_2***  
|  ClusterA   |  X          |  ClusterB_2 |  Cluster Failover  |
|  ---------  |  ---------  |  ---------  |  ----------------  |
|  0          |  0         |  100         |  0                 |

***(9) restart ClusterB_1***  
|  ClusterA   |  ClusterB_1 |  ClusterB_2 |  Cluster Failover  |
|  ---------  |  ---------  |  ---------  |  ----------------  |
|  0          |  50         |  50         |  0                |

***(10) restart XDS***  
|  ClusterA   |  ClusterB_1 |  ClusterB_2 |  Cluster Failover  |
|  ---------  |  ---------  |  ---------  |  ----------------  |
|  90         |  5          |  5          |  0                 |