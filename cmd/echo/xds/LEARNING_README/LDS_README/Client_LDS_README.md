[LDS RFC](https://github.com/grpc/proposal/blob/master/A27-xds-global-load-balancing.md#lds)

## Overview
LDS or listeners are the initial resources typically used as an entry point into xDS configurations.  
Grpc uses the target URI with the prefixed form `xds:///` to retrieve the LDS when creating a channel.  
```go
grpc.NewClient("xds:///this.is.the.lds", ...)
```
LDS has a relation ship with [RDS](/cmd/echo/xds/LEARNING_README/RDS_README/RDS_README.md)

### Grpc Client specific LDS
```go
func makeClientListener() *listener.Listener {
	routerConfig, _ := anypb.New(&router.Router{})

	httpConnectionManager := &hcm.HttpConnectionManager{
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource: &core.ConfigSource{
                    // tells control plane to use ADS for Route Configs
					ConfigSourceSpecifier: &core.ConfigSource_Ads{ 
						Ads: &core.AggregatedConfigSource{},
					},
				},
				RouteConfigName: "local_route",
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name:       "http-router",
			ConfigType: &hcm.HttpFilter_TypedConfig{TypedConfig: routerConfig},
		}},
	}

	httpConnectionManagerAsAny, err := anypb.New(httpConnectionManager)
	if err != nil {
		panic(err)
	}

	return &listener.Listener{
		Name: "this.is.the.lds",
        // grpc only support APIListener
		ApiListener: &listener.ApiListener{
			ApiListener: httpConnectionManagerAsAny, 
		},
		FilterChains: []*listener.FilterChain{{
			Name: "filter-chain",
			Filters: []*listener.Filter{{
				Name:       wellknown.HTTPConnectionManager,
				ConfigType: &listener.Filter_TypedConfig{TypedConfig: httpConnectionManagerAsAny},
			}},
		}},
	}
}
```

