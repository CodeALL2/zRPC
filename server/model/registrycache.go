package model

import (
	"fmt"
	"net"
	"sync"
)

type RegistryCache struct {
	mateInfoCache     map[string][]*ServiceMetaInfo //这是一个存储etcd的缓存信息
	serverClientCache map[string]net.Conn           //存储etcd对应的服务端连接对象
	mateInfoLock      sync.RWMutex
	serverClientLock  sync.RWMutex
}

func NewRegistryCache() *RegistryCache {
	return &RegistryCache{
		mateInfoCache:     make(map[string][]*ServiceMetaInfo),
		serverClientCache: make(map[string]net.Conn),
	}
}

func (r *RegistryCache) WriteCacheToMateInfoCache(key string, value []*ServiceMetaInfo) {
	r.mateInfoLock.Lock()
	defer r.mateInfoLock.Unlock()
	r.mateInfoCache[key] = value
}

func (r *RegistryCache) WriteCacheToServerClientCache(key string, value net.Conn) {

	r.serverClientLock.Lock()
	defer r.serverClientLock.Unlock()
	r.serverClientCache[key] = value
}

func (r *RegistryCache) ReadCacheFromServerClientCache(key string) net.Conn {
	r.serverClientLock.RLock()
	defer r.serverClientLock.RUnlock()
	conn := r.serverClientCache[key]
	if conn == nil {
		return nil
	}
	return conn
}

func (r *RegistryCache) ReadCacheFromMateInfoCache(key string) []*ServiceMetaInfo {
	r.mateInfoLock.RLock()
	defer r.mateInfoLock.RUnlock()
	metaInfos := r.mateInfoCache[key]
	if metaInfos == nil {
		return nil
	}
	return metaInfos
}

func (r *RegistryCache) GetMateInfoCache() map[string][]*ServiceMetaInfo {
	return r.mateInfoCache
}

func (r *RegistryCache) GetServerClientCache() map[string]net.Conn {
	return r.serverClientCache
}

func (r *RegistryCache) FlushMateInfoCache(key string) {
	r.mateInfoLock.Lock()
	defer r.mateInfoLock.Unlock()
	fmt.Println("键值", key, "已清空")
	delete(r.mateInfoCache, key)
}

func (r *RegistryCache) FlushServerClientCache(key string) {
	r.serverClientLock.Lock()
	defer r.serverClientLock.Unlock()
	delete(r.serverClientCache, key)
}
