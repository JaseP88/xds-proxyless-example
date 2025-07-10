package main

import (
	"flag"
        "os" 
        "strings"
	"app/src/echo"

	"log"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
        "google.golang.org/grpc/credentials/insecure"
	_ "google.golang.org/grpc/resolver" 
	_ "google.golang.org/grpc/xds"      
)

const ()

var (
	conn *grpc.ClientConn
)

func boolFromEnv(envVar string, def bool) bool {
	if def {
		return !strings.EqualFold(os.Getenv(envVar), "false")
	}
	return strings.EqualFold(os.Getenv(envVar), "true")
}

func main() {
        log.Printf("CHECK FALL BACK")
        var XDSFallbackSupport = boolFromEnv("GRPC_EXPERIMENTAL_XDS_FALLBACK", false)
        log.Printf("FALLBACK:%t", XDSFallbackSupport)

	address := flag.String("host", "dns:///be.cluster.local:50051", "dns:///be.cluster.local:50051 or xds-experimental:///be-srv")
	flag.Parse()

        creds := insecure.NewCredentials()
        conn, err := grpc.NewClient(*address, grpc.WithTransportCredentials(creds))
        if err != nil {
		log.Fatalf("grpc.NewClient(%s) failed: %v", *address, err)
	}
	defer conn.Close()

        ctx, cancel := context.WithTimeout(context.Background(), 60*time.Minute)
        defer cancel()

	c := echo.NewEchoServerClient(conn)
	for i := 0; i < 3000; i++ {
		r, err := c.SayHello(ctx, &echo.EchoRequest{Name: "request from client"}, grpc.WaitForReady(true))
		if err != nil {
			log.Fatalf("Error during RPC call: %v", err)
		}
		log.Printf("Client: Received response- %v %v", i, r)
		time.Sleep(2 * time.Second)
	}

}
