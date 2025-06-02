package gateway

import (
	"errors"
)

// Gateway request error indicating no service ID was provided by the user.
var GatewayErrNoServiceIDProvided = errors.New("no service ID provided")
