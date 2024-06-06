package main

import (
	"encoding/hex"
	"log"

	kzg_sdk "github.com/MultiAdaptive/kzg-sdk"
	"github.com/btcsuite/btcd/btcec/v2"
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

	unspentList, err := client.ListUnspent()
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
	log.Println("dataCM:", hex.EncodeToString(dataCM[:]))
	changeaddr, err := client.GetNewAddress("default")
	if err != nil {
		log.Fatalf("GetRawChangeAddress failed with error:%v", err)
	}
	dataList := make([]ord.InscriptionData, 0)
	dataList = append(dataList, ord.InscriptionData{
		ContentType: "text/plain;charset=utf-8",
		Body:        dataCM[:],
		Destination: changeaddr.String(),
	})

	// nodePubKey, _ := btcec.NewPrivateKey()
	// pubkeyHexStr := hex.EncodeToString(nodePubKey.PubKey().SerializeCompressed())
	// pubkeyHex, _ := hex.DecodeString(pubkeyHexStr)
	// newPubKey, _ := btcec.ParsePubKey(pubkeyHex)
	// fmt.Println("pk1", hex.EncodeToString(nodePubKey.PubKey().SerializeCompressed()))
	// fmt.Println("pk2", hex.EncodeToString(newPubKey.SerializeCompressed()))

	// nodeUrl := ""
	// rpcCli, err := rpc.DialOptions(context.Background(), nodeUrl)
	// if err != nil {
	// 	log.Fatalf("dial node failed, %v", err)
	// }
	// defer rpcCli.Close()

	pubkeyHexStr := ""
	pubkeyHex, _ := hex.DecodeString(pubkeyHexStr)
	nodePubKey, _ := btcec.ParsePubKey(pubkeyHex)
	sigNode := ord.SignNodeInfo{
		RpcClient: nil,
		PublicKey: nodePubKey,
	}

	sigNodes := make([]*ord.SignNodeInfo, 0)
	sigNodes = append(sigNodes, &sigNode)
	request := ord.InscriptionRequest{
		CommitTxOutPointList: commitTxOutPointList,
		CommitFeeRate:        25,
		FeeRate:              26,
		DataList:             dataList,
		SingleRevealTxOnly:   false,
		SignNodes:            sigNodes,
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
}
