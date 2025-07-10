package main

import (
         "flag"
        "io"
        "fmt"
        "log"
        "net"
        "net/http"
        "strings"

        "app/src/echo"

	"golang.org/x/net/context"
        "google.golang.org/grpc"
)


var (
	grpcport   = flag.Int("grpcport", 50051, "grpc port")
	servername = flag.String("servername", "server1", "server name")

	conn *grpc.ClientConn
)

const (
	address string = ":50051"
)

type server struct {
	echo.UnimplementedEchoServerServer
        serverName string
}

func isGrpcRequest(r *http.Request) bool {
	return r.ProtoMajor == 2 && strings.HasPrefix(r.Header.Get("Content-Type"), "application/grpc")
}

func (s *server) SayHello(ctx context.Context, in *echo.EchoRequest) (*echo.EchoReply, error) {

	log.Println("Got rpc: --> ", in.Name)

	return &echo.EchoReply{Message: "Hello " + in.Name + "  ----> from " + *servername}, nil
}

func (s *server) SayHelloStream(in *echo.EchoRequest, stream echo.EchoServer_SayHelloStreamServer) error {

	log.Println("Got stream:  -->  ")
	stream.Send(&echo.EchoReply{Message: "stream Hello " + in.Name})
	return nil
}

func (s *server) SayHelloBidStream(strm echo.EchoServer_SayHelloBidStreamServer) error {
    
    log.Println("BiD Got stream:  -->  ")
    
    for {
		// receive data from stream
		req, err := strm.Recv()
		if err == io.EOF {
			// return will close stream from server side
			log.Println("SayHelloBidStream exit - EOF")
			return nil
		}
		if err != nil {
			log.Printf("SayHelloBidStream received error %v", err)
			continue
		}

                //Process request
                log.Println("BiD received ->", req.Name)
                if req.Name == "Bret" {
                    log.Printf("SayHelloBidStream send hello to Bret")
                    strm.Send(&echo.EchoReply{Message: "Hello " + req.Name + "  ----> from " + *servername})
                } else {
                    continue
                }
    }
                
    return nil
}


func getnonloopback_address() string {
    addrs, err := net.InterfaceAddrs()
    if err != nil {
        return ""
    }
    for _, address := range addrs {
        // check the address type and if it is not a loopback the display it
        if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
            if ipnet.IP.To4() != nil {
                return ipnet.IP.String()
            }
        }
    }
    return ""
}


func main() {

	flag.Parse()

        ip_addr := getnonloopback_address()
        if ip_addr == ""{
            log.Fatalf("failed to get ip address from getnonloopback_address")
        }
        log.Println("Got IP:%s", ip_addr)

       

        lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", ip_addr, *grpcport))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	echoServer := grpc.NewServer()


	echo.RegisterEchoServerServer(echoServer, &server{})
	if err := echoServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

