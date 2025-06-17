#!/bin/bash

# Parameterize the network to alpha|beta|main
# Make it possible to pass in home or default to ~/.pocket_prod
pocketd query service all-services --network=main --home=~/.pocket_prod --grpc-insecure=false -o json | jq '.service[].id'

