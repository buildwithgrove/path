package morse

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pokt-foundation/pocket-go/provider"
	sdkrelayer "github.com/pokt-foundation/pocket-go/relayer"
	"github.com/pokt-network/poktroll/pkg/polylog"

	"github.com/buildwithgrove/path/relayer"
)

// relayer package's Protocol interface is fulfilled by the Protocol struct
// below using Morse-specific methods.
var _ relayer.Protocol = &Protocol{}

// TODO_TECHDEBT: Make this configurable via an env variable.
const defaultRelayTimeoutMillisec = 5000

// OffChainBackend allows enhancing an onchain application with extra fields that are required to sign/send relays.
// This is used to supply AAT data to a Morse application, which is needed for sending relays on behalf of the application.
type OffChainBackend interface {
	// GetSignedAAT returns the AAT created by AppID offchain
	GetSignedAAT(appID relayer.AppAddr) (provider.PocketAAT, bool)
}

// FullNode defines the functionality expected by the Protocol struct
// from a Morse full node.
type FullNode interface {
	GetAllApps(context.Context) ([]provider.App, error)
	GetSession(ctx context.Context, chainID, appPublicKey string) (provider.Session, error)
	SendRelay(context.Context, *sdkrelayer.Input) (*sdkrelayer.Output, error)
}

// TODO_UPNEXT(@adshmh): Add unit/E2E tests for the implementation of the Morse relayer.
func NewProtocol(ctx context.Context, fullNode FullNode, offChainBackend OffChainBackend) (*Protocol, error) {
	protocol := &Protocol{
		fullNode:        fullNode,
		offChainBackend: offChainBackend,
		logger:          polylog.Ctx(ctx),
	}

	go func() {
		// TODO_IMPROVE: make the refresh interval configurable.
		ticker := time.NewTicker(time.Minute)
		for {
			protocol.updateAppCache()
			protocol.updateSessionCache()

			<-ticker.C
		}
	}()

	return protocol, nil
}

type Protocol struct {
	logger polylog.Logger

	fullNode        FullNode
	offChainBackend OffChainBackend

	appCache   map[relayer.ServiceID][]app
	appCacheMu sync.RWMutex
	// TODO_IMPROVE: Add a sessionCacheKey type with the necessary helpers to concat a key
	// sessionCache caches sessions for use by the Relay function.
	// map keys are of the format "serviceID-appID"
	sessionCache   map[string]provider.Session
	sessionCacheMu sync.RWMutex
}

func (p *Protocol) Endpoints(serviceID relayer.ServiceID) (map[relayer.AppAddr][]relayer.Endpoint, error) {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()

	apps, found := p.appCache[serviceID]
	if !found {
		return nil, fmt.Errorf("Endpoints: no apps found for service %s", serviceID)
	}

	endpoints := make(map[relayer.AppAddr][]relayer.Endpoint)
	for _, app := range apps {
		session, found := p.getSession(serviceID, app.address)
		if !found {
			p.logger.Warn().
				Str("service", string(serviceID)).
				Str("address", app.address).
				Msg("Endpoints: no session found for service/app combination. Skipping.")

			continue
		}

		endpoints[app.Addr()] = endpointsFromSession(session)
	}

	if len(endpoints) == 0 {
		return nil, fmt.Errorf("Endpoints: no sessions found for service %s", serviceID)
	}

	return endpoints, nil
}

func (p *Protocol) SendRelay(req relayer.Request) (relayer.Response, error) {
	app, found := p.getApp(req.ServiceID, req.AppAddr)
	if !found {
		return relayer.Response{}, fmt.Errorf("relay: service %s app %s not found", req.ServiceID, req.AppAddr)
	}

	session, found := p.getSession(req.ServiceID, app.address)
	if !found {
		return relayer.Response{}, fmt.Errorf("relay: session not found for service %s app %s", req.ServiceID, req.AppAddr)
	}

	endpoint, err := getEndpoint(session, req.EndpointAddr)
	if err != nil {
		return relayer.Response{}, fmt.Errorf("relay: error getting node %s for service %s app %s", req.EndpointAddr, req.ServiceID, req.AppAddr)
	}

	output, err := p.sendRelay(
		string(req.ServiceID),
		endpoint,
		session,
		app.aat,
		// TODO_IMPROVE: chain-specific timeouts
		0, // SDK to use the default timeout.
		req.Payload,
	)

	return relayer.Response{
		Bytes:          []byte(output.Response),
		HTTPStatusCode: output.StatusCode,
	}, err
}

