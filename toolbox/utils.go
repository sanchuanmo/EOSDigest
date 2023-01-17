package toolbox

import (
	"fmt"
	"time"

	poly_go_sdk "github.com/polynetwork/poly-go-sdk"
	"github.com/polynetwork/poly/common"
)

func WaitPolyTx(txHash common.Uint256, polySdk *poly_go_sdk.PolySdk) {
	fmt.Printf("waiting poly transaction %s confirmed...\n", txHash.ToHexString())
	tick := time.NewTicker(100 * time.Millisecond)
	var h uint32
	startTime := time.Now()
	for range tick.C {
		h, _ = polySdk.GetBlockHeightByTxHash(txHash.ToHexString())
		curr, _ := polySdk.GetCurrentBlockHeight()
		if h > 0 && curr > h {
			break
		}

		if startTime.Add(100 * time.Millisecond); startTime.Second() > 300 {
			panic(fmt.Errorf("tx( %s ) is not confirm for a long time ( over %d sec )",
				txHash.ToHexString(), 300))
		}
	}
}
