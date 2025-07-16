# XDS Control Plane

## Overview 
This project implements a functional xDS Control Plane using [proxyless GRPC](https://grpc.github.io/grpc/core/md_doc_grpc_xds_features.html) instead of sidecars with active weighted endpoint load-balancing.  
The codebase is predominantly sourced from Envoy's [go-control-plane](https://github.com/envoyproxy/go-control-plane/tree/main) library.  
This project was to help me learn about xDS, control plane, service meshes, along with golang.  
Given at this current moment there are not a lot of working examples of a control plane / xDS server with proxyless grpc in the open community, I hope this example can help some of those fortunate folks who stumble across this repo.  
Secured mTLS is also implemented across the data-plane (grpc clients to servers) and the control plane (grpc clients and servers to xDS server)  
Cheers!  

## Learning References
- [Grpc Client setup](/cmd/greet/grpcclient/README.md)  
- [Grpc Server setup](/cmd/greet/grpcserver/README.md)  
- [xDS: Grpc Client LDS](/cmd/greet/xds/LEARNING_README/LDS_README/Client_LDS_README.md)  
- [xDS: Grpc Server LDS](/cmd/greet/xds/LEARNING_README/LDS_README/Server_LDS_README.md)  
- [xDS: Grpc Client RDS](/cmd/greet/xds/LEARNING_README/RDS_README/Client_RDS_README.md)  
- [xDS: Grpc Server RDS](/cmd/greet/xds/LEARNING_README/RDS_README/Server_RDS_README.md)  
- [xDS: Grpc Client CDS](/cmd/greet/xds/LEARNING_README/CDS_README/Client_CDS_README.md)  
- [xDS: Grpc Client EDS](/cmd/greet/xds/LEARNING_README/EDS_README/Client_EDS_README.md)  

## Run
### generate certs if not there
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
> note: Initial weighted LB policy is 90% to server_A and 10% to server_B`
>

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
1 grpc server (s1) in Locality1  
2 grpc servers (s2 & s3) in Locality2  
1 grpc server (s4) in Locality3 "Failover - lower priority"
  
### scenarios
This outlines the traffic split by percentage across all the grpc servers   
  
*** (1) Initial State by percentage Locality1 = 90% & Locality2 = 10% ***  
|  Locality1  |  Locality2  |  Locality2  |  Locality3  |
|  ---------  |  ---------  |  ---------  |  ---------  |
|  90 (s1)    |  5 (s2)     |  5 (s3)     |  0 (s4)     |

*** (2) Remove s1 ***  
|  Locality1  |  Locality2  |  Locality2  |  Locality3  |
|  ---------  |  ---------  |  ---------  |  ---------  |
|  X (s1)     |  50 (s2)    |  50 (s3)    |  0 (s4)     |

*** (3) Remove s3 ***  
|  Locality1  |  Locality2  |  Locality2  |  Locality3  |
|  ---------  |  ---------  |  ---------  |  ---------  |
|  X (s1)     |  100 (s2)   |  X (s3)     |  0 (s4)     |

*** (4) Remove s2 ***  
|  Locality1  |  Locality2  |  Locality2  |  Locality3  |
|  ---------  |  ---------  |  ---------  |  ---------  |
|  X (s1)     |  X (s2)     |  X (s3)     |  100 (s4)   |

*** (5) Update Locality1 weight to 10%, Update Locality2 weight to 90% ***  
|  Locality1  |  Locality2  |  Locality2  |  Locality3  |
|  ---------  |  ---------  |  ---------  |  ---------  |
|  X (s1)     |  X (s2)     |  X (s3)     |  100 (s4)   |

*** (6) Restart s1 ***  
|  Locality1  |  Locality2  |  Locality2  |  Locality3  |
|  ---------  |  ---------  |  ---------  |  ---------  |
|  100 (s1)   |  X (s2)     |  X (s3)     |  0 (s4)     |

*** (7) Restart s2 ***  
|  Locality1  |  Locality2  |  Locality2  |  Locality3  |
|  ---------  |  ---------  |  ---------  |  ---------  |
|  10 (s1)    |  90 (s2)    |  X (s3)     |  0 (s4)     |

*** (8) Restart s3 ***  
|  Locality1  |  Locality2  |  Locality2  |  Locality3  |
|  ---------  |  ---------  |  ---------  |  ---------  |
|  10 (s1)    |  45 (s2)    |  45 (s3)    |  0 (s4)     |

*** (9) Remove XDS ***  
|  Locality1  |  Locality2  |  Locality2  |  Locality3  |
|  ---------  |  ---------  |  ---------  |  ---------  |
|  10 (s1)    |  45 (s2)    |  45 (s3)    |  0 (s4)     |

*** (10) Restart XDS ***  
|  Locality1  |  Locality2  |  Locality2  |  Locality3  |
|  ---------  |  ---------  |  ---------  |  ---------  |
|  90 (s1)    |  5 (s2)     |  5 (s3)     |  0 (s4)     |