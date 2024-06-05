package main

import (
	"log"

	kzg_sdk "github.com/MultiAdaptive/kzg-sdk"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/vincentdebug/go-ord-tx/ord"
)

const dSrsSize = 1 << 16

func main() {
	netParams := &chaincfg.RegressionNetParams
	connCfg := &rpcclient.ConnConfig{
		Host:         "127.0.0.1:18443/wallet/hdd",
		User:         "user1",
		Pass:         "pwd123",
		HTTPPostMode: true,
		DisableTLS:   true,
	}

	client, err := rpcclient.New(connCfg, nil)
	if err != nil {
		log.Fatalf("Failed to create RPC client: %v", err)
	}
	defer client.Shutdown()

	commitTxOutPointList := make([]*wire.OutPoint, 0)
	// you can get from `client.ListUnspent()`
	utxoAddress := "bcrt1qkdv0tchxy33ctx4h33squ0ym39say6dwxwvuft"
	address, err := btcutil.DecodeAddress(utxoAddress, netParams)
	if err != nil {
		log.Fatalf("decode address err %v", err)
	}
	unspentList, err := client.ListUnspentMinMaxAddresses(1, 9999999, []btcutil.Address{address})

	if err != nil {
		log.Fatalf("list err err %v", err)
	}

	for i := range unspentList {
		inTxid, err := chainhash.NewHashFromStr(unspentList[i].TxID)
		if err != nil {
			log.Fatalf("decode in hash err %v", err)
		}
		commitTxOutPointList = append(commitTxOutPointList, wire.NewOutPoint(inTxid, unspentList[i].Vout))
	}
	kzgsdk, err := kzg_sdk.InitDomiconSdk(dSrsSize, "./srs")
	if err != nil {
		log.Fatalf("kzg sdk InitDomiconSdk failed")
	}
	cm, err := kzgsdk.GenerateDataCommit([]byte("multiadaptive data"))
	if err != nil {
		log.Fatalf("kzg sdk GenerateDataCommit failed")
	}
	dataCM := cm.Bytes()
	dataList := make([]ord.InscriptionData, 0)
	dataList = append(dataList, ord.InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        dataCM[:],
		Destination: "tb1p3m6qfu0mzkxsmaue0hwekrxm2nxfjjrmv4dvy94gxs8c3s7zns6qcgf8ef",
	})

	request := ord.InscriptionRequest{
		CommitTxOutPointList: commitTxOutPointList,
		CommitFeeRate:        25,
		FeeRate:              26,
		DataList:             dataList,
		SingleRevealTxOnly:   false,
	}

	tool, err := ord.NewInscriptionTool(netParams, client, &request)
	if err != nil {
		log.Fatalf("Failed to create inscription tool: %v", err)
	}
	err = tool.BackupRecoveryKeyToRpcNode()
	if err != nil {
		log.Fatalf("Failed to backup recovery key: %v", err)
	}
	commitTxHash, revealTxHashList, inscriptions, fees, err := tool.Inscribe()
	if err != nil {
		log.Fatalf("send tx errr, %v", err)
	}
	log.Println("commitTxHash, " + commitTxHash.String())
	for i := range revealTxHashList {
		log.Println("revealTxHash, " + revealTxHashList[i].String())
	}
	for i := range inscriptions {
		log.Println("inscription, " + inscriptions[i])
	}
	log.Println("fees: ", fees)
	// signet server
	// http://signet.ordinals.com/
	// https://signet.ordapi.xyz/
	// https://signet.earlyordies.com/
	// https://explorer-signet.openordex.org/

	//commitTxHash, 1ead71ecc9b17d449ec86ef217c5fd9476b8c7c27220834f9eadac32068d6194
	//revealTxHash, aafc791814d098cd5c7a9750812c532cbd415f0426b92c9482497cd53041ab59
	//revealTxHash, 1fd91dd783c041872583826840947d4c28f10770a059e69da9cf7c48ce77a016
	//revealTxHash, 7a003b0534c8a840d61ab7555daa8701da9f619a80aba8858d4f101cd7d62ee4
	// http://signet.ordinals.com/inscription/aafc791814d098cd5c7a9750812c532cbd415f0426b92c9482497cd53041ab59i0
	// http://signet.ordinals.com/inscription/1fd91dd783c041872583826840947d4c28f10770a059e69da9cf7c48ce77a016i0
	// http://signet.ordinals.com/inscription/7a003b0534c8a840d61ab7555daa8701da9f619a80aba8858d4f101cd7d62ee4i0

	//commitTxHash, b752d80e97196582fd02303f76b4b886c222070323fb7ccd425f6c89f5445f6c
	//revealTxHash, dceab59e310b94612dd2b746c188e1a4f5bb0f3d77c6b10d220c37951631f36a
	// http://signet.ordinals.com/inscription/dceab59e310b94612dd2b746c188e1a4f5bb0f3d77c6b10d220c37951631f36ai0
	// Currently, the official Ordinal parser can only parse a single NFT per transaction.
	// When the official Ordinal parser supports parsing multiple NFTs in the future, https://github.com/casey/ord/blob/0.5.1/src/inscription.rs#L32
	// we can consider using a single reveal transaction.
	// http://signet.ordinals.com/inscription/dceab59e310b94612dd2b746c188e1a4f5bb0f3d77c6b10d220c37951631f36ai1
	// http://signet.ordinals.com/inscription/dceab59e310b94612dd2b746c188e1a4f5bb0f3d77c6b10d220c37951631f36ai2
}
