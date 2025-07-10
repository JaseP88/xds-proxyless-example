# XDS Control Plane


### Dynamic discovery (works anywhere)
    - Use requests from gRPC server to identity (new) and register. 
```sh
URL type: type.googleapis.com/envoy.config.listener.v3.Listener
Resource name: grpc/server?xds.resource.listening_address=11.222.333.444:5555"
```
![alt text](Dynamic_Discovery.png)


### xDS Fallback
    - Can configure multiple xDS services.

### xDS Federation
    - slice & dice, load distribution and isolation. 
![alt text](grouping_multi_cluster.png)


### Lesson
    - Single point of failure: Watcher established at client side. 
    ![alt text](multiple_xDS.png)
    - Multi-cluster, absence of failure recognition.
    ![alt text](grouping_bad_cluster.png)
