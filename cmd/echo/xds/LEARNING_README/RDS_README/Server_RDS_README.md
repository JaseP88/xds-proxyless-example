[Server xDS Configurations](https://github.com/grpc/proposal/blob/master/A36-xds-for-servers.md). 

### Grpc Server specific RDS
```go
func makeServerRoute() *route.RouteConfiguration {
	return &route.RouteConfiguration{
		Name: RouteName,
		VirtualHosts: []*route.VirtualHost{{
			Name:    "VH",
			Domains: []string{"*"},
			Routes: []*route.Route{{
				Name: "http-router",
				Match: &route.RouteMatch{
					PathSpecifier: &route.RouteMatch_Prefix{
						Prefix: "/",
					},
				},
				Action: &route.Route_NonForwardingAction{
					NonForwardingAction: &route.NonForwardingAction{
						// A36: Route.non_forwarding_action is expected for all Routes used on server-side and Route.route continues to be expected for all Routes used on client-side
					},
				},
			}},
		}},
	}
}
```