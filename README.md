# Jackal EVM Bridge

## Init
To init your system, you can get a new address like this.
```shell
mulberry wallet address
```
This will also create the `~/.mulberry` directory and all the config files. You can adjust where this goes with the `home` flag.

## Config
You will get a new config file that looks something like this, feel free to change out the entries as you see fit.
```yaml
jackal_config:
    rpc: https://testnet-rpc.jackalprotocol.com:443
    grpc: jackal-testnet-grpc.polkachu.com:17590
    seed_file: seed.json
    contract: jkl1tjaqmxerltfqgxwp5p9sdwr9hncmdvevg9mllac2qjqfrjte6wkqwnq3h9
networks_config:
    - name: Ethereum Sepolia
      rpc: wss://ethereum-sepolia-rpc.publicnode.com
      contract: 0x730fdF2ee985Ac0F7792f90cb9e1E5485d340208
      chain_id: 11155111
      finality: 2
    - name: Base Sepolia
      rpc: wss://base-sepolia-rpc.publicnode.com
      contract: 0x20738B8eaB736f24c7881bA48263ee60Eb2a0A2a
      chain_id: 84532
      finality: 2
```

### New EVM Networks
To support a new network, create a new entry under `networks_config` like this:
```yaml
    - name: OP Sepolia
      rpc: wss://optimism-sepolia-rpc.publicnode.com
      contract: 0x5eb3B1f07b33da11D91290B57952b4b6f312e8dd
      chain_id: 11155420
      finality: 1
```
## Testing

Run `./scripts/test.sh` to start a test environment.