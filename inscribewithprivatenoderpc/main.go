package main

import (
	"context"
	"encoding/hex"
	"log"

	kzg_sdk "github.com/MultiAdaptive/kzg-sdk"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/vincentdebug/go-ord-tx/ord"
)

const dSrsSize = 1 << 16

func main() {
	netParams := &chaincfg.RegressionNetParams
	connCfg := &rpcclient.ConnConfig{
		Host:         "52.221.9.230:18332/wallet/newwallet.dat",
		User:         "testuser",
		Pass:         "123456",
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
	originData := make([]byte, 1024)
	for i := range originData {
		originData[i] = 1
	}
	cm, proof, err := kzgsdk.GenerateDataCommitAndProof(originData)
	if err != nil {
		log.Fatalf("kzg sdk GenerateDataCommitAndProof failed, %v", err)
	}
	dataCM := cm.Bytes()
	log.Println("dataCM:", hex.EncodeToString(dataCM[:]))
	dataProofH := proof.H.Bytes()
	dataProofClaimedValue := proof.ClaimedValue.Bytes()
	changeaddr, err := client.GetNewAddress("default")
	if err != nil {
		log.Fatalf("GetRawChangeAddress failed with error:%v", err)
	}

	dataList := make([]ord.InscriptionData, 0)
	dataList = append(dataList, ord.InscriptionData{
		ContentType:       "MultiAaptiveCM;charset=utf-8",
		DataCM:            dataCM[:],
		DataOrigin:        originData,
		Destination:       changeaddr.String(),
		ProofH:            dataProofH[:],
		ProofClaimedValue: dataProofClaimedValue[:],
	})

	nodeUrl := "http://13.125.118.52:8545"
	rpcCli, err := rpc.DialOptions(context.Background(), nodeUrl)
	if err != nil {
		log.Fatalf("dial node failed, %v", err)
	}
	defer rpcCli.Close()

	pubkeyHexStr := "020b0bae055c4e33c8561c080e8dd6c80b9f40f4a7fdf406c8c1da3b68dbc8a9f2"
	pubkeyHex, _ := hex.DecodeString(pubkeyHexStr)
	nodePubKey, _ := btcec.ParsePubKey(pubkeyHex)
	sigNode := ord.SignNodeInfo{
		RpcClient: rpcCli,
		PublicKey: nodePubKey,
	}

	sigNodes := make([]*ord.SignNodeInfo, 0)
	sigNodes = append(sigNodes, &sigNode)
	request := ord.InscriptionRequest{
		CommitTxOutPointList: commitTxOutPointList,
		CommitFeeRate:        200,
		FeeRate:              100,
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
