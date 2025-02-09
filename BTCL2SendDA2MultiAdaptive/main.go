package main

import (
	"context"
	"encoding/hex"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/MultiAdaptive/go-ord-tx/ord"
	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/btcsuite/btcd/wire"
	kzg_sdk "github.com/domicon-labs/kzg-sdk"
	"github.com/ethereum/go-ethereum/rpc"
)

const dSrsSize = 1 << 16

func main() {
	host := os.Getenv("BTCHOST")
	user := os.Getenv("BTCUSER")
	pass := os.Getenv("BTCPASS")
	if host == "" || user == "" || pass == "" {
		log.Fatal("please set environments: BTCHOST, BTCUSER, BTCPASS")
	}

	maNodeRPC1 := os.Getenv("MULTIADAPTIVENODERPC1")
	maNodePubKey1 := os.Getenv("MULTIADAPTIVENODEPUBKEY1")
	maNodeRPC2 := os.Getenv("MULTIADAPTIVENODERPC2")
	maNodePubKey2 := os.Getenv("MULTIADAPTIVENODEPUBKEY2")
	if maNodeRPC1 == "" || maNodePubKey1 == "" || maNodeRPC2 == "" || maNodePubKey2 == "" {
		log.Fatal("please set envrionments: MULTIADAPTIVENODERPC1, MULTIADAPTIVENODEPUBKEY1, MULTIADAPTIVENODERPC2, MULTIADAPTIVENODEPUBKEY2 ")
	}
	netParams := &chaincfg.RegressionNetParams
	connCfg := &rpcclient.ConnConfig{
		Host:         host,
		User:         user,
		Pass:         pass,
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
	rand.Seed(time.Now().UnixNano())
	originData := make([]byte, 1024*1024*5)
	for i := range originData {
		originData[i] = uint8(rand.Intn(256))
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

	rpcCli1, err := rpc.DialOptions(context.Background(), maNodeRPC1)
	if err != nil {
		log.Fatalf("dial node failed, %v", err)
	}
	defer rpcCli1.Close()

	pubkeyHex1, _ := hex.DecodeString(maNodePubKey1)
	nodePubKey1, _ := btcec.ParsePubKey(pubkeyHex1)
	sigNode1 := ord.SignNodeInfo{
		RpcClient: rpcCli1,
		PublicKey: nodePubKey1,
	}

	rpcCli2, err := rpc.DialOptions(context.Background(), maNodeRPC2)
	if err != nil {
		log.Fatalf("dial node failed, %v", err)
	}
	defer rpcCli2.Close()
	pubkeyHex2, _ := hex.DecodeString(maNodePubKey2)
	nodePubKey2, _ := btcec.ParsePubKey(pubkeyHex2)
	sigNode2 := ord.SignNodeInfo{
		RpcClient: rpcCli2,
		PublicKey: nodePubKey2,
	}

	sigNodes := make([]*ord.SignNodeInfo, 0)
	sigNodes = append(sigNodes, &sigNode1)
	sigNodes = append(sigNodes, &sigNode2)

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
