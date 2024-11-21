// The endpointstore package contains the implementation of an in-memory store that stores
// GatewayEndpoints and their associated data from the connected Postgres database.
// It fetches this data from the remote gRPC server through an initial store update
// on startup, then listens for updates from the remote gRPC server to update the store.
package endpointstore

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/envoy/auth_server/proto"
)

// EndpointStore is an in-memory store that stores gateway endpoints and their associated data.
type EndpointStore struct {
	grpcClient proto.GatewayEndpointsClient

	gatewayEndpoints   map[string]*proto.GatewayEndpoint
	gatewayEndpointsMu sync.RWMutex

	logger polylog.Logger
}

// NewEndpointStore creates a new endpoint store, which stores GatewayEndpoints in memory for fast access.
// It initializes the store by requesting data from a remote gRPC server and listens for updates from the remote server to update the store.
func NewEndpointStore(ctx context.Context, grpcClient proto.GatewayEndpointsClient, logger polylog.Logger) (*EndpointStore, error) {
	store := &EndpointStore{
		grpcClient: grpcClient,

		gatewayEndpoints:   make(map[string]*proto.GatewayEndpoint),
		gatewayEndpointsMu: sync.RWMutex{},

		logger: logger.With("component", "endpoint_data_store"),
	}

	// Initialize the endpoint store with the GatewayEndpoints from the remote gRPC server.
	if err := store.initializeStoreFromRemote(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to set store: %w", err)
	}

	// Start listening for updates from the remote gRPC server.
	go store.listenForRemoteUpdates(ctx)

	return store, nil
}

// GetGatewayEndpoint returns a GatewayEndpoint from the store and a bool indicating if it exists in the store.
func (c *EndpointStore) GetGatewayEndpoint(endpointID string) (*proto.GatewayEndpoint, bool) {
	c.gatewayEndpointsMu.RLock()
	defer c.gatewayEndpointsMu.RUnlock()

	gatewayEndpoint, ok := c.gatewayEndpoints[endpointID]
	return gatewayEndpoint, ok
}

// initializeStoreFromRemote requests the initial data from the remote gRPC server to set the store.
func (c *EndpointStore) initializeStoreFromRemote(ctx context.Context) error {
	gatewayEndpointsResponse, err := c.grpcClient.FetchAuthDataSync(ctx, &proto.AuthDataRequest{})
	if err != nil {
		return fmt.Errorf("failed to get initial data from remote server: %w", err)
	}

	c.gatewayEndpointsMu.Lock()
	defer c.gatewayEndpointsMu.Unlock()
	c.gatewayEndpoints = gatewayEndpointsResponse.GetEndpoints()

	return nil
}

// listenForRemoteUpdates listens for updates from the remote gRPC server and updates the store accordingly.
// Updates will be one of three cases:
// 3. A new GatewayEndpoint was created
// 1. An existing GatewayEndpoint was updated
// 2. An existing GatewayEndpoint was deleted
func (c *EndpointStore) listenForRemoteUpdates(ctx context.Context) {
	for {
		// TODO_IMPROVE(@commoddity): improve the reconnection logic to better handle the
		// remote server restarting or other connection issues that may arise.
		if err := c.connectAndProcessUpdates(ctx); err != nil {
			c.logger.Error().Err(err).Msg("error in update stream, retrying")
			<-time.After(time.Second * 2)
		}
	}
}

// connectAndProcessUpdates connects to the remote gRPC server and processes updates from the server.
func (c *EndpointStore) connectAndProcessUpdates(ctx context.Context) error {
	stream, err := c.grpcClient.StreamAuthDataUpdates(ctx, &proto.AuthDataUpdatesRequest{})
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