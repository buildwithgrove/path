#!/bin/bash


clear
echo "🌿 Welcome to PATH. This will guide you through the steps to start the service."
echo "🐳 In order to proceed, Docker must be installed and running on your machine."
echo "❔ Would you like to proceed? (y/n)"
read proceed

if [[ "$proceed" != "y" ]]; then
  echo "❌ Setup aborted."
  exit 1
fi

CONFIG_YAML="./cmd/.config.yaml"

if [[ -f "$CONFIG_YAML" ]]; then
  echo "❌ Configuration file already exists. Setup aborted."
  exit 1
fi

if ! command -v docker &> /dev/null; then
  echo "❌ Docker is not installed. Please install Docker and try again."
  exit 1
fi

if ! docker info &> /dev/null; then
  echo "❌ Docker daemon is not running. Please start Docker and try again."
  exit 1
fi

clear
echo "🔧 In order to proceed with setup you will need a Shannon Full Node and the following values for actors staked on the Shannon protocol:"
echo "- A Gateway address"
echo "- A Gateway private key"
echo "- An address of an Application delegated to the Gateway"
echo ""
echo "📄 For instructions on how to set all of this up yourself, please see:"
echo "https://dev.poktroll.com/operate/quickstart/docker_compose_walkthrough"
echo ""
echo "❓ Do you have all of the above and would like to proceed? (y/n)"
read proceed

if [[ "$proceed" != "y" ]]; then
  echo "❌ Setup aborted."
  exit 1
fi

clear

validate_url() {
  if [[ ! "$1" =~ ^http(s)?://[a-zA-Z0-9.-]+ ]]; then
    echo "❌ Invalid URL. Must be a valid URL (e.g. https://example.com). Please try again."
    return 1
  fi
  return 0
}

validate_host_port() {
  if [[ ! "$1" =~ ^[a-zA-Z0-9.-]+:[0-9]+$ ]]; then
    echo "❌ Invalid host port. Must be in the format 'hostname:port' (e.g. localhost:9090). Please try again."
    return 1
  fi
  return 0
}

validate_address() {
  if [[ ! "$1" =~ ^pokt1[0-9a-zA-Z]{38}$ ]]; then
    echo "❌ Invalid address. Must be 43 characters long and start with 'pokt1'. Please try again."
    return 1
  fi
  return 0
}

validate_gateway_private_key() {
  if [[ ! "$1" =~ ^[0-9a-fA-F]{64}$ ]]; then
    echo "❌ Invalid gateway private key. Must be a 64-character hexadecimal string. Please try again."
    return 1
  fi
  return 0
}

while true; do
  echo "🔗 Please enter your Full Node URL (e.g. http://path-service:26657):"
  read rpc_url
  validate_url "$rpc_url" && break
done

clear
while true; do
  echo "🔗 Please enter your Full Node gRPC host & port (e.g. path-service:9090):"
  read host_port
  validate_host_port "$host_port" && break
done
echo "❓ Does your Full Node gRPC connection use TLS? (y/n)"
read use_tls

clear
while true; do
  echo "🔗 Please enter your Gateway address (43 characters starting with pokt1...):"
  read gateway_address
  validate_address "$gateway_address" && break
done

clear
while true; do
  echo "🔗 Please enter your Gateway private key (64 characters hexadecimal string):"
  echo "NOTE: It will not be displayed on screen as you type."
  read -s gateway_private_key
  validate_gateway_private_key "$gateway_private_key" && break
done

clear
while true; do
  echo "🔗 Please enter your delegated Application address (43 characters starting with pokt1...):"
  read delegated_app_address
  validate_address "$delegated_app_address" && break
done

clear
echo "📂 Running make copy_config..."
make copy_config

sed -i '' "s|^\([[:space:]]*\)rpc_url:.*|\1rpc_url: $rpc_url|g" "$CONFIG_YAML"
sed -i '' "s|^\([[:space:]]*\)host_port:.*|\1host_port: $host_port|g" "$CONFIG_YAML"
sed -i '' "s|^\([[:space:]]*\)gateway_address:.*|\1gateway_address: $gateway_address|g" "$CONFIG_YAML"
sed -i '' "s|^\([[:space:]]*\)gateway_private_key:.*|\1gateway_private_key: $gateway_private_key|g" "$CONFIG_YAML"
sed -i '' "/^\([[:space:]]*\)delegated_app_addresses:/,/^\([[:space:]]*\)-/d" "$CONFIG_YAML"
awk '/gateway_private_key:/ {print; print "    delegated_app_addresses:\n      - \"'$delegated_app_address'\""; next}1' "$CONFIG_YAML" > temp.yaml && mv temp.yaml "$CONFIG_YAML"

if [[ "$use_tls" == "y" ]]; then
  sed -i '' "/^\([[:space:]]*\)insecure:/d" "$CONFIG_YAML"
fi

echo "🌿 Starting PATH service... "
make path_up

timeout=20
interval=1
elapsed=0

while [[ $elapsed -lt $timeout ]]; do
  status_code=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/healthz)
  if [[ $status_code -eq 200 ]]; then
    clear
    echo "🌿 PATH Service is now running!"
    echo "You may now send service requests for service '0021' (eth-mainnet) using http://eth-mainnet.localhost:3000/v1"
    echo ""
    echo "💡 Example service request using cURL:"
    echo 'curl http://eth-mainnet.localhost:3000/v1 -d "{"jsonrpc": "2.0", "id": 1, "method": "eth_blockNumber" }"'
    echo ""
    echo "🌱 To enable additional services, edit the 'services' section of the .config.yaml file and restart the PATH service using 'make path_restart'."
    echo ""
    echo "💚 Happy relaying!"
    exit 0
  fi
  sleep $interval
  elapsed=$((elapsed + interval))
done

echo "❌ Service health check failed after $timeout seconds."
exit 1
