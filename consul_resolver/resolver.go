package consul_resolver

import (
	"context"
	"fmt"
	consulapi "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/resolver"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"
)

const scheme = "consul"
const defaultInterval = "5m"
type consulBuilder struct {}
type consulResolver struct{
	ctx context.Context
	cancel context.CancelFunc
	wg sync.WaitGroup
	resolveNow chan struct{}
	svcName string
	clientCfg *consulapi.Config
	cc resolver.ClientConn
	server string
	opts url.Values
}


func (b *consulBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	svcName,Opts:=parseEndpoint(target.Endpoint)
	consulCfg:=new(consulapi.Config)
	consulCfg.Address = target.Authority
	consulCfg.Token=Opts.Get("token")
	ctx,cancel:=context.WithCancel(context.Background())
	cr:=consulResolver{
		ctx: ctx,
		cancel: cancel,
		wg: sync.WaitGroup{},
		resolveNow: make(chan struct{},1),
		svcName: svcName,
		clientCfg: consulCfg,
		cc: cc,
		server: target.Authority,
		opts: Opts,

	}
	cr.wg.Add(1)
	go cr.Watcher()
	return cr,nil
}
func (b *consulBuilder)Scheme() string{
	return scheme
}

func (r consulResolver ) ResolveNow(resolver.ResolveNowOptions)  {
	r.resolveNow<-struct{}{}
}

func (r consulResolver) Close()  {
	r.cancel()
	r.wg.Wait()
}
func (r consulResolver) Watcher()  {
	defer r.wg.Done()
	client,err:=consulapi.NewClient(r.clientCfg)
	if err!=nil {
		grpclog.Error(err)
	}
	tags:=make([]string,0)
	var passing bool
	opts:=new(consulapi.QueryOptions)
	var interval time.Duration
	if r.opts!=nil {
		tags=r.opts["tag"]
    	passing=r.opts.Get("passing")=="true" ||r.opts.Get("passing")==""
		opts.Namespace=r.opts.Get("ns")
		opts.Datacenter=r.opts.Get("dc")
		interval,err=time.ParseDuration(r.opts.Get("interval"))
		if err!=nil {
			interval,_=time.ParseDuration(defaultInterval)
		}
	}
	for  {
		svcInfo,_,err:=client.Health().ServiceMultipleTags(r.svcName,tags,passing,opts)
		if err!=nil { r.cc.ReportError(err) ;return }
		if len(svcInfo)!=0 {
		state:=resolver.State{}
			for _, info := range svcInfo {
				state.Addresses=append(state.Addresses,resolver.Address{Addr: fmt.Sprint(info.Service.Address,":",info.Service.Port)})
			}

		r.cc.UpdateState(state)
                } else {
			grpclog.Info("Resolver returns 0 available host , sleep 10 seconds for next try.")
                time.Sleep(10*time.Second)
                r.resolveNow<- struct{}{}
                }
		select {
		case <-r.ctx.Done():
			return
		case <-time.After(interval):
		case <-r.resolveNow:
		}
	}


}

func init(){
	resolver.Register(&consulBuilder{})
}

func parseEndpoint(s string) (string,url.Values) {
	svcAndOpts:=strings.SplitN(s,"?",2)
	switch len(svcAndOpts) {
	case 1:
		return svcAndOpts[0],nil
	default:
		opts,err:= url.ParseQuery(svcAndOpts[1])
		if err!=nil {
			log.Fatalln(err)
		}
		return svcAndOpts[0],opts
	}
}




