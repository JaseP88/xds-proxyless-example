package xdscp

import (
        //"context"
        "fmt"
        "os"
        "gopkg.in/yaml.v3"
        "log/slog"
        //discoveryservice "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
        //resource "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
)

type ConfigData struct {
    Service     map[string]Listener     `yaml:"service"`
    ClusterLoadAssignment                map[string][]Locality          `yaml:"clusterloadassignment"`
}

type Listener map[string]Route
type Route struct {
        Name    string  `yaml:"route"`
        Clusters []Cluster   `yaml:"cluster"`
}

type Cluster struct {
        Name    string  `yaml:"clusterName"`
        Weight  uint32  `yaml:"weight"`
        FederatedEdsName string  `yaml:"federatedEdsName"`
}

type Locality   struct {
        Region  string          `yaml:"region"`
        Zone    string          `yaml:"zone"`
        SubZone string          `yaml:"subzone"`
        Endpoints      []Endpoint  `yaml:"endpoint"`
        Weight          uint32          `yaml:"weight"`
        Priority    uint32          `yaml:"priority"`
}

type Endpoint struct {
        Host    string          `yaml:"host"`
        Port     []uint32                `yaml:"port"`
        Weight      uint32          `yaml:"weight"`
}


/*
type ConfigData struct {
    Service     []Service       `yaml:"service"`
    Listener    []Listener      `yaml:"upstream"`
    Cluster     []Cluster      `yaml:"cluster"`
    Endpoint    []Endpoint      `yaml:"endpoint"`
    Location    []Location      `yaml:"location"`
    Locality    []Locality      `yaml:"locality"`
    Federation  []Federation    `yaml:"federation"`
}

type ConfigMap struct {
    Service     map[string]Service
    Listener    map[string]Listener
    Cluster     map[string]Cluster
    Endpoint    map[string]Endpoint
    Location    map[string]Location
    Locality    map[string]Locality
    Federation  map[string]Federation
}

type Service struct {
        Name    string          `yaml:"name"`
        Upstream        []string  `yaml:"upstream"`
}

type Listener struct {
        Name    string          `yaml:"name"`
        Cluster []string        `yaml:"cluster"`
        Route   string          `yaml:"route"`
}

type Cluster struct {
        Name            string          `yaml:"name"`
        Weight          uint32          `yaml:"weight"`
}

type Endpoint struct {
        Name            string          `yaml:"name"`
        Host            string          `yaml:"host"`
        Ports           []uint32        `yaml:"ports"`
        Weight          uint32          `yaml:"weight"`
}

type Location struct {
        Name            string          `yaml:"name"`
        Region          string          `yaml:"region"`
        Zone            string          `yaml:"zone"`
        SubZone         string          `yaml:"subzone"`
}

type Locality struct {
        Name            string          `yaml:"name"`
        Weight          uint32          `yaml:"weight"`
        Priority        uint32          `yaml:"priority"`
        Backend        []string       `yaml:"endpoint"`
        Location        string          `yaml:"location"`
}

type Federation struct {
        Authority       string          `yaml:"authority"`
        Route           []string        `yaml:"route"`
        Cluster         []string        `yaml:"cluster"`
        Backend         []string        `yaml:"endpoint"`
}



type ConfigData struct {
        Service		[]Service	`yaml:"node"`
        //Listener        []Listener      `yaml:"listener"`
	Federation	[]Federation	`yaml:"federation"`
        Route           []Route         `yaml:"route"`
        Cluster         []Cluster       `yaml:"cluster"`
        Endpoint        []Endpoint      `yaml:"endpoint"`
        Locality        []Locality      `yaml:"locality"`
        //Backend         []Backend       `yaml:"backend"`
        Location        []Location      `yaml:"location"`
}

type Service struct {
        Name            string          `yaml:"name"`
        UpstreamName	[]string        `yaml:"upstream"`
	EnableFederation	bool	`yaml:"enableFederation"`
	FederationName	string     	`yaml:"federationName"`
}

type Federation struct {
	Name            string          `yaml:"name"`
	Authority	string		`yaml:"authority"`
	//Resource	string		`yaml:"resource"`
	ResourceName	string          `yaml:"resourcename"`
	//UpstreamName	string		`yaml:"upstreamName"`
}
*/
/*
type Upstream struct {
	Name            string          `yaml:"name"`
	//cluster can be empty, if federated
	Cluster		[]string        `yaml:"upstreamName"` //merge with weight
}


type Listener struct {
        Name            string          `yaml:"name"`
        Route           string          `yaml:"route"`
	Xdstp           bool            `yaml:"xdstp"`
	XdstpName     string          `yaml:"xdstpName"`
}
*/
/*
type Route struct {
        Name            string          `yaml:"name"`
	UpstreamName    string        `yaml:"upstreamName"`
        //Multiple clusters - Enabler
        //Cluster with weight. Default weight = 1
        ClusterName     []string]`yaml:"clusterName"`
        //Domain based routing, accepts special charater. Example: *.nextphase.com
        //Should match(regex) listener name. Example: tsi.nextphase.com
        Domain          string          `yaml:"domain"`
}

type Cluster struct {
	Name            string          `yaml:"name"`
	//Optional - either Route or Upstream
	UpstreamName    string        `yaml:"upstreamName"`
	//OR - Optional
	//RouteName	string          `yaml:"name"`
	//Optional
	Weight            string          `yaml:"weight"`
        //Optional - required for federation
        //Endpoint     string	`yaml:"endpointServiceName"`
	//Cluster  ---> an EDS ---> many locality --> a location and many endpoints
        LocalityName	[]string	`yaml:"locality"`
}

type Locality struct {
        Name            string          `yaml:"name"`
        //Default = 1
        Weight          uint32          `yaml:"weight"`
        Priority        uint32           `yaml:"priority"`
        Endpoint       []string       `yaml:"endpoint"`
        Location       string           `yaml:"location"`
}
//reusable
type Location struct {
        Name            string          `yaml:"name"`
        Region          string          `yaml:"region"`
        Zone            string          `yaml:"zone"`
        SubZone         string          `yaml:"subZone"`
}

//resuable
type Endpoint  struct {
        Name            string          `yaml:"name"`
	Host            string          `yaml:"host"`
        Port           []uint32        `yaml:"port"`
        Weight          uint32          `yaml:"weight"`
}
*/
/*
type Endpoint struct {
        Name            string          `yaml:"name"`
        Host            string          `yaml:"host"`
        LocalityName    string          `yaml:"localityName"`
}

type Locality struct {
        Name            string          `yaml:"name"`
        //Default = 1
        Weight          uint32          `yaml:"weight"`
        Priority        uint32           `yaml:"priority"`
        Endpoints	[]string       `yaml:"backend"`
        Location       string		`yaml:"location"`
}

type Endpoints  struct {
        Name            string          `yaml:"name"`
        Ports           []uint32        `yaml:"ports"`
        //Defualt = UNKNOWN
        HealthStatus    int32          `yaml:"healthStatus"`
        //Default = 1
        Weight          uint32          `yaml:"weight"`
        LocationName    string          `yaml:"locationName"`
}

type Location struct {
        Name            string          `yaml:"name"`
        Region          string          `yaml:"region"`
        Zone            string          `yaml:"zone"`
        SubZone         string          `yaml:"subZone"`
}
*/

