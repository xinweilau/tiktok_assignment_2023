package main

import (
	"context"
	"fmt"
	"log"

	rpc "github.com/TikTokTechImmersion/assignment_demo_2023/rpc-server/kitex_gen/rpc/imservice"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/server"
	etcd "github.com/kitex-contrib/registry-etcd"
)

var rdb = &RedisClient{}

func main() {
	ctx := context.Background()

	if err := rdb.InitRedisClient(ctx, "redis:6379", ""); err != nil {
		errMsg := fmt.Sprintf("Failed to initialise Redis client, err: %v", err)
		log.Fatal(errMsg)
	}

	r, err := etcd.NewEtcdRegistry([]string{"etcd:2379"}) // r should not be reused.
	if err != nil {
		log.Fatal(err)
	}

	svr := rpc.NewServer(new(IMServiceImpl), server.WithRegistry(r), server.WithServerBasicInfo(&rpcinfo.EndpointBasicInfo{
		ServiceName: "demo.rpc.server",
	}))

	if err = svr.Run(); err != nil {
		log.Println(err.Error())
	}
}
