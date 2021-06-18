module metamaskServer

go 1.16

require (
	github.com/ethereum/go-ethereum v1.10.2
	github.com/goinggo/mapstructure v0.0.0-20140717182941-194205d9b4a9
	github.com/spf13/viper v1.7.0
	google.golang.org/grpc v1.36.0
	kortho v0.0.0
)

replace kortho => ../kbft_dex/
