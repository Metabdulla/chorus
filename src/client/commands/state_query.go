package commands

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"gopkg.in/urfave/cli.v1"

	"github.com/Baptist-Publication/angine/types"
	"github.com/Baptist-Publication/chorus-module/lib/eth/common"
	ethtypes "github.com/Baptist-Publication/chorus-module/lib/eth/core/types"
	"github.com/Baptist-Publication/chorus-module/lib/eth/rlp"
	ac "github.com/Baptist-Publication/chorus-module/lib/go-common"
	cl "github.com/Baptist-Publication/chorus-module/lib/go-rpc/client"
	civil "github.com/Baptist-Publication/chorus/src/chain/node"
	"github.com/Baptist-Publication/chorus/src/client/commons"
	exsdk "github.com/Baptist-Publication/chorus/src/example/sdk"
	extypes "github.com/Baptist-Publication/chorus/src/example/types"
)

var (
	QueryCommands = cli.Command{
		Name:     "query",
		Usage:    "operations for query state",
		Category: "Query",
		Subcommands: []cli.Command{
			{
				Name:   "nonce",
				Usage:  "query account's nonce",
				Action: queryNonce,
				Flags: []cli.Flag{
					anntoolFlags.addr,
				},
			},
			{
				Name:   "balance",
				Usage:  "query account's balance",
				Action: queryBalance,
				Flags: []cli.Flag{
					anntoolFlags.addr,
				},
			},
			{
				Name:   "power",
				Usage:  "query account's vote power",
				Action: queryPower,
				Flags: []cli.Flag{
					anntoolFlags.addr,
				},
			},
			{
				Name:   "receipt",
				Usage:  "",
				Action: queryReceipt,
				Flags: []cli.Flag{
					anntoolFlags.hash,
				},
			},
			{
				Name:   "events",
				Usage:  "query events on the node",
				Action: queryEvents,
				Flags:  []cli.Flag{},
			},
			{
				Name:   "event_code",
				Usage:  "",
				Action: queryEventCode,
				Flags: []cli.Flag{
					anntoolFlags.codeHash,
				},
			},
			{
				Name:   "rmtreceipt",
				Usage:  "",
				Action: queryRemoteReceipt,
				Flags: []cli.Flag{
					anntoolFlags.hash,
				},
			},
			{
				Name:   "apps",
				Usage:  "query apps on the node",
				Action: queryNodeApps,
				Flags:  []cli.Flag{},
			},
		},
	}
)

func queryNonce(ctx *cli.Context) error {
	var chainID string
	if !ctx.GlobalIsSet("target") {
		chainID = "chorus"
	} else {
		chainID = ctx.GlobalString("target")
	}
	nonce, err := getNonce(chainID, ctx.String("address"))
	if err != nil {
		return err
	}

	fmt.Println("query result:", nonce)

	return nil
}

func getNonce(chainID, address string) (nonce uint64, err error) {
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	addrHex := ac.SanitizeHex(address)
	addr := common.Hex2Bytes(addrHex)
	query := append([]byte{civil.QueryNonce}, addr...)

	_, err = clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return 0, cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)
	nonce = binary.LittleEndian.Uint64(res.Result.Data)
	return nonce, nil
}

func queryBalance(ctx *cli.Context) error {
	var chainID string
	if !ctx.GlobalIsSet("target") {
		chainID = "chorus"
	} else {
		chainID = ctx.GlobalString("target")
	}
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	addrHex := ac.SanitizeHex(ctx.String("address"))
	addr := common.Hex2Bytes(addrHex)
	query := append([]byte{civil.QueryBalance}, addr...)

	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	balance := string(res.Result.Data)

	fmt.Println("query result:", balance)

	return nil
}

func queryPower(ctx *cli.Context) error {
	var chainID string
	if !ctx.GlobalIsSet("target") {
		chainID = "chorus"
	} else {
		chainID = ctx.GlobalString("target")
	}
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	addrHex := ac.SanitizeHex(ctx.String("address"))
	addr := common.Hex2Bytes(addrHex)
	query := append([]byte{civil.QueryPower}, addr...)

	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	var info []string
	if err := json.Unmarshal(res.Result.Data, &info); err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println("power:", info[0], "mheight:", info[1])

	return nil
}

func queryReceipt(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	hashHex := ac.SanitizeHex(ctx.String("hash"))
	hash := common.Hex2Bytes(hashHex)
	query := append([]byte{types.QueryTxExecution}, hash...)
	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	resultMap := map[string]interface{}{
		"code": res.Result.Code,
		"data": string(res.Result.Data),
		"log":  res.Result.Log,
	}

	fmt.Println("query result:", resultMap)

	return nil
}

func queryRemoteReceipt(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	hashHex := ac.SanitizeHex(ctx.String("hash"))
	hash := common.Hex2Bytes(hashHex)
	query := append([]byte{3}, hash...)
	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	receipt := extypes.ExternalReceipt{}
	err = exsdk.ToMsg(res.Result.Data, &receipt)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	fmt.Println("query result:", receipt)

	return nil
}

func queryTx(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")

	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	hashHex := ac.SanitizeHex(ctx.String("hash"))
	fmt.Println(hashHex)
	hash := common.Hex2Bytes(hashHex)
	query := append([]byte{3}, hash...)
	fmt.Println(len(query))
	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	receiptForStorage := new(ethtypes.ReceiptForStorage)

	err = rlp.DecodeBytes(res.Result.Data, receiptForStorage)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	fmt.Println("query result:", receiptForStorage.ContractAddress.Hex())

	return nil
}

func queryEvents(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	chainID := ctx.GlobalString("target")

	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	query := []byte{civil.QueryEvents}
	_, err := clientJSON.Call("query", []interface{}{chainID, query}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	if res.Result.IsOK() {
		buffers := bytes.NewBuffer(res.Result.Data)
		dec := gob.NewDecoder(buffers)
		keys := make([]string, 0)
		if err := dec.Decode(&keys); err != nil {
			return cli.NewExitError("fail to decode result: "+err.Error(), 127)
		}

		fmt.Println(keys)
		return nil
	}

	return cli.NewExitError(res.Result.Log, 127)
}

func queryEventCode(ctx *cli.Context) error {
	if !ctx.GlobalIsSet("target") {
		return cli.NewExitError("target chainid is missing", 127)
	}
	if !ctx.IsSet("code_hash") {
		return cli.NewExitError("query code_hash is missing", 127)
	}
	chainID := ctx.GlobalString("target")

	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)

	code_hash, err := hex.DecodeString(ctx.String("code_hash"))
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	_, err = clientJSON.Call("event_code", []interface{}{chainID, code_hash}, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}

	res := (*tmResult).(*types.ResultQuery)

	if res.Result.IsOK() {
		fmt.Println(string(res.Result.Data))
		return nil
	}

	return cli.NewExitError(res.Result.Log, 127)
}

func queryNodeApps(ctx *cli.Context) error {
	clientJSON := cl.NewClientJSONRPC(logger, commons.QueryServer)
	tmResult := new(types.RPCResult)
	_, err := clientJSON.Call("organizations", nil, tmResult)
	if err != nil {
		return cli.NewExitError(err.Error(), 127)
	}
	res := (*tmResult).(*types.ResultOrgs)
	fmt.Println(*res)
	return nil
}
