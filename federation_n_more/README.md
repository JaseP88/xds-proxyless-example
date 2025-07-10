# XDS Control Plane


### Dynamic discovery (works anywhere)
#### - Use upstream LDS request to register. 
```sh
URL type: type.googleapis.com/envoy.config.listener.v3.Listener
Resource name: grpc/server?xds.resource.listening_address=11.222.333.444:5555"
```
![alt text](Dynamic_Discovery.png)


### xDS Fallback
#### Configure multiple xDS services.

### xDS Federation
#### Slice & dice, load distribution and isolation. 
![alt text](grouping_multi_cluster.png)


## Lesson
### xDS Fallback
#### 1. Can configure multiple xDS services.
#### 2. (xDS server goes down, after Watcher is established at client side - You're on Your Own, Kid.) 
![alt text](multiple_xDS.png)

### Support for Multiple cluster
#### 1. Isolation (Federation)
#### 2. Distribution of load
#### 3. Upstream/End-point Failure recognition 
#### 4. (Avoidance of bad clusters)
![alt text](grouping_bad_cluster.png)
