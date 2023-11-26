[![Go Report Card](https://goreportcard.com/badge/github.com/aergoio/aergo)](https://goreportcard.com/report/github.com/aergoio/aergo)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Maintainability](https://api.codeclimate.com/v1/badges/8ae0a363155bd9e8bccb/maintainability)](https://codeclimate.com/github/aergoio/aergo/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/8ae0a363155bd9e8bccb/test_coverage)](https://codeclimate.com/github/aergoio/aergo/test_coverage)
[![API Reference](https://godoc.org/github.com/aergoio/aergo?status.svg)](https://godoc.org/github.com/aergoio/aergo)

# Aergo - Distributed Trust at Scale

Official Chain Software of Aergo Protocol

We are developing the most practical and powerful platform for blockchain businesses.
There are four main ideologies guiding this project:

1. Developer-friendly
2. Guaranteed performance
3. Scalable architecture
4. Connect with the world

You can see the current public network status on [Aergoscan](https://aergoscan.io).

**[Official Documentation](http://docs.aergo.io)**

## Roadmaps

### 5th: Aergo Future
* Will be updated

### 4th: Aergo Mainnet 2.0 (planned in 4Q, 2019)
* DAO
* Staking/Voting
* Reward for Staking
* Alpha net (SQL enabled) launched
* Advanced performance features
* RAFT for Private Chain
* Authorization for Private Chain
* 2Way Swap with Ethereum
* 2Way Swap with Binance Dex

### 3rd: Aergo Mainnet (planned in March, 2019)
* BFT-dPOS with 13 BPs
* Aergo Scan
* Aergo Connect

### 2nd: Aergo Testnet (28, Dec, 2018)
* BFT-dPOS with Voting
  * The pre-test net has the function of agreeing blocks among the set BPs. TestNet has a function to select BP through voting.
* Named Account
  * For user's convenience, Named Account function that can be accessed based on Name rather than Address is provided.
* Expanded Aergo Lua
  * The Aergo Lua feature has been extended for more convenient development.
* Advanced Client Framework
  * Provides a wallet interface that interacts with keystore and manages nonce.
  * Provides the ability to make smart contracts through interface calls.
  * Provides a contract library to issue tokens based on Aergo.
* Hub Enterprise
  * Enterprise customers view management and monitoring of their networks as a prerequisite.
  * We provide Hub Enterprise control solution to solve this problem.
* Merkle Bridge Verification
  * StateTrie Merkle proof verifications and delegated token transfers are now implemented in the merkle-bridge.
* Various Smart Contract Examples
  * Provide some standard smart contracts
* TestNet
  * Launch the Test Network to provide network for community and experment
  * We provide https://aergoscan.io.

### 1st: Aergo Alpha (31, Oct, 2018)
* Consensus - BFT-dPOS (election not integrated)
  * We provide BFT by solving various problems that may occur in dpos.
* Aergo SQL smart contract (Lua-jit)
  * It is a powerful smart contract language providing DB function.
* Client - Ship
  * Client framework and development environment
  * Provides a package management and testing environment similar to NPM.
* Client SDK
  * heraj (java)
  * herajs (javascript)
  * herapy (python)
* Browser Wallet (1~2 weeks later)
  * Chrome Extension provides a coin transfer wallet.
* Sub Project
  * Litetree
    * Improved SQLite is used to provide DB functionality in a block chain.
    * Provides higher performance through LMDB.
  * Sparse Merkle Tree
    * Provides fast, space-saving sparse merkle tree.
  * Pre-Testnet
    * Launch the pre-testnet to monitor operation environment.
    * We provide https://aergoscan.io.

### beginning: Skeleton (31, July, 2018)
* Platform framework
* Stub consensus(dpos without voting)
* Account model
* Mempool
* Networking - p2p/protocol
* Cmd aergocli/aergosvr
* Simple client API
* Smart contract will not be released - you can see the prototype in [coinstack3sp2](https://github.com/coinstack/coinstackd)

## Key thoughts of the architecture

MVP based, Forward compatibility, Iteration

## Information

### Server port usages

| Usage | Port |
|-------|------|
|  gRPC | 7845 |
|  P2P  | 7846 |

## Installation

### Prerequisites

* Go1.20 - https://golang.org/dl
* Proto Buffers - https://github.com/google/protobuf
* CMake 3.0.0 or higher - https://cmake.org

### Build

[Building from Source](https://docs.aergo.io/en/latest/contribution/building-from-source.html)

## Contribution

Following Golang conventions
TBD

## License

All code is licensed under the MIT License (https://opensource.org/licenses/MIT).
