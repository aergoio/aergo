[![Go Report Card](https://goreportcard.com/badge/github.com/aergoio/aergo)](https://goreportcard.com/report/github.com/aergoio/aergo)
[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Travis_ci](https://travis-ci.org/aergoio/aergo.svg?branch=master)](https://travis-ci.org/aergoio/aergo)
[![Maintainability](https://api.codeclimate.com/v1/badges/8ae0a363155bd9e8bccb/maintainability)](https://codeclimate.com/github/aergoio/aergo/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/8ae0a363155bd9e8bccb/test_coverage)](https://codeclimate.com/github/aergoio/aergo/test_coverage)
[![LoC](https://tokei.rs/b1/github/aergoio/aergo)](https://github.com/aergoio/aergo)

# Aergo - The Most Practical Blockchain Platform for Blockchain Businesses

Official Chain Software of Aergo Protocol

We are developing the most practical and powerful platform for blockchain businesses. This will be a huge challenge. There are 4 main ideologies regarding this project.

1. Developer-friendly
2. Guaranteed performance 
3. Scalable architecture
4. Connect with the world

## Roadmaps

### beginning: Skeleton (31, July, 2018)
* Platform framework
* Stub consensus(dpos without voting)
* Account model
* Mempool
* Networking - p2p/protocol
* Cmd aergocli/aergosvr
* Simple client API
* Smart contract will not be released - you can see the prototype in [coinstack3sp2](https://github.com/coinstack/coinstackd)

### 1st: Aergo Alpha (planned in Sep, 2018)
* Simple dPOS
* Aergo SQL smart contract (LLVM)
* Basic client framework and development environment (other repositories)

### 2nd: Aergo Testnet (planned in Dec, 2018)
* Advanced dPOS
* Governance with DAO
* Advanced client framework (including domain-specific parts)

### 3rd: Aergo Mainnet (planned in March, 2019)
* Parallelism (inter-contract)
* Simple branching (2WP or simple Plasma)

### 4th: Aergo World Launch (planned in 4Q, 2019)
* Orchestration with Aergo Horde
* Service with Aergo Hub
* Advanced performance features 

### 5th: Aergo Future
* Will be updated

## Key thoughts of the architecture

1. MVP based, Forward compatibility, Iteration
2. Following Golang conventions

## Information

### Server port usages

| Usage | Port |
|-------|------|
|  gRPC | 7845 |
|*  P2P | 7846 |
|  REST | 8080 |

## Installation

### Prerequisites

* Go1.10 or higher - https://golang.org/dl
* Glide - https://github.com/Masterminds/glide
* Proto Buffers - https://github.com/google/protobuf

### Build

```
$ go get -d github.com/aergoio/aergo/account
$ cd ${GOPATH}/src/github.com/aergoio/aergo
$ make deps
$ make
```

## Contribution

## License

All code is licensed under the MIT License (https://opensource.org/licenses/MIT).

