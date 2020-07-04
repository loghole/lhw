package transport

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewClientsPool(t *testing.T) {
	tests := []struct {
		name        string
		urls        []string
		wantErr     bool
		expectedErr string
		expectedRes interface{}
	}{
		{
			name:        "Error",
			urls:        []string{},
			wantErr:     true,
			expectedErr: "no servers available for connection",
			expectedRes: nil,
		},
		{
			name:        "SinglePool",
			urls:        []string{"http://127.0.0.1:9200"},
			wantErr:     false,
			expectedRes: new(SinglePool),
		},
		{
			name:        "ClusterPool",
			urls:        []string{"http://127.0.0.1:9200", "http://127.0.0.1:9201"},
			wantErr:     false,
			expectedRes: new(ClusterPool),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool, err := NewClientsPool(tt.urls, "")
			if (err != nil) != tt.wantErr {
				t.Error(err)
			}

			assert.IsType(t, tt.expectedRes, pool)

			if tt.wantErr {
				assert.EqualError(t, err, tt.expectedErr)
			}
		})
	}
}

func TestSinglePool(t *testing.T) {
	pool := SinglePool{
		client: &NodeClient{
			host:        "http://127.0.0.1:9200",
			lastUseTime: time.Now().UnixNano(),
		},
	}

	client, err := pool.NextLive()
	assert.Error(t, err, ErrNoAvailableClients.Error())
	assert.IsType(t, (*NodeClient)(nil), client)

	client, err = pool.NextDead()
	assert.Nil(t, err)
	assert.IsType(t, new(NodeClient), client)

	pool.OnSuccess(client)

	client, err = pool.NextDead()
	assert.Error(t, err, ErrNoAvailableClients.Error())
	assert.IsType(t, (*NodeClient)(nil), client)

	client, err = pool.NextLive()
	assert.Nil(t, err)
	assert.IsType(t, new(NodeClient), client)

	pool.OnFailure(client)

	client, err = pool.NextLive()
	assert.Error(t, err, ErrNoAvailableClients.Error())
	assert.IsType(t, (*NodeClient)(nil), client)

	client, err = pool.NextDead()
	assert.Nil(t, err)
	assert.IsType(t, new(NodeClient), client)
}

func TestClusterPool(t *testing.T) {
	clients := []*NodeClient{
		{
			host:        "http://127.0.0.1:9200",
			lastUseTime: time.Now().UnixNano(),
		},
		{
			host:        "http://127.0.0.1:9201",
			lastUseTime: time.Now().UnixNano(),
		},
	}

	pool := ClusterPool{clients: clients}

	client, err := pool.NextLive()
	assert.EqualError(t, err, ErrNoAvailableClients.Error())
	assert.IsType(t, (*NodeClient)(nil), client)

	client, err = pool.NextDead()
	assert.Nil(t, err)
	assert.IsType(t, new(NodeClient), client)
	assert.Equal(t, clients[0], client)

	clients[0].lastUseTime = time.Now().UnixNano()

	client, err = pool.NextDead()
	assert.Nil(t, err)
	assert.IsType(t, new(NodeClient), client)
	assert.Equal(t, clients[1], client)

	clients[1].lastUseTime = time.Now().UnixNano()

	pool.OnSuccess(clients[1])

	client, err = pool.NextLive()
	assert.Nil(t, err)
	assert.IsType(t, new(NodeClient), client)
	assert.Equal(t, clients[1], client)

	clients[1].lastUseTime = time.Now().UnixNano()

	pool.OnSuccess(clients[0])

	client, err = pool.NextLive()
	assert.Nil(t, err)
	assert.IsType(t, new(NodeClient), client)
	assert.Equal(t, clients[0], client)

	pool.OnFailure(clients[1])

	client, err = pool.NextDead()
	assert.Nil(t, err)
	assert.IsType(t, new(NodeClient), client)
	assert.Equal(t, clients[1], client)
}

func BenchmarkClusterPool_NextLive(b *testing.B) {
	b.StopTimer()

	pool, err := NewClientsPool(
		[]string{"http://127.0.0.1:9200",
			"http://127.0.0.1:9201",
			"http://127.0.0.1:9202",
			"http://127.0.0.1:9203",
			"http://127.0.0.1:9204",
			"http://127.0.0.1:9205",
			"http://127.0.0.1:9206",
			"http://127.0.0.1:9207",
			"http://127.0.0.1:9208",
			"http://127.0.0.1:9209",
		},
		"test-user-agent",
	)
	if err != nil {
		b.Fatal(err)
	}

	for {
		client, err := pool.NextDead()
		if err != nil {
			break
		}

		pool.OnSuccess(client)
	}

	b.StartTimer()

	b.Run("Single", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			client, err := pool.NextLive()
			if err != nil {
				b.Fatal(err)
			}

			pool.OnSuccess(client)
		}
	})

	b.Run("Parallel (10)", func(b *testing.B) {
		b.SetParallelism(10)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				client, err := pool.NextLive()
				if err != nil {
					b.Fatal(err)
				}

				pool.OnSuccess(client)
			}
		})
	})

	b.Run("Parallel (100)", func(b *testing.B) {
		b.SetParallelism(100)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				client, err := pool.NextLive()
				if err != nil {
					b.Fatal(err)
				}

				pool.OnSuccess(client)
			}
		})
	})

	b.Run("Parallel (1000)", func(b *testing.B) {
		b.SetParallelism(1000)

		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				client, err := pool.NextLive()
				if err != nil {
					b.Fatal(err)
				}

				pool.OnSuccess(client)
			}
		})
	})
}
