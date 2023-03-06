module github.com/aergoio/aergo

go 1.12

require (
	github.com/Shopify/sarama v1.22.1 // indirect
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/VictoriaMetrics/fastcache v1.12.1 // indirect
	github.com/Workiva/go-datastructures v1.0.50 // indirect
	github.com/aergoio/aergo-actor v0.0.0-20190219030625-562037d5fec7
	github.com/aergoio/aergo-lib v1.1.0
	github.com/aergoio/badger v1.6.0-gcfix // indirect
	github.com/aergoio/etcd v0.0.0-20190429013412-e8b3f96f6399
	github.com/anaskhan96/base58check v0.0.0-20181220122047-b05365d494c4
	github.com/apache/thrift v0.12.0 // indirect
	github.com/bluele/gcache v0.0.0-20190518031135-bc40bd653833
	github.com/btcsuite/btcd v0.0.0-20190824003749-130ea5bddde3
	github.com/btcsuite/btcd/btcec/v2 v2.3.2 // indirect
	github.com/c-bata/go-prompt v0.2.3
	github.com/cockroachdb/pebble v0.0.0-20230302152029-717cbce0c2e3 // indirect
	github.com/coreos/go-semver v0.3.0
	github.com/davecgh/go-spew v1.1.1
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/derekparker/trie v0.0.0-20190322172448-1ce4922c7ad9
	github.com/emirpasic/gods v1.12.0
	github.com/ethereum/go-ethereum v1.11.2 // indirect
	github.com/fsnotify/fsnotify v1.6.0
	github.com/funkygao/assert v0.0.0-20160929004900-4a267e33bc79 // indirect
	github.com/funkygao/golib v0.0.0-20180314131852-90d4905c1961
	github.com/gin-gonic/gin v1.8.1
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/golang/mock v1.5.0
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d
	github.com/holiman/uint256 v1.2.1 // indirect
	github.com/improbable-eng/grpc-web v0.9.6
	github.com/json-iterator/go v1.1.12
	github.com/klauspost/compress v1.16.0 // indirect
	github.com/libp2p/go-addr-util v0.0.1
	github.com/libp2p/go-libp2p v0.4.0
	github.com/libp2p/go-libp2p-core v0.2.3
	github.com/libp2p/go-libp2p-peerstore v0.1.3
	github.com/libp2p/go-nat v0.0.3 // indirect
	github.com/magiconair/properties v1.8.5
	github.com/mattn/go-colorable v0.1.13
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/mattn/go-tty v0.0.0-20190424173100-523744f04859 // indirect
	github.com/minio/sha256-simd v0.1.1
	github.com/mr-tron/base58 v1.1.2
	github.com/multiformats/go-multiaddr v0.1.1
	github.com/multiformats/go-multiaddr-dns v0.2.0 // indirect
	github.com/multiformats/go-multiaddr-net v0.1.0
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.1.0
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5
	github.com/orcaman/concurrent-map v0.0.0-20190314100340-2693aad1ed75 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/term v0.0.0-20190109203006-aa71e9d9e942 // indirect
	github.com/prometheus/common v0.41.0 // indirect
	github.com/prometheus/tsdb v0.10.0 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	github.com/rs/zerolog v1.16.1-0.20191111091419-e709c5d91e35
	github.com/sanity-io/litter v1.2.0
	github.com/serialx/hashring v0.0.0-20190515033939-7706f26af194 // indirect
	github.com/shirou/gopsutil v3.21.11+incompatible // indirect
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cobra v1.5.0
	github.com/spf13/viper v1.8.1
	github.com/stretchr/testify v1.8.0
	github.com/tklauser/go-sysconf v0.3.11 // indirect
	github.com/willf/bitset v1.1.10 // indirect
	github.com/willf/bloom v2.0.3+incompatible
	golang.org/x/crypto v0.6.0
	golang.org/x/exp v0.0.0-20230224173230-c95f2b4c22f2 // indirect
	golang.org/x/net v0.7.0
	google.golang.org/grpc v1.38.0
)

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.8.1
