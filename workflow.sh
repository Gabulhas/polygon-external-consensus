#!/bin/bash

# Prepare stuff before
protoc --go_out=. --go-grpc_out=. ./consensus/external/**/*.proto

# Arguments
totalNodes=$1
bootNodes=$2

# Constants
genesisCommand="./polygon-edge genesis --consensus external --ibft-validators-prefix-path test-chain- --dir temp-stuff/genesis.json"

# Cleanup
killall polygon-edge
rm -rf temp-stuff/*
make build

# Create Nodes
toAddToList=$bootNodes
nodeStartCommands=()

for ((i=1; i<=$totalNodes; i++)); do
    nodeID=$(./polygon-edge secrets init --data-dir ./temp-stuff/test-chain-$i| sed -n "s/Node ID.*= \(.*\)/\1/p")

    if test $toAddToList -gt 0
    then
        genesisCommand="$genesisCommand --bootnode /ip4/127.0.0.1/tcp/"$i"001/p2p/$nodeID"
        toAddToList=$((toAddToList-1))
    fi
    nodeStartCommands+=("./polygon-edge server --data-dir ./temp-stuff/test-chain-$i --chain ./temp-stuff/genesis.json --grpc-address :"$i"000 --libp2p :"$i"001 --jsonrpc :"$i"002 --seal &")
done

eval $genesisCommand

for c in "${nodeStartCommands[@]}";
do
    eval $c
done
