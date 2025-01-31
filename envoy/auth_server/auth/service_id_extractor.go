package auth

import (
	"fmt"
	"strings"
)

const reqHeaderServiceID = "target-service-id"

type ServiceIDExtractor struct {
	ServiceAliases map[string]string
}

func (e *ServiceIDExtractor) extractServiceID(headers map[string]string, host string) (string, error) {
	var serviceID string

	switch {
	// First, check for the target-service-id header
	case headers[reqHeaderServiceID] != "":
		serviceID = headers[reqHeaderServiceID]

	// Otherwise, check for the subdomain in the host
	case strings.Contains(host, "."):
		parts := strings.Split(host, ".")
		if len(parts) >= 3 {
			serviceID = parts[0]
		}
	}

	// Then, check for the service ID in the service aliases
	// and substitute the alias with the resolved ID if it is an alias
	if resolvedID, isAlias := e.ServiceAliases[serviceID]; isAlias {
		serviceID = resolvedID
	}

	// Return the service ID if it was found
	if serviceID != "" {
		return serviceID, nil
	}

	// Otherwise, return an error
	return "", fmt.Errorf("service ID not provided")
}
