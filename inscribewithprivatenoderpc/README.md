# notes

- step1
```sh
make clean
make build
```

- step2  
edit test-script.sh, fill the envrionments
```
export btcRPC="http://13.228.170.151:18443/"
export BTCHOST="13.228.170.151:18443/wallet/test-wallet-1"
export BTCUSER="multiadaptiveUser1"
export BTCPASS="pwd123"
export MULTIADAPTIVENODERPC="http://13.125.118.52:8545"
export MULTIADAPTIVENODEPUBKEY="020b0bae055c4e33c8561c080e8dd6c80b9f40f4a7fdf406c8c1da3b68dbc8a9f2"

```

- step3 
```sh
./test-script.sh
```