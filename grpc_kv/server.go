package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/wrapperspb"
	pb "grpc_kv/grpc_kv_shared"
	"log"
	"net"
	"os"
	"strconv"
)

type kvserver struct {
	m map[string]string
}

func (s *kvserver) Getkv(ctx context.Context, key *pb.Key) (*pb.Value, error) {
	value,ok:=s.m[key.Key]
	log.Println(value)
	if ok {
		return &pb.Value{
			Value: value,
		},nil
	}
	return nil,errors.New("key is not existing in server")
}

func (s *kvserver) Putkv(ctx context.Context, kvpair *pb.Kvpair) (*wrapperspb.BoolValue, error) {
	     s.m[kvpair.Key]=kvpair.Value
	     pbBool:=new(wrapperspb.BoolValue)
	     pbBool.Value=true
	     return pbBool,nil
}

func (s *kvserver) Delkv(ctx context.Context, key *pb.Key) (*wrapperspb.BoolValue, error) {
	pbBool:=new(wrapperspb.BoolValue)
	_,ok:=s.m[key.Key]
	if ok {
		delete(s.m,key.Key)
		pbBool.Value=true
		return pbBool,nil
	}
	pbBool.Value=false
	return pbBool,errors.New("key is not existing in server")
}

func main() {
	registerservices()
	s:=new(kvserver)
	s.m=make(map [string]string)
	s.m["server"]=os.Getenv("SERVICE_ID")
	listener,err:=net.Listen("tcp",":"+os.Getenv("PORT"))
	if err!=nil {log.Fatalln(err)}
	defer listener.Close()

	grpcServer:=grpc.NewServer()
	pb.RegisterKvserverServer(grpcServer,s)

	if err:=grpcServer.Serve(listener);err!=nil {
		log.Fatalln(err)
	}

}

func registerservices(){
	client,err:=api.NewClient(api.DefaultConfig())
	if err!=nil {
		panic(err)
	}
	sr:=new(api.AgentServiceRegistration)
	sr.Name=os.Getenv("SERVICE_NAME")
	sr.ID=os.Getenv("SERVICE_ID")
	sr.Address=sr.ID
	sr.Tags=[]string{os.Getenv("TAG")}
	port,err:=strconv.Atoi(os.Getenv("PORT"))
	if err!=nil {panic(err)}
	sr.Port=port
	sr.Check=new(api.AgentServiceCheck)
	sr.Check.TCP=fmt.Sprintf("%s:%v",sr.Address,port)
	sr.Check.Interval="10s"
	err=client.Agent().ServiceRegister(sr)
	if err!=nil{
		panic(err)
	}
}