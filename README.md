## Chain Registry validation tool

```bash
crv dymension-chain-registry validate '/tmp/chain-registry' [--mainnet] [--testnet] [--devnet] [--internal-devnet]
# crv dym v '/tmp/chain-registry'
```

Flags:
- `mainnet`: Validate mainnet chains
- `testnet`: Validate testnet chains
- `devnet`: Validate devnet chains
- `internal-devnet`: Validate internal devnet chains
- None of above provided: Validate all chains
- `addition-chain-types-allowed`: Allow additional chain types defined bypass validation. By default, only following are allowed: "RollApp", "Regular", "EVM", "Hub", "Solana"