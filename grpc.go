package kcache

import (
	"kcache/consistenthash"
	pb "kcache/kcachepb"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	defaultReplicas = 50
)

var (
	defaultEtcdConfig = clientv3.Config{
		Endpoints:   []string{"localhost:2379"},
		DialTimeout: 5 * time.Second,
	}
)

// Server 和 Group 是解耦合的 所以server要自己实现并发控制
type Server struct {
	pb.UnimplementedGroupCacheServer
	self       string
	status     bool
	stopSignal chan error
	mu         sync.Mutex
	peers      *consistenthash.Map
	clients    map[string]*Client
}

// Client 模块实现kcache访问其他远程节点,从而获取缓存的能力
type Client struct {
	baseURL string // 服务名称 kcache/ip:addr
}
