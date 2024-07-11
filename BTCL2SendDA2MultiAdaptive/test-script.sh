#!/bin/bash
export btcRPC="http://13.228.170.151:18443/"
export BTCHOST="13.228.170.151:18443/wallet/test-wallet-1"
export BTCUSER="multiadaptiveUser1"
export BTCPASS="pwd123"
export BTCMINEADDR="bcrt1qatdxfq0ttd4kxhpxrgc07mpx2p0rpcaua26jht"
export MULTIADAPTIVENODERPC1="http://54.86.78.227:8545"
export MULTIADAPTIVENODEPUBKEY1="024063dc56b68904e2f7c0e4ddee10d2da9625d4bdf2fe0002cdf381bf3d13f7cb"
export MULTIADAPTIVENODERPC2="http://54.177.13.87:8545"
export MULTIADAPTIVENODEPUBKEY2="0398baebc991514b611a2e59b33de5a3a10b91b617e1056f1ffda4e0a7dfa6c342"

inscriptDACM="./sendBTCDA"
mineCommand="curl -X POST --user $BTCUSER:$BTCPASS -H 'content-type: text/plain;' $btcRPC --data-binary '{\"jsonrpc\": \"1.0\", \"id\":\"curltest\", \"method\": \"generatetoaddress\", \"params\": [1,\"$BTCMINEADDR\"] }'"

while true; do
  eval $inscriptDACM
  eval $mineCommand
  sleep 10
done