func (p *Protocol) getApp(serviceID relayer.ServiceID, appAddr relayer.AppAddr) (app, bool) {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()

	apps, found := p.appCache[serviceID]
	if !found {
		return app{}, false
	}

	for _, app := range apps {
		if app.Addr() == appAddr {
			return app, true
		}
	}

	return app{}, false
}

func (p *Protocol) getSession(serviceID relayer.ServiceID, appAddr string) (provider.Session, bool) {
	p.sessionCacheMu.RLock()
	defer p.sessionCacheMu.RUnlock()

	session, found := p.sessionCache[sessionCacheKey(serviceID, appAddr)]
	return session, found
}

func (p *Protocol) updateAppCache() {
	appData := p.fetchAppData()

	if len(appData) == 0 {
		p.logger.Warn().Msg("updateAppCache: received an empty app list; skipping update")
		return
	}

	p.appCacheMu.Lock()
	defer p.appCacheMu.Unlock()
	p.appCache = appData
}

func (p *Protocol) fetchAppData() map[relayer.ServiceID][]app {
	onchainApps, err := p.fullNode.GetAllApps(context.Background())
	if err != nil {
		p.logger.Warn().
			Err(err).
			Msg("fetchAppData: error getting list of onchain applications")

		return nil
	}

	appData := make(map[relayer.ServiceID][]app)
	for _, onchainApp := range onchainApps {
		if len(onchainApp.Chains) == 0 {
			p.logger.Warn().
				Str("publicKey", onchainApp.PublicKey).
				Msg("fetchAppData: app has no chains specified onchain. Skipping the app.")

			continue
		}

		// TODO_IMPROVE: validate the AAT received from the offChainBackend
		signedAAT, ok := p.offChainBackend.GetSignedAAT(relayer.AppAddr(onchainApp.Address))
		if !ok {
			p.logger.Warn().Str("publicKey", onchainApp.PublicKey).Msg("fetchAppData: no AAT found for app. Skipping the app.")

			continue
		}
		app := app{
			address:   onchainApp.Address,
			publicKey: onchainApp.PublicKey,
			aat:       signedAAT,
		}

		for _, chainID := range onchainApp.Chains {
			serviceID := relayer.ServiceID(chainID)
			appData[serviceID] = append(appData[serviceID], app)
		}
	}

	return appData
}

func (p *Protocol) updateSessionCache() {
	sessions := p.fetchSessions()
	if len(sessions) == 0 {
		p.logger.Warn().Msg("updateSessionCache: received empty session list; skipping update.")
		return
	}

	p.sessionCacheMu.Lock()
	defer p.sessionCacheMu.Unlock()
	p.sessionCache = sessions
}

func (p *Protocol) fetchSessions() map[string]provider.Session {
	p.appCacheMu.RLock()
	defer p.appCacheMu.RUnlock()

	sessions := make(map[string]provider.Session)
	// TODO_TECHDEBT: use multiple go routines.
	for serviceID, apps := range p.appCache {
		for _, app := range apps {
			// NOTE: We use the application's public key here because that is what Morse full nodes require to return a session,
			// but we use an application's address to cache it and its corresponding session(s).
			session, err := p.fullNode.GetSession(context.Background(), string(serviceID), app.publicKey)
			if err != nil {
				p.logger.Warn().
					Err(err).
					Str("service", string(serviceID)).
					Str("appPublicKey", string(app.publicKey)).
					Msg("fetchSessions: error getting a session")

				continue
			}
			sessions[sessionCacheKey(serviceID, app.address)] = session
		}
	}

	return sessions
}

func (p *Protocol) sendRelay(
	chainID string,
	node provider.Node,
	session provider.Session,
	aat provider.PocketAAT,
	timeoutMillisec int,
	payload relayer.Payload,
) (provider.RelayOutput, error) {
	fullNodeInput := &sdkrelayer.Input{
		Blockchain: chainID,
		Node:       &node,
		Session:    &session,
		PocketAAT:  &aat,
		Data:       payload.Data,
		Method:     payload.Method,
		Path:       payload.Path,
	}

	timeout := timeoutMillisec
	if timeout == 0 {
		timeout = defaultRelayTimeoutMillisec
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Millisecond)
	defer cancel()

	output, err := p.fullNode.SendRelay(ctx, fullNodeInput)
	if output.RelayOutput == nil {
		return provider.RelayOutput{}, fmt.Errorf("relay: received null output from the SDK")
	}

	// TODO_DISCUSS: do we need to verify the node/proof structs?
	return *output.RelayOutput, err
}

func sessionCacheKey(serviceID relayer.ServiceID, appAddr string) string {
	return fmt.Sprintf("%s-%s", serviceID, appAddr)
}
