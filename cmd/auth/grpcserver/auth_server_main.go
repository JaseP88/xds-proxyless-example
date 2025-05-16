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
	"os"

	"github.com/JaseP88/xds-poc/api/auth"
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

// server implements auth service.
type authServer struct {
	auth.UnimplementedAuthServer
}

func (s *authServer) DualMessage(stream auth.Auth_DualMessageServer) error {
	var bigerr error
out:
	for {
		req, err := stream.Recv()
		if err != nil {
			fmt.Println("failed to receive request", err)
			bigerr = err
			break out
		}
		hn, _ := os.Hostname()
		counter++
		fmt.Printf("received request %v, distribution:%f %%", req, float64(counter)/float64(req.TransactionCounter)*100)
		resp := auth.AuthResponse{
			Result:     hn,
			ResPayload: req.GetReqPayload(),
		}
		stream.Send(&resp)
	}
	return bigerr
}

func (s *authServer) DualMessageRequestResponse(_ context.Context, request *auth.AuthRequest) (*auth.AuthResponse, error) {
	counter++
	fmt.Printf("Received rpc request %v, with distribution: %f %%", request, float64(counter)/float64(request.TransactionCounter)*100)
	resp := &auth.AuthResponse{
		Result:     serverName,
		ResPayload: request.GetReqPayload(),
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

	authAddy := fmt.Sprintf("%s:%d", address, port)
	authLis, err := net.Listen("tcp4", authAddy)
	if err != nil {
		log.Fatalf("net.Listen(tcp4, %q) failed: %v", authAddy, err)
	}

	// xdsclient within server
	as, err := xds.NewGRPCServer(grpc.Creds(creds))
	// as := grpc.NewServer(grpc.Creds(creds))
	if err != nil {
		log.Fatalf("Failed to create an auth gRPC server: %v", err)
	}
	auth.RegisterAuthServer(as, &authServer{})

	healthAddy := fmt.Sprintf("%s:%d", address, port+1)
	healthLis, err := net.Listen("tcp4", healthAddy)
	if err != nil {
		log.Fatalf("failed health")
	}
	hs := grpc.NewServer()
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthgrpc.RegisterHealthServer(hs, healthServer)

	log.Printf("Serving AuthService on %s and HealthService on %s", authLis.Addr().String(), healthLis.Addr().String())
	go func() {
		as.Serve(authLis)
	}()
	hs.Serve(healthLis)
}
