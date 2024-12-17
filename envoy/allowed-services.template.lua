-- IMPORTANT: All Services for the PATH Service Gateway must be listed here for Envoy Proxy to forward requests to PATH.
--
-- If you wish to define aliases for existing services, you must define the alias as the key and the service ID as the value.
--
-- eg. the alias "eth = F00C" enables the URL "http://eth.path.grove.city" to be routed to the service with the ID "F00C".
--
-- For the service to utilize PATH's Quality of Service (QoS) features, the service ID value must match the values defined in
-- PATH's `qos` module. TODO_IMPROVE(@commoddity): Add link to the file & line in the QoS module.
return {
    -- Morse Service IDs
    F00C = "F00C", -- Ethereum Service (Authoritative ID)
    eth = "F00C",  -- Ethereum Service (Alias)
    
    -- Shannon Service IDs
    anvil = "anvil",  -- Anvil Service (Authoritative ID)
  }
