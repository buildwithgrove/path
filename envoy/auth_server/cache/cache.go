// The cache package contains the implementation of an in-memory cache that stores
// GatewayEndpoints and their associated data from the connected Postgres database.
// It fetches this data from the remote gRPC server through an initial cache update
// on startup, then listens for updates from the remote gRPC server to update the cache.
package cache

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/auth-server/proto"
)

// EndpointDataCache is an in-memory cache that stores gateway endpoints and their associated data.
type EndpointDataCache struct {
	grpcClient proto.GatewayEndpointsClient

	gatewayEndpoints   map[string]*proto.GatewayEndpoint
	gatewayEndpointsMu sync.RWMutex

	logger polylog.Logger
}

// NewEndpointDataCache creates a new endpoint data cache, which stores GatewayEndpoints in memory for fast access.
// It initializes the cache by requesting data from a remote gRPC server and listens for updates from the remote server to update the cache.
func NewEndpointDataCache(ctx context.Context, grpcClient proto.GatewayEndpointsClient, logger polylog.Logger) (*EndpointDataCache, error) {
	cache := &EndpointDataCache{
		grpcClient: grpcClient,

		gatewayEndpoints:   make(map[string]*proto.GatewayEndpoint),
		gatewayEndpointsMu: sync.RWMutex{},

		logger: logger.With("component", "endpoint_data_cache"),
	}

	// Initialize the cache with the GatewayEndpoints from the remote gRPC server.
	if err := cache.initializeCacheFromRemote(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to set cache: %w", err)
	}

	// Start listening for updates from the remote gRPC server.
	go cache.listenForRemoteUpdates(ctx)

	return cache, nil
}

// GetGatewayEndpoint returns a GatewayEndpoint from the cache and a bool indicating if it exists in the cache.
func (c *EndpointDataCache) GetGatewayEndpoint(endpointID string) (*proto.GatewayEndpoint, bool) {
	c.gatewayEndpointsMu.RLock()
	defer c.gatewayEndpointsMu.RUnlock()

	gatewayEndpoint, ok := c.gatewayEndpoints[endpointID]
	return gatewayEndpoint, ok
}

// initializeCacheFromRemote requests the initial data from the remote gRPC server to set the cache.
func (c *EndpointDataCache) initializeCacheFromRemote(ctx context.Context) error {
	gatewayEndpointsResponse, err := c.grpcClient.GetInitialData(ctx, &proto.InitialDataRequest{})
	if err != nil {
		return fmt.Errorf("failed to get initial data from remote server: %w", err)
	}

	c.gatewayEndpointsMu.Lock()
	defer c.gatewayEndpointsMu.Unlock()
	c.gatewayEndpoints = gatewayEndpointsResponse.GetEndpoints()

	return nil
}

// listenForRemoteUpdates listens for updates from the remote gRPC server and updates the cache accordingly.
// Updates will be one of three cases:
// 3. A new GatewayEndpoint was created
// 1. An existing GatewayEndpoint was updated
// 2. An existing GatewayEndpoint was deleted
func (c *EndpointDataCache) listenForRemoteUpdates(ctx context.Context) {
	for {
		if err := c.connectAndProcessUpdates(ctx); err != nil {
			c.logger.Error().Err(err).Msg("error in update stream, retrying")
			<-time.After(time.Second * 2)
		}
	}
}

func (c *EndpointDataCache) connectAndProcessUpdates(ctx context.Context) error {
	stream, err := c.grpcClient.StreamUpdates(ctx, &proto.UpdatesRequest{})
	if err != nil {
		return fmt.Errorf("failed to stream updates from remote server: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("context cancelled, stopping update stream")
			return nil
		default:
			update, err := stream.Recv()
			if err == io.EOF {
				c.logger.Info().Msg("update stream ended, attempting to reconnect")
				return nil // Return to trigger a reconnection
			}
			if err != nil {
				return fmt.Errorf("error receiving update: %w", err)
			}
			if update == nil {
				c.logger.Error().Msg("received nil update")
				continue
			}

			c.gatewayEndpointsMu.Lock()
			if update.Delete {
				delete(c.gatewayEndpoints, update.EndpointId)
				c.logger.Info().Str("endpoint_id", update.EndpointId).Msg("deleted gateway endpoint")
			} else {
				c.gatewayEndpoints[update.EndpointId] = update.GatewayEndpoint
				c.logger.Info().Str("endpoint_id", update.EndpointId).Msg("updated gateway endpoint")
			}
			c.gatewayEndpointsMu.Unlock()
		}
	}
}
