package tgbotapi

// https://github.com/rs/dnscache

import (
	"context"
	"net"
	"sync"
	"time"
)

type CacheResolver struct {
	lock     sync.RWMutex
	cache    map[string][]net.IP
	resolver *net.Resolver
}

func NewCachedResolver(refreshRate time.Duration) *CacheResolver {
	return NewCustomCachedResolver(net.DefaultResolver, refreshRate)
}

func NewCustomCachedResolver(resolver *net.Resolver, refreshRate time.Duration) *CacheResolver {
	cacheResolver := &CacheResolver{
		resolver: resolver,
		cache:    make(map[string][]net.IP, 64),
	}
	if refreshRate > 0 {
		go cacheResolver.autoRefresh(refreshRate)
	}
	return cacheResolver
}

func (r *CacheResolver) Fetch(address string) ([]net.IP, error) {
	r.lock.RLock()
	ips, exists := r.cache[address]
	r.lock.RUnlock()
	if exists {
		if bot.Debug {
			log.Printf("hit DNS cache ", address, ips)
		}
		return ips, nil
	}

	return r.Lookup(address)
}

func (r *CacheResolver) FetchOne(address string) (net.IP, error) {
	ips, err := r.Fetch(address)
	if err != nil || len(ips) == 0 {
		return nil, err
	}
	return ips[0], nil
}

func (r *CacheResolver) FetchOneString(address string) (string, error) {
	ip, err := r.FetchOne(address)
	if err != nil || ip == nil {
		return "", err
	}
	return ip.String(), nil
}

func (r *CacheResolver) Refresh() {
	i := 0
	r.lock.RLock()
	addresses := make([]string, len(r.cache))
	for key, _ := range r.cache {
		addresses[i] = key
		i++
	}
	r.lock.RUnlock()

	for _, address := range addresses {
		r.Lookup(address)
		time.Sleep(time.Second * 2)
	}
}

func (r *CacheResolver) Lookup(address string) ([]net.IP, error) {
	ips, err := r.LookupIP(address)
	if err != nil {
		return nil, err
	}

	r.lock.Lock()
	r.cache[address] = ips
	r.lock.Unlock()
	return ips, nil
}

func (r *CacheResolver) autoRefresh(rate time.Duration) {
	for {
		time.Sleep(rate)
		r.Refresh()
	}
}

func (r *CacheResolver) LookupIP(host string) ([]net.IP, error) {
	address, err := r.resolver.LookupIPAddr(context.Background(), host)
	if err != nil {
		return nil, err
	}
	ips := make([]net.IP, len(address))
	for i, ia := range address {
		ips[i] = ia.IP
	}
	if bot.Debug {
		log.Println("LookupIPAddr ", host, ips)
	}

	return ips, nil
}
