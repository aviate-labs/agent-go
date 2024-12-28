package ledger

import (
	"fmt"

	"github.com/aviate-labs/agent-go"
	v1 "github.com/aviate-labs/agent-go/clients/ledger/proto/v1"
	"github.com/aviate-labs/agent-go/principal"
)

var LEDGER_PRINCIPAL = principal.MustDecode("ryjl3-tyaaa-aaaaa-aaaba-cai")

const MaxBlocksPerRequest = 2000

type BlockIndex uint64

type DataProvider struct {
	a *agent.Agent
}

func NewDataProvider() (*DataProvider, error) {
	a, err := agent.New(agent.DefaultConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}
	return &DataProvider{a: a}, nil
}

func (d DataProvider) GetArchiveIndex() ([]*v1.ArchiveIndexEntry, error) {
	var resp v1.ArchiveIndexResponse
	if err := d.a.QueryProto(
		LEDGER_PRINCIPAL,
		"get_archive_index_pb",
		nil,
		&resp,
	); err != nil {
		return nil, fmt.Errorf("failed to get archive index: %w", err)
	}
	return resp.Entries, nil
}

func (d DataProvider) GetRawBlock(height BlockIndex) (*v1.EncodedBlock, error) {
	var resp v1.BlockResponse
	if err := d.a.QueryProto(
		LEDGER_PRINCIPAL,
		"block_pb",
		&v1.BlockRequest{
			BlockHeight: uint64(height),
		},
		&resp,
	); err != nil {
		return nil, fmt.Errorf("failed to get block: %w", err)
	}
	switch blockResponse := resp.BlockContent.(type) {
	case *v1.BlockResponse_Block:
		return blockResponse.Block, nil
	case *v1.BlockResponse_CanisterId:
		archiveCanisterID := principal.Principal{Raw: blockResponse.CanisterId.SerializedId}
		var archiveResp v1.BlockResponse
		if err := d.a.QueryProto(
			archiveCanisterID,
			"get_block_pb",
			&v1.BlockRequest{
				BlockHeight: uint64(height),
			},
			&archiveResp,
		); err != nil {
			return nil, fmt.Errorf("failed to get blocks: %w", err)
		}
		// Will never return a CanisterId block.
		return archiveResp.GetBlock(), nil
	default:
		return nil, fmt.Errorf("unexpected block content type: %T", blockResponse)
	}
}

func (d DataProvider) GetRawBlocks(start, end BlockIndex) ([]*v1.EncodedBlock, error) {
	if end-start < 2000 {
		blocks, err := d.GetRawBlocksRange(LEDGER_PRINCIPAL, start, end)
		if err == nil {
			return blocks, nil
		}
	}
	archives, err := d.GetArchiveIndex()
	if err != nil {
		return nil, fmt.Errorf("failed to get archive index: %w", err)
	}
	var blocks []*v1.EncodedBlock
	for _, archive := range archives {
		if archive.HeightTo < uint64(start) || uint64(end) < archive.HeightFrom {
			continue
		}
		for start < min(BlockIndex(archive.HeightTo), end) {
			archiveEnd := min(end, BlockIndex(archive.HeightTo), start+MaxBlocksPerRequest)
			archiveBlocks, err := d.GetRawBlocksRange(principal.Principal{Raw: archive.CanisterId.SerializedId}, start, archiveEnd)
			if err != nil {
				return nil, fmt.Errorf("failed to get archive blocks: %w", err)
			}
			blocks = append(blocks, archiveBlocks...)
			start += BlockIndex(len(archiveBlocks))
		}
	}
	return blocks, nil
}

func (d DataProvider) GetRawBlocksRange(canisterID principal.Principal, start, end BlockIndex) ([]*v1.EncodedBlock, error) {
	var resp v1.GetBlocksResponse
	if err := d.a.QueryProto(
		canisterID,
		"get_blocks_pb",
		&v1.GetBlocksRequest{
			Start:  uint64(start),
			Length: uint64(end - start),
		},
		&resp,
	); err != nil {
		return nil, fmt.Errorf("failed to get blocks: %w", err)
	}
	switch blocksResponse := resp.GetBlocksContent.(type) {
	case *v1.GetBlocksResponse_Blocks:
		return blocksResponse.Blocks.Blocks, nil
	case *v1.GetBlocksResponse_Error:
		return nil, fmt.Errorf("failed to get blocks: %s", blocksResponse.Error)
	default:
		return nil, fmt.Errorf("unexpected get block content type: %T", blocksResponse)
	}
}

func (d DataProvider) GetTipOfChain() (*BlockIndex, error) {
	var resp v1.TipOfChainResponse
	if err := d.a.QueryProto(
		LEDGER_PRINCIPAL,
		"tip_of_chain_pb",
		&v1.TipOfChainRequest{},
		&resp,
	); err != nil {
		return nil, fmt.Errorf("failed to get tip of chain: %w", err)
	}
	height := BlockIndex(resp.ChainLength.Height)
	return &height, nil
}
