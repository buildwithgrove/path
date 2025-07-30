# TODO: Add EVM Chain ID Validation to Endpoint Selection

## Issue
The EVM chain ID check (`checkEVMChainID`) is being populated for endpoints but is not being validated during endpoint selection in `basicEndpointValidation()`.

## Current State
- The `checkEVMChainID` field is added to the `endpoint` struct
- The check is being run and observations are being collected
- The check results are NOT being used to validate endpoints

## Required Changes
1. Add `validateEndpointEVMChecks()` method similar to existing validation methods
2. Check if the service supports JSON_RPC (for EVM chains)
3. Validate that the EVM chain ID matches the expected value from config
4. Call this validation in `basicEndpointValidation()` when appropriate

## Implementation Notes
- Only validate EVM chain ID if `evmChainID` is configured (non-empty) in the service config
- Follow the same pattern as CometBFT and CosmosSDK validation methods
- Add appropriate error variables for EVM validation failures