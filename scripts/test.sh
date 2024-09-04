#!/bin/bash

forge cache clean
forge clean
rm -rf out
rm -rf ~/.foundry/anvil

# Start a new screen session and run 'anvil'
screen -dmS anvil-session anvil --block-time 1

# Wait for a few seconds to ensure 'anvil' is up and running
sleep 5

# Run the forge create command
forge create contracts/JackalV1.sol:JackalBridge --rpc-url http://127.0.0.1:8545 --private-key "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

# send myself some tokens >:)
cast send 0x9443A8C2aa7788EEE05f9734Ad4174a6C5CA0A25 --value 10ether  --private-key "0xac0974bec39a17e36ba4a6b4d238ff944bacb478cbed5efcae784d7bf4f2ff80"

# Run the Go program
go run ./...