func ParseYaml(file string) (*ConfigData, error) {
        yamlData, err := os.ReadFile(file)
        if err != nil {
                slog.Error("Failed to load manifest yaml file", file)
        }

        var configData ConfigData
        err = yaml.Unmarshal(yamlData, &configData)
	if err != nil {
		slog.Error("Failed to load manifest file", file)
	}

        fmt.Println(configData)

	return &configData, err

}

/*
func structToMap(configData ConfigData) ConfigMap {
        //Run through CofigData structure and turn data into resource map
	var configMap ConfigMap

        for _, svcitem := range configData.Service {
                configMap.Service[svcitem.Name] = svcitem
        }

        for _, upsitem := range configData.Listener {
                configMap.Listener[upsitem.Name] = upsitem
        }

        for _, clsitem := range configData.Cluster {
                configMap.Cluster[clsitem.name] = clsitem
        }

        for _, edsitem := range configData.Endpoint {
                configMap.Endpoint[edsmap.Name] = edsitem
        }

        for _, lctitem := range configData.Locality {
                configMap.Locality[lctitem.Name] = lctitem
        }

        for _, locitem := range configData.Location {
                configMap.Location[locitem.Name] = locitem
        }

        for _, feditem := range configData.Federation {
                configMap.Federation[feditem.Name] = feditem
        }

        return configMap
}

*/
