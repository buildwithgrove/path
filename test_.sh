#!/usr/bin/env bash
set -e
set -o nounset

# Step 1: Run make copy_morse_e2e_config
make copy_morse_e2e_config

# Step 2: Set environment variables
export MORSE_GATEWAY_SIGNING_KEY="c0f234a308e79741748aeb1d633f050f696d34a1b739b85412be2c8123848b25e7e91202573bdd1927b00fce9b0b46fa7944b06e6fe1bc987abc86e4d0dd47d6"
export MORSE_FULLNODE_URL="https://mainnet.rpc.grove.city/v1/61fc4532151e63003b23d628"
export MORSE_AATS=$(cat <<EOF
"1c2daba9e55f354875cfacd23f93f62cc19bf015":
  client_public_key: "e7e91202573bdd1927b00fce9b0b46fa7944b06e6fe1bc987abc86e4d0dd47d6"
  application_public_key: "c31aaefa0bb1732a8799e6fd70bbaf11897dfedacb11a7776607a5662ae950d4"
  application_signature: "a4fa7d70366aa58b8fe1a3697ff718dec2c7be11461013c8ce15f27e08428e863f8c86f4d329eecbb2d54def36a4834cccea9781609b7487bc06baab0c332c08"
EOF
)

# Step 3: Run the update script
./e2e/scripts/update_morse_config_from_secrets.sh

# Step 4: Run make test_e2e_morse_relay
make test_e2e_morse_relay