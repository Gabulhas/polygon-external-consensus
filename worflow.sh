#!/bin/bash

rm -rf temp-stuff/*
make build
./polygon-edge secrets init --data-dir ./temp-stuff/test-chain-1

./polygon-edge genesis --consensus external --ibft-validators-prefix-path test-chain- \
--bootnode /ip4/127.0.0.1/tcp/10001/p2p/16Uiu2HAmN8H1tSofB3FnxdhAd1JRpeLPDT7FvFSbHV8PvrwyYCe5 \
--bootnode /ip4/127.0.0.1/tcp/20001/p2p/16Uiu2HAm8VcctbC3G4BDa5UGEWknVJnyWyP6HxcR58cUmAY6xyom \
--dir temp-stuff/genesis_external.json


./polygon-edge server --data-dir ./test-chain-1 --chain temp-stuff/genesis_external.json --grpc-address :10000 --libp2p :10001 --jsonrpc :10002 --seal &
