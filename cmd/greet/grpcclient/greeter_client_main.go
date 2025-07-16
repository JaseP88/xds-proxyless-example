// /*
//  *
//  * Copyright 2020 gRPC authors.
//  *
//  * Licensed under the Apache License, Version 2.0 (the "License");
//  * you may not use this file except in compliance with the License.
//  * You may obtain a copy of the License at
//  *
//  *     http://www.apache.org/licenses/LICENSE-2.0
//  *
//  * Unless required by applicable law or agreed to in writing, software
//  * distributed under the License is distributed on an "AS IS" BASIS,
//  * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  * See the License for the specific language governing permissions and
//  * limitations under the License.
//  *
//  */

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/JaseP88/xds-poc/api/greeter"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	xdscreds "google.golang.org/grpc/credentials/xds"
	_ "google.golang.org/grpc/xds" // To install the xds resolvers and balancers.
)

var (
	target           string
	xdsCreds         bool
	transactionCount int64
	clientName       string
)

var names = []string{
	"Alice", "Bob", "Charlie", "David", "Eve", "Frank", "Grace", "Hannah", "Ivan", "Julia",
	"Kevin", "Laura", "Mike", "Nina", "Oscar", "Paul", "Quincy", "Rachel", "Steve", "Tina",
	"Uma", "Victor", "Wendy", "Xander", "Yara", "Zane", "Aaron", "Bella", "Cody", "Diana",
	"Elena", "Felix", "Gina", "Henry", "Isabel", "Jack", "Kara", "Liam", "Mona", "Noah",
	"Olivia", "Peter", "Queen", "Rita", "Sam", "Tracy", "Ulysses", "Vera", "Will", "Xenia",
	"Yusuf", "Zoey", "Amber", "Blake", "Carmen", "Derek", "Elsa", "Finn", "Gloria", "Harvey",
	"Ingrid", "Jonas", "Kelsey", "Leon", "Megan", "Nate", "Opal", "Preston", "Quinn", "Rosa",
	"Shawn", "Tara", "Uri", "Val", "Wade", "Ximena", "Yvette", "Zack", "Anya", "Brent",
	"Cara", "Damon", "Erin", "Freya", "Gavin", "Hazel", "Irene", "Jared", "Kurt", "Lena",
	"Martin", "Nora", "Omar", "Penny", "Quentin", "Rex", "Sophie", "Tom", "Ulric", "Violet",
}

func init() {
	flag.StringVar(&target, "t", "xds:///connect.me.to.grpcserver", "uri of the Greeter Server, e.g. 'xds:///helloworld-service:8080'")
	flag.StringVar(&clientName, "c", "client123", "client name")
	flag.BoolVar(&xdsCreds, "xds_creds", true, "whether the server should use xDS APIs to receive security configuration")
	flag.Int64Var(&transactionCount, "tc", 1000, "number of transactions to send")
}

func main() {
	flag.Parse()

	if !strings.HasPrefix(target, "xds:///") {
		log.Fatalf("-target must use a URI with scheme set to 'xds'")
	}

	creds := insecure.NewCredentials()
	if xdsCreds {
		log.Println("Using xDS credentials...")
		var err error
		if creds, err = xdscreds.NewClientCredentials(xdscreds.ClientOptions{FallbackCreds: insecure.NewCredentials()}); err != nil {
			log.Fatalf("failed to create client-side xDS credentials: %v", err)
		}
	}
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatalf("grpc.NewClient(%s) failed: %v", target, err)
	}
	defer conn.Close()

	client := greeter.NewGreeterClient(conn)

	counter := 0
	for i := 0; i < int(transactionCount); i++ {
		counter++
		req := &greeter.GreetRequest{
			Name:               names[rand.Intn(100)],
			FromClient:         clientName,
			TransactionCounter: int64(counter),
		}

		// res, _ := client.SayHello(context.Background(), req)
		res, _ := client.SayHelloInVietnamese(context.Background(), req)
		fmt.Printf("got res: %v", res)

		time.Sleep(100 * time.Millisecond)
	}

	for {
	}
}
