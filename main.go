package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/blocks"
	"github.com/prysmaticlabs/prysm/v3/consensus-types/interfaces"
	types "github.com/prysmaticlabs/prysm/v3/consensus-types/primitives"
	ethpb "github.com/prysmaticlabs/prysm/v3/proto/prysm/v1alpha1"
	"google.golang.org/grpc"
)

var (
	output     = flag.String("output", "output.csv", "output csv file path (default $PWD/output.csv)")
	endpoint   = flag.String("grpc-endpoint", "localhost:4000", "gRPC endpoint for a Prysm node")
	startEpoch = flag.Uint64("start-epoch", 0, "start epoch for the requests (default: 0)")
)

func main() {
	flag.Parse()
	csvFile, err := os.Create(*output)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := csvFile.Close(); err != nil {
			panic(err)
		}
	}()
	csvWriter := csv.NewWriter(csvFile)
	if err := csvWriter.Write([]string{"slot", "block_root", "graffiti"}); err != nil {
		panic(err)
	}
	conn, err := grpc.Dial(*endpoint, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	beaconClient := ethpb.NewBeaconChainClient(conn)
	ctx := context.Background()
	chainHead, err := beaconClient.GetChainHead(ctx, &empty.Empty{})
	if err != nil {
		panic(err)
	}
	fmt.Printf(
		"Retrieved chain head data, head epoch is %d and head slot is %d\n",
		chainHead.HeadEpoch,
		chainHead.HeadSlot,
	)
	pageSize := int32(50) // 50 blocks per page. More than enough as there are typically 32 blocks per epoch.
	for i := types.Epoch(*startEpoch); i < chainHead.HeadEpoch; i++ {
		pageToken := "0"
		totalBlocks := int32(0)
		records := make([][]string, 0)
		// Retrieve paginated lists of blocks until there are no more.
		for pageToken != "" {
			resp, err := beaconClient.ListBeaconBlocks(ctx, &ethpb.ListBlocksRequest{
				QueryFilter: &ethpb.ListBlocksRequest_Epoch{
					Epoch: i,
				},
				PageSize:  pageSize,
				PageToken: pageToken,
			})
			if err != nil {
				panic(err)
			}
			totalBlocks = resp.TotalSize
			pageToken = resp.NextPageToken

			// For every block in the response, extract the graffiti, slot, and block root data.
			for _, container := range resp.BlockContainers {
				wrappedBlock, err := extractBlockInterface(container)
				if err != nil {
					panic(err)
				}
				records = append(records, []string{
					fmt.Sprintf("%d", wrappedBlock.Block().Slot()),
					fmt.Sprintf("%#x", container.BlockRoot),
					fmt.Sprintf("%#x", wrappedBlock.Block().Body().Graffiti()),
				})
			}
		}
		fmt.Printf("Finished retrieving blocks for epoch %d. Extracted %d total block graffitis\n", i, totalBlocks)
		if err := csvWriter.WriteAll(records); err != nil {
			panic(err)
		}
		csvWriter.Flush()
	}
}

func extractBlockInterface(cntr *ethpb.BeaconBlockContainer) (interfaces.SignedBeaconBlock, error) {
	switch blk := cntr.Block.(type) {
	case *ethpb.BeaconBlockContainer_Phase0Block:
		return blocks.NewSignedBeaconBlock(blk.Phase0Block)
	case *ethpb.BeaconBlockContainer_AltairBlock:
		return blocks.NewSignedBeaconBlock(blk.AltairBlock)
	case *ethpb.BeaconBlockContainer_BellatrixBlock:
		return blocks.NewSignedBeaconBlock(blk.BellatrixBlock)
	case *ethpb.BeaconBlockContainer_BlindedBellatrixBlock:
		return blocks.NewSignedBeaconBlock(blk.BlindedBellatrixBlock)
	default:
		return nil, fmt.Errorf("unsupported block type %T", blk)
	}
}
