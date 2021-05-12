FROM golang as kv-client
ENV GO111MODULE="off"
RUN go get google.golang.org/grpc
RUN go get github.com/hashicorp/consul/api
WORKDIR /go/src/
COPY consul_resolver ./consul_resolver
WORKDIR /go/src/grpc_kv
COPY grpc_kv/client.go .
COPY grpc_kv/grpc_kv_shared ./grpc_kv_shared
RUN go build client.go
ENTRYPOINT /go/src/grpc_kv/client


FROM golang as kv-server
ENV GO111MODULE="off"
RUN go get google.golang.org/grpc
RUN go get github.com/hashicorp/consul/api
WORKDIR /go/src/grpc_kv
COPY grpc_kv/server.go .
COPY grpc_kv/grpc_kv_shared ./grpc_kv_shared
RUN go build server.go
ENTRYPOINT /go/src/grpc_kv/server




