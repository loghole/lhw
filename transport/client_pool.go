package transport

import (
	"crypto/tls"
	"errors"
	"net/http"
	"sync/atomic"
)

// https://groups.google.com/group/golang-nuts/msg/71c307e4d73024ce?pli=1
const maxInt = int(^uint(0) >> 1)

var (
	ErrNoAvailableClients = errors.New("no available clients")
	ErrNoAvailableServers = errors.New("no servers available for connection")
)

type ClientsPool interface {
	NextLive() (*NodeClient, error)
	NextDead() (*NodeClient, error)
	OnFailure(c *NodeClient)
	OnSuccess(c *NodeClient)
}

func NewClientsPool(servers []string, insecure bool) (pool ClientsPool, err error) {
	transport := &http.Transport{
		ForceAttemptHTTP2: true,
	}

	if insecure {
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // nolint:gosec // skip.
		}
	}

	if len(servers) == 0 {
		return nil, ErrNoAvailableServers
	}

	clients := make([]*NodeClient, len(servers))

	for idx, server := range servers {
		clients[idx], err = NewNodeClient(server, transport)
		if err != nil {
			return nil, err
		}
	}

	if len(clients) == 1 {
		return &SinglePool{client: clients[0]}, nil
	}

	return &ClusterPool{clients: clients}, nil
}

type SinglePool struct {
	client *NodeClient
}

func (p *SinglePool) NextLive() (*NodeClient, error) {
	if atomic.LoadInt32(&p.client.status) != isLive {
		return nil, ErrNoAvailableClients
	}

	return p.client, nil
}

func (p *SinglePool) NextDead() (*NodeClient, error) {
	if atomic.LoadInt32(&p.client.status) != isDead {
		return nil, ErrNoAvailableClients
	}

	return p.client, nil
}

func (p *SinglePool) OnFailure(c *NodeClient) {
	atomic.StoreInt32(&c.status, isDead)
}

func (p *SinglePool) OnSuccess(c *NodeClient) {
	atomic.StoreInt32(&c.status, isLive)
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
	atomic.StoreInt32(&c.status, isDead)
}

func (p *ClusterPool) OnSuccess(c *NodeClient) {
	atomic.StoreInt32(&c.status, isLive)
}

func (p *ClusterPool) next(status int32) (*NodeClient, error) {
	clients := p.clients

	var (
		minC *NodeClient
		minR = maxInt
		minT = maxInt
	)

	for _, client := range clients {
		if atomic.LoadInt32(&client.status) != status {
			continue
		}

		r := client.ActiveRequests()
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
