module github.com/aergoio/aergo

go 1.13

require (
	github.com/Shopify/sarama v1.22.1 // indirect
	github.com/Workiva/go-datastructures v1.0.50 // indirect
	github.com/aergoio/aergo-actor v0.0.0-20190219030625-562037d5fec7
	github.com/aergoio/aergo-lib v1.1.1-rc13
	github.com/aergoio/etcd v0.0.0-20190429013412-e8b3f96f6399
	github.com/anaskhan96/base58check v0.0.0-20181220122047-b05365d494c4
	github.com/apache/thrift v0.12.0 // indirect
	github.com/bluele/gcache v0.0.0-20190518031135-bc40bd653833
	github.com/btcsuite/btcd v0.0.0-20190824003749-130ea5bddde3
	github.com/c-bata/go-prompt v0.2.3
	github.com/coreos/go-semver v0.3.0
	github.com/davecgh/go-spew v1.1.1
	github.com/derekparker/trie v0.0.0-20190322172448-1ce4922c7ad9
	github.com/emirpasic/gods v1.12.0
	github.com/fsnotify/fsnotify v1.4.8-0.20180830220226-ccc981bf8038
	github.com/funkygao/assert v0.0.0-20160929004900-4a267e33bc79 // indirect
	github.com/funkygao/golib v0.0.0-20180314131852-90d4905c1961
	github.com/gin-gonic/gin v1.7.4
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/golang/mock v1.3.1
	github.com/golang/protobuf v1.3.3
	github.com/grpc-ecosystem/grpc-opentracing v0.0.0-20180507213350-8e809c8a8645
	github.com/hashicorp/golang-lru v0.5.1
	github.com/improbable-eng/grpc-web v0.9.6
	github.com/json-iterator/go v1.1.9
	github.com/libp2p/go-addr-util v0.0.1
	github.com/libp2p/go-libp2p v0.4.0
	github.com/libp2p/go-libp2p-core v0.2.3
	github.com/libp2p/go-libp2p-peerstore v0.1.3
	github.com/magiconair/properties v1.8.1
	github.com/mattn/go-colorable v0.1.4
	github.com/mattn/go-runewidth v0.0.4 // indirect
	github.com/mattn/go-tty v0.0.0-20190424173100-523744f04859 // indirect
	github.com/minio/highwayhash v1.0.0 // indirect
	github.com/minio/sha256-simd v0.1.1
	github.com/mr-tron/base58 v1.1.2
	github.com/multiformats/go-multiaddr v0.1.1
	github.com/multiformats/go-multiaddr-dns v0.2.0 // indirect
	github.com/multiformats/go-multiaddr-net v0.1.0
	github.com/opentracing-contrib/go-observer v0.0.0-20170622124052-a52f23424492 // indirect
	github.com/opentracing/opentracing-go v1.0.2
	github.com/openzipkin-contrib/zipkin-go-opentracing v0.3.5
	github.com/orcaman/concurrent-map v0.0.0-20190314100340-2693aad1ed75 // indirect
	github.com/pkg/errors v0.9.1
	github.com/pkg/term v0.0.0-20190109203006-aa71e9d9e942 // indirect
	github.com/prometheus/client_golang v1.0.0 // indirect
	github.com/rs/cors v1.6.0 // indirect
	github.com/rs/zerolog v1.22.0
	github.com/sanity-io/litter v1.2.0
	github.com/serialx/hashring v0.0.0-20190515033939-7706f26af194 // indirect
	github.com/soheilhy/cmux v0.1.4
	github.com/spf13/cobra v0.0.5
	github.com/spf13/viper v1.5.0
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/willf/bitset v1.1.10 // indirect
	github.com/willf/bloom v2.0.3+incompatible
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20201021035429-f5854403a974
	google.golang.org/grpc v1.21.1
)

replace github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.8.1

replace github.com/dgraph-io/badger/v3 => github.com/shepelt/badger/v3 v3.2104.3
