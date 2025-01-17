package shannon

// TODO_MVP(@adshmh): implement a FullNode interface which caches the results.
// This needs to consider the GatewayMode:
//	A. Centralized: the list of owned apps is specified in advance, and onchain data can be cached before any requests are received.
//	B. Delegated: cache needs to be done in Lazy/incremental way, as user requests specifying different apps are received.
