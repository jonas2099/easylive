package cache

import (
	"github.com/haroldleong/easylive/container"
)

type PackageCache struct {
	full bool
	p    *container.Packet
}

func NewPackageCache() *PackageCache {
	return &PackageCache{}
}

func (specialCache *PackageCache) Write(p *container.Packet) {
	specialCache.p = p
	specialCache.full = true
}

func (specialCache *PackageCache) Send(pChan chan *container.Packet) error {
	if !specialCache.full {
		return nil
	}
	pChan <- specialCache.p
	return nil
}
