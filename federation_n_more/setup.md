
## Example: 
* Clone this code.
* For simplicity, below steps assumes the code is cloned under "/home/user/cms/" 

---
### Combination of Static, Dynamic Service providers and two CMS services to support federation
This example will show static and dynamic registration of service providers with different CMSs. A service consumer getting service provider detail from multiple CMSs and executing RPCs with multiple service providers.


Best way to run this example, will require 2 hosts and steps below will guide through CMS setup on each host. 

#### Assumption, the values (like ip, machine name, etc) will differ. 
> HOST1 & HOST2: Below steps will show ip as "127.0.0.1" i.e. both CMS running on same machine<br>

#### TIP: As you go through below steps, HOST2 setup is very similar to above example setup. It would be easier to setup new machine/host as HOST1 steps.
#### Pre-Work:
1. Clone code to both hosts 
2. Replace host IP with correct values in the client bootstrap files and ResourceFile.yml.
    * Resource file: Under "endpoint", replace ip with static service ip/ports (leave other fields as is)
	  > Replace port/ip under "ClusterLoadAssignment:Endpoint:" with Static service ip and port
    * Client Bootstrap file: Replace ip with CMS service ip. 
	  > Line 4: "server_uri": "127.0.0.1:8001", <br> 
	   &nbsp;&nbsp; Replace with HOST2 CMS IP and port. 
	  
	  > Line 17: "server_uri": "127.0.0.1:8005", <br>
	    &nbsp;&nbsp; Replace with HOST1 CMS IP and port.
		  
	
#### HOST1 SETUP - This host will run a CMS and one instance of service provider. 
  1. Start Configuration Management with static service provider endpoints (gRPC Server).
		* Go to "xds" folder/directory
			> cd /home/user/cms/xds
		* Export configuration file 
			> export XDS_RESOURCE_FILE=/home/user/cms/xds/ResourceFile.yml
		* Start CMS
			> go run main.go --port 8005
	
  2. Start service for dynamic registration with CMS 
		* Go to "app" folder/directory
			> cd /home/user/cms/app
		* Export bootstrap file 
			> export GRPC_XDS_BOOTSTRAP='../xds/bootstrap/server_register_federation_bootstrap.json'
		* Start service 
			> go run src/xds_enabled_server.go --grpcport 50057 --servername Dynamic_FEDERATION_Server1

#### HOST2 SETUP - This host will run a CMS, static and dynamic service providers and a client/consumer service. The setup is very similar to Above example "Combination of Static & Dynamic Service providers ".
  1. Start services, that will be part of static registry in CMS.
        * Go to "app" folder/directory
		    > cd /home/user/cms/app
        * Start 3 services (these will be registered in CMS via file loading)
        	> go run src/server.go  --grpcport 50051 --servername Static_Server_1 <br>
		> go run src/server.go  --grpcport 50053 --servername Static_Server_2 <br>
		> go run src/server.go  --grpcport 50055 --servername Static_Server_3

  2. Start Configuration Management with static service provider endpoints (gRPC Server).
		* Go to "xds" folder/directory
			> cd /home/user/cms/xds
		* Export configuration file 
			> export XDS_RESOURCE_FILE=/home/user/cms/xds/ResourceFile.yml
		* Start CMS
			> go run main.go --port 8001
	
  3. Start service for dynamic registration with CMS 
		* Go to "app" folder/directory
			> cd /home/user/cms/app
		* Export bootstrap file 
			> export GRPC_XDS_BOOTSTRAP='../xds/bootstrap/server_add_to_existing_cluster_xdstp_bootstrap.json'
		* Start service 
			> go run src/xds_enabled_server.go --grpcport 50057 --servername Dynamic_Server_4
	
  4. Start Consumer service (gRPC Client), that will connect to CMS and connect to 2 of static services and a dynamic registered service provider. 
		* Go to "app" folder/directory
			> cd /home/user/cms/app
		* Export bootstrap file 
			> export GRPC_XDS_BOOTSTRAP='../xds/bootstrap/client_federation_support_bootstrap.json'
		* Start client service 
			> go run src/client.go --host xds:///xdstp.upstream.xdspoc.com 

#### Expected Behavior: Observe Client RPC call going to service providers registered with different CMS (running on HOST1 & HOST2). Service provider can be identified based on client stdout, printed response will contain service provider name. 

