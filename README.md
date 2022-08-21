# Graffiti Fetch

This repository is a simple Go binary that retrieves the entire set of block graffitis in the [Ethereum beacon chain](https://beaconcha.in) since a specified epoch and writes the output to a CSV file.

# Installing

You will need to a connect to a [Prysm](https://github.com/prysmaticlabs/prysm) beacon chain node to use this tool. This tool uses [gRPC](https://grpc.io/) to retrieve data via Prysm's [public beacon API](https://github.com/prysmaticlabs/prysm/blob/develop/proto/prysm/v1alpha1/beacon_chain.proto). Prysm nodes expose a gRPC server on localhost:4000 by default. To install Prysm, see [here](https://docs.prylabs.network/docs/install/install-with-script).

Download [Go](https://go.dev/dl/), then:

```
git clone https://github.com/rauljordan/graffiti-fetcher && cd graffiti-fetcher
go build .
```

# Usage

```
./graffiti-fetcher --help
Usage of ./graffiti-fetcher:
  -grpc-endpoint string
    	gRPC endpoint for a Prysm node (default "localhost:4000")
  -output string
    	output csv file path (default $PWD/output.csv) (default "output.csv")
  -start-epoch uint
    	start epoch for the requests (default: 0)
```
