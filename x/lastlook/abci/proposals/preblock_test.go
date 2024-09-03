package proposals_test

import (
	"crypto/rand"
	"testing"

	"github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/neutron-org/neutron/v4/testutil/lastlook/keeper"
	"github.com/neutron-org/neutron/v4/x/lastlook/abci/proposals"
	lastlookkeeper "github.com/neutron-org/neutron/v4/x/lastlook/keeper"
	lastlooktypes "github.com/neutron-org/neutron/v4/x/lastlook/types"
)

func randomTxsWithSize(n int, size int) [][]byte {
	txs := make([][]byte, 0, n)
	for i := 0; i < n; i++ {
		buf := make([]byte, size)
		if _, err := rand.Read(buf); err != nil {
			panic(err)
		}

		txs = append(txs, buf)
	}

	return txs
}

func TestPrepareProposalHandler(t *testing.T) {
	type testCase struct {
		name string
		// this method must return RequestPrepareProposal as the first argument, expected txs to be returned after proposal preparation and expected error
		malleate func(*testing.T, sdk.Context, *lastlookkeeper.Keeper) (types.RequestPrepareProposal, [][]byte, error)
	}

	testCases := []testCase{
		{
			name: "All good, 10 txs in mempool, 0 in queue, height 0",
			malleate: func(t *testing.T, ctx sdk.Context, k *lastlookkeeper.Keeper) (types.RequestPrepareProposal, [][]byte, error) {
				txReq := randomTxsWithSize(10, 1)
				expectedBatch := lastlooktypes.Batch{
					Proposer: []byte("ProposerAddress"),
					Txs:      txReq,
				}
				expectedBatchBz, err := expectedBatch.Marshal()
				require.NoError(t, err)

				return types.RequestPrepareProposal{
						MaxTxBytes:      100,
						Txs:             txReq,
						Height:          0,
						ProposerAddress: []byte("ProposerAddress"),
					},
					[][]byte{expectedBatchBz},
					nil
			},
		},
		{
			name: "All good, 10 txs in mempool, 5 in queue, height 2",
			malleate: func(t *testing.T, ctx sdk.Context, k *lastlookkeeper.Keeper) (types.RequestPrepareProposal, [][]byte, error) {
				txsInQueue := randomTxsWithSize(5, 2)
				require.NoError(t, k.StoreBatch(ctx, 2, sdk.AccAddress([]byte("ProposerAddress")), txsInQueue))

				txReq := randomTxsWithSize(10, 1)
				expectedBatch := lastlooktypes.Batch{
					Proposer: []byte("ProposerAddress"),
					Txs:      txReq,
				}
				expectedBatchBz, err := expectedBatch.Marshal()
				require.NoError(t, err)

				expectedTxs := make([][]byte, 0)
				expectedTxs = append(expectedTxs, expectedBatchBz)
				expectedTxs = append(expectedTxs, txsInQueue...)

				return types.RequestPrepareProposal{
						MaxTxBytes:      10000,
						Txs:             txReq,
						Height:          2,
						ProposerAddress: []byte("ProposerAddress"),
					},
					expectedTxs,
					nil
			},
		},
		{
			name: "All good, 100 txs in mempool each 1 byte (total 100 byte), 0 in queue, MaxTxSize limit 50, height 0",
			malleate: func(t *testing.T, ctx sdk.Context, k *lastlookkeeper.Keeper) (types.RequestPrepareProposal, [][]byte, error) {
				txReq := randomTxsWithSize(100, 1)
				expectedBatch := lastlooktypes.Batch{
					Proposer: []byte("ProposerAddress"),
					// we expected only 11 txs to be in a batch because there is:
					// * overhead for encoded Batch structure - 17 bytes;
					// * overhead for each tx item in a batch - 3 bytes;
					// * each tx consumes 1 byte of memory (just for simple calculations).
					// In total with MaxTxBytes equals to 50 bytes, batch with 11 txs consumes exactly 50 bytes:
					// 17 + 3*11 = 50
					Txs: txReq[:11],
				}
				expectedBatchBz, err := expectedBatch.Marshal()
				require.NoError(t, err)

				expectedTxs := make([][]byte, 0)
				expectedTxs = append(expectedTxs, expectedBatchBz)

				return types.RequestPrepareProposal{
						MaxTxBytes:      50,
						Txs:             txReq,
						Height:          0,
						ProposerAddress: []byte("ProposerAddress"),
					},
					expectedTxs,
					nil
			},
		},
	}

	for _, tc := range testCases {
		lastLookKeeper, ctx := keeper.LastLookKeeper(t)
		proposalHandler := proposals.NewProposalHandler(ctx.Logger(), lastLookKeeper)

		t.Run(tc.name, func(t *testing.T) {
			req, expectedTxs, expectedError := tc.malleate(t, ctx, lastLookKeeper)

			resp, err := proposalHandler.PrepareProposalHandler()(ctx, &req)
			require.Equal(t, expectedError, err)
			require.Equal(t, expectedTxs, resp.Txs)

			totalBytes := int64(0)
			for _, tx := range resp.Txs {
				totalBytes += int64(len(tx))
			}

			// total bytes consumption must not exceed provided req.MaxTxBytes
			require.GreaterOrEqual(t, req.MaxTxBytes, totalBytes)
		})
	}
}
