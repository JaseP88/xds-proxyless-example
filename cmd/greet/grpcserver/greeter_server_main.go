/*
 *
 * Copyright 2020 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/JaseP88/xds-poc/api/greeter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"

	"google.golang.org/grpc/credentials/insecure"
	xdscreds "google.golang.org/grpc/credentials/xds"
	healthgrpc "google.golang.org/grpc/health/grpc_health_v1"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/xds" // To install the xds resolvers and balancers.
)

var (
	address    string
	port       int
	xdsCreds   bool
	serverName string
	counter    = 0
)

func init() {
	flag.StringVar(&address, "a", "127.0.0.1", "server address")
	flag.IntVar(&port, "p", 50051, "the port to serve Greeter service requests on. Health service will be served on `port+1`")
	flag.BoolVar(&xdsCreds, "xds_creds", true, "whether the server should use xDS APIs to receive security configuration")
	flag.StringVar(&serverName, "n", "server_A", "server name")
}

type greeterServer struct {
	greeter.UnimplementedGreeterServer
}

func (s *greeterServer) SayHello(_ context.Context, request *greeter.GreetRequest) (*greeter.GreetReply, error) {
	counter++
	log.Printf("Received rpc request %v, with distribution: %f %%", request, float64(counter)/float64(request.TransactionCounter)*100)
	resp := &greeter.GreetReply{
		Greet:     fmt.Sprintf("Hello %s!", request.Name),
		FromServer: serverName,
		ToClient: request.FromClient,
	}
	return resp, nil
}

func (s *greeterServer) SayHelloInVietnamese(_ context.Context, request *greeter.GreetRequest) (*greeter.GreetReply, error) {
	counter++
	log.Printf("Received rpc request %v, with distribution: %f %%", request, float64(counter)/float64(request.TransactionCounter)*100)
	resp := &greeter.GreetReply{
		Greet:     fmt.Sprintf("Xin Chao %s!", request.Name),
		FromServer: serverName,
		ToClient: request.FromClient,
	}
	return resp, nil
}

func main() {
	flag.Parse()
	creds := insecure.NewCredentials()
	if xdsCreds {
		log.Println("Using xDS credentials...")
		var err error
		if creds, err = xdscreds.NewServerCredentials(xdscreds.ServerOptions{FallbackCreds: insecure.NewCredentials()}); err != nil {
			log.Fatalf("failed to create server-side xDS credentials: %v", err)
		}
	}

	greeterAddy := fmt.Sprintf("%s:%d", address, port)
	greeterLis, err := net.Listen("tcp4", greeterAddy)
	if err != nil {
		log.Fatalf("net.Listen(tcp4, %q) failed: %v", greeterAddy, err)
	}

	// xdsclient within server
	greeterService, err := xds.NewGRPCServer(grpc.Creds(creds))
	// as := grpc.NewServer(grpc.Creds(creds))
	if err != nil {
		log.Fatalf("Failed to create an Greeter gRPC server: %v", err)
	}
	greeter.RegisterGreeterServer(greeterService, &greeterServer{})

	healthAddy := fmt.Sprintf("%s:%d", address, port+1)
	healthLis, err := net.Listen("tcp4", healthAddy)
	if err != nil {
		log.Fatalf("failed health")
	}
	hs := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthgrpc.RegisterHealthServer(hs, healthServer)

	log.Printf("Serving GreeterService on %s and HealthService on %s", greeterLis.Addr().String(), healthLis.Addr().String())
	go func() {
		greeterService.Serve(greeterLis)
	}()
	hs.Serve(healthLis)
}
