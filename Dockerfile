FROM centos:latest
RUN yum -y update && yum install -y openssh-server passwd java-11-openjdk git && echo root | passwd root --stdin 
RUN ssh-keygen -t rsa -f /etc/ssh/ssh_host_rsa_key && ssh-keygen -t ed25519 -f /etc/ssh/ssh_host_ed25519_key  && ssh-keygen -t ecdsa -f /etc/ssh/ssh_host_ecdsa_key
EXPOSE 22
ENTRYPOINT /usr/sbin/sshd


FROM golang as restapi
ENV GO111MODULE="off"
RUN go get github.com/hashicorp/consul/api
WORKDIR /root/go/src/restapi/
COPY server.go .
RUN go build server.go
ENTRYPOINT /root/go/src/restapi/server



FROM golang as client
ENV GO111MODULE="off"
RUN go get github.com/hashicorp/consul/api
WORKDIR /root/go/src/restapi/
COPY client.go .
RUN go build client.go
ENTRYPOINT /root/go/src/restapi/client



from golang as kv_client
ENV GO111MODULE="off"
RUN go get google.golang.org/grpc
RUN go get github.com/hashicorp/consul/api
WORKDIR /go/src/
COPY consul_resolver ./consul_resolver
WORKDIR /go/src/grpc_kv
COPY client.go .
COPY grpc_kv_shared ./grpc_kv_shared
RUN go build client.go
ENTRYPOINT /go/src/grpc_kv/client


from golang as kv_server
ENV GO111MODULE="off"
RUN go get google.golang.org/grpc
RUN go get github.com/hashicorp/consul/api
WORKDIR /go/src/grpc_kv
COPY server.go .
COPY grpc_kv_shared ./grpc_kv_shared
RUN go build server.go
ENTRYPOINT /go/src/grpc_kv/server




