package tracker

import (
	"context"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	tellorCommon "github.com/tellor-io/TellorMiner/common"
	"github.com/tellor-io/TellorMiner/config"
	tellor "github.com/tellor-io/TellorMiner/contracts"
	"github.com/tellor-io/TellorMiner/db"
	"github.com/tellor-io/TellorMiner/rpc"
)

func TestFetchData(t *testing.T) {
	cfg, err := config.GetConfig()
	if err != nil {
		t.Fatal(err)
	}
	startBal := big.NewInt(456000)
	top50 := make([]*big.Int, 51)
	mockQueryParams := &rpc.MockQueryMeta{QueryString: "json(https://api.gdax.com/products/ETH-USD/ticker).price", Granularity: 1000}
	paramsMap := make(map[uint]*rpc.MockQueryMeta)
	for i := range top50 {
		top50[i] = big.NewInt(int64(i + 51))
		paramsMap[uint(i+51)] = mockQueryParams
	}

	opts := &rpc.MockOptions{ETHBalance: startBal, Nonce: 1, GasPrice: big.NewInt(700000000),
		TokenBalance: big.NewInt(0), Top50Requests: top50, QueryMetadata: paramsMap}
	client := rpc.NewMockClientWithValues(opts)

	contractAddress := common.HexToAddress(cfg.ContractAddress)
	masterInstance, err := tellor.NewTellorMaster(contractAddress, client)
	if err != nil {
		t.Fatal(err)
	}

	dbPath := filepath.Join(os.TempDir(), "test_fetchData")
	fmt.Println("Using DB at path", dbPath)
	DB, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	tracker1 := &Top50Tracker{}
	ctx1 := context.WithValue(context.Background(), tellorCommon.ClientContextKey, client)
	ctx1 = context.WithValue(ctx1, tellorCommon.DBContextKey, DB)
	ctx1 = context.WithValue(ctx1, tellorCommon.MasterContractContextKey, masterInstance)
	err = tracker1.Exec(ctx1)
	if err != nil {
		t.Fatal(err)
	}

	tracker := &RequestDataTracker{}
	ctx := context.WithValue(context.Background(), tellorCommon.ClientContextKey, client)
	ctx = context.WithValue(ctx, tellorCommon.DBContextKey, DB)
	err = tracker.Exec(ctx)
	if err != nil {
		t.Fatal(err)
	}
	v, err := DB.Get(fmt.Sprintf("%s%d", db.QueriedValuePrefix, 61))
	if err != nil {
		t.Fatal(err)
	}
	b, err := hexutil.DecodeBig(string(v))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("Request 61 result", b)
	t.Logf("Data stored: %v\n", string(v))
	if b.Cmp(big.NewInt(1)) != 1 {
		t.Fatalf("Expected data to be stored for request id 61 but found: " + b.String())
	}
}