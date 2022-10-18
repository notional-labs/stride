#!/bin/bash

KEY="test"
CHAINID="stride-testnet-1"
KEYRING="test"
MONIKER="localtestnet"
KEYALGO="secp256k1"
LOGLEVEL="info"

rm -rf ~/.stride*
# retrieve all args
WILL_START_FRESH=0
WILL_RECOVER=0
WILL_INSTALL=0
WILL_CONTINUE=0
# $# is to check number of arguments
if [ $# -gt 0 ];
then
    # $@ is for getting list of arguments
    for arg in "$@"; do
        case $arg in
        --fresh)
            WILL_START_FRESH=1
            shift
            ;;
        --recover)
            WILL_RECOVER=1
            shift
            ;;
        --install)
            WILL_INSTALL=1
            shift
            ;;
        --continue)
            WILL_CONTINUE=1
            shift
            ;;
        *)
            printf >&2 "wrong argument somewhere"; exit 1;
            ;;
        esac
    done
fi

# continue running if everything is configured
if [ $WILL_CONTINUE -eq 1 ];
then
    # Start the node (remove the --pruning=nothing flag if historical queries are not needed)
    strided start --pruning=nothing --log_level $LOGLEVEL --minimum-gas-prices=0.0001stake
    exit 1;
fi

# validate dependencies are installed
command -v jq > /dev/null 2>&1 || { echo >&2 "jq not installed. More info: https://stedolan.github.io/jq/download/"; exit 1; }

if [ $WILL_START_FRESH -eq 1 ];
then
    rm -rf $HOME/.strides*
fi

# install strided if not exist
if [ $WILL_INSTALL -eq 0 ];
then 
    command -v strided > /dev/null 2>&1 || { echo >&1 "installing strided"; make install; }
else
    echo >&1 "installing strided"
    rm -rf $HOME/.strides*
    go install ./...
fi

strided config keyring-backend $KEYRING
strided config chain-id $CHAINID

# determine if user wants to recorver or create new
rm debug/keys.txt

if [ $WILL_RECOVER -eq 0 ];
then
    KEY_INFO=$(strided keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO)
    echo $KEY_INFO >> debug/keys.txt
else
    KEY_INFO=$(strided keys add $KEY --keyring-backend $KEYRING --algo $KEYALGO --recover)
    echo $KEY_INFO >> debug/keys.txt
fi

echo >&1 "\n"

# init chain
strided init $MONIKER --chain-id $CHAINID

# Change parameter token denominations to stake
cat $HOME/.strides/config/genesis.json | jq '.app_state["staking"]["params"]["bond_denom"]="stake"' > $HOME/.strides/config/tmp_genesis.json && mv $HOME/.strides/config/tmp_genesis.json $HOME/.strides/config/genesis.json
cat $HOME/.strides/config/genesis.json | jq '.app_state["crisis"]["constant_fee"]["denom"]="stake"' > $HOME/.strides/config/tmp_genesis.json && mv $HOME/.strides/config/tmp_genesis.json $HOME/.strides/config/genesis.json
cat $HOME/.strides/config/genesis.json | jq '.app_state["gov"]["deposit_params"]["min_deposit"][0]["denom"]="stake"' > $HOME/.strides/config/tmp_genesis.json && mv $HOME/.strides/config/tmp_genesis.json $HOME/.strides/config/genesis.json
cat $HOME/.strides/config/genesis.json | jq '.app_state["mint"]["params"]["mint_denom"]="stake"' > $HOME/.strides/config/tmp_genesis.json && mv $HOME/.strides/config/tmp_genesis.json $HOME/.strides/config/genesis.json

# Set gas limit in genesis
# cat $HOME/.strides/config/genesis.json | jq '.consensus_params["block"]["max_gas"]="10000000"' > $HOME/.strides/config/tmp_genesis.json && mv $HOME/.strides/config/tmp_genesis.json $HOME/.strides/config/genesis.json

sed -i -E 's|swagger = false|swagger = true|g' $HOME/.strides/config/app.toml
sed -i -E 's|enable = false|enable = true|g' $HOME/.strides/config/app.toml

# Allocate genesis accounts (cosmos formatted addresses)
strided add-genesis-account $KEY 50000000000000000000000000stake --keyring-backend $KEYRING

# Sign genesis transaction
strided gentx $KEY 1000000stake --keyring-backend $KEYRING --chain-id $CHAINID

# Collect genesis tx
strided collect-gentxs

# Run this to ensure everything worked and that the genesis file is setup correctly
strided validate-genesis

# Start the node (remove the --pruning=nothing flag if historical queries are not needed)
strided start --pruning=nothing --log_level $LOGLEVEL --minimum-gas-prices=0.0001stake --rpc.laddr tcp://0.0.0.0:26657