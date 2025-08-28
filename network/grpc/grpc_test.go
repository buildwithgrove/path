package grpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	s := grpc.NewServer()
	// Register your service here
	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()
}

func Test_connectGRPC(t *testing.T) {
	tests := []struct {
		name   string
		config GRPCConfig
	}{
		{
			name: "should connect with insecure",
			config: GRPCConfig{
				HostPort:          "bufnet",
				Insecure:          true,
				BackoffBaseDelay:  1 * time.Second,
				BackoffMaxDelay:   5 * time.Second,
				MinConnectTimeout: 20 * time.Second,
				KeepAliveTime:     10 * time.Second,
				KeepAliveTimeout:  5 * time.Second,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			c := require.New(t)
			if test.config.HostPort == "bufnet" {
				test.config.HostPort = lis.Addr().String()
			}
			conn, err := ConnectGRPC(test.config)
			c.NoError(err)

			if conn != nil {
				conn.Close()
			}
		})
	}
}
