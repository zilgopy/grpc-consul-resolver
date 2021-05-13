package main

import (
	_ "consul_resolver"
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	pb "grpc_kv/grpc_kv_shared"
	"log"
	"os"
	"time"
)

func main() {
	conn,err:=grpc.Dial(os.Getenv("TARGET"),grpc.WithInsecure(),grpc.WithBlock(),grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": "%s"}`, roundrobin.Name)))
	if err!=nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	client:=pb.NewKvserverClient(conn)

	for true {
		v,err:=client.Getkv(context.Background(),&pb.Key{Key: "server"})
		fmt.Println(v,err)

		time.Sleep(time.Second)
	}
}
