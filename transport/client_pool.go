package transport

import (
	"errors"
	"sync/atomic"
)

// https://groups.google.com/group/golang-nuts/msg/71c307e4d73024ce?pli=1
const maxInt = int(^uint(0) >> 1)

type ClientsPool interface {
	NextLive() (*NodeClient, error)
	NextDead() (*NodeClient, error)
	OnFailure(c *NodeClient)
	OnSuccess(c *NodeClient)
}

func NewClientsPool(urls []string, useragent string) (ClientsPool, error) {
	if len(urls) == 0 {
		return nil, errors.New("no servers available for connection")
	}

	if len(urls) == 1 {
		return &SinglePool{client: NewNodeClient(urls[0], useragent)}, nil
	}

	clients := make([]*NodeClient, 0, len(urls))

	for _, url := range urls {
		clients = append(clients, NewNodeClient(url, useragent))
	}

	return &ClusterPool{clients: clients}, nil
}

var (
	ErrNoAvailableClients = errors.New("no available clients")
)

type SinglePool struct {
	client *NodeClient
}

func (p *SinglePool) NextLive() (*NodeClient, error) {
	if atomic.LoadUint32(&p.client.status) != isLive {
		return nil, ErrNoAvailableClients
	}

	return p.client, nil
}

func (p *SinglePool) NextDead() (*NodeClient, error) {
	if atomic.LoadUint32(&p.client.status) != isDead {
		return nil, ErrNoAvailableClients
	}

	return p.client, nil
}

func (p *SinglePool) OnFailure(c *NodeClient) {
	atomic.StoreUint32(&c.status, isDead)
}

func (p *SinglePool) OnSuccess(c *NodeClient) {
	atomic.StoreUint32(&c.status, isLive)
}

type ClusterPool struct {
	clients []*NodeClient
}

func (p *ClusterPool) NextLive() (*NodeClient, error) {
	return p.next(isLive)
}

func (p *ClusterPool) NextDead() (*NodeClient, error) {
	return p.next(isDead)
}

func (p *ClusterPool) OnFailure(c *NodeClient) {
	atomic.StoreUint32(&c.status, isDead)
}

func (p *ClusterPool) OnSuccess(c *NodeClient) {
	atomic.StoreUint32(&c.status, isLive)
}

func (p *ClusterPool) next(status uint32) (*NodeClient, error) {
	clients := p.clients

	var (
		minC *NodeClient
		minR = maxInt
		minT = maxInt
	)

	for _, client := range clients {
		if atomic.LoadUint32(&client.status) != status {
			continue
		}

		r := client.PendingRequests()
		t := client.LastUseTime()

		if r < minR || (r == minR && t < minT) {
			minC = client
			minR = r
			minT = t
		}
	}

	if minC == nil {
		return nil, ErrNoAvailableClients
	}

	return minC, nil
}
