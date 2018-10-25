# 笔记

## 设置环境变量
```bash
export FBRIC_CFG_PATH=$GOPATH/src/github.com/hyperledger/fabric/imocc/deploy
```

## 环境清理
```bash
rm -rf config/*
rm -rf crypto-config/*
```
## 生成证书文件
```bash
cryptogen generate --config=./crypto-config.yaml
```
## 生成创世区块
``` bash
configtxgen -profile OneOrgOrdererGenesis -outputBlock ./config/genesis.block
```

## 生成通道的创世交易
``` bash
configtxgen -profile TwoOrgChannel -outputCreateChannelTx ./config/mychannel.tx -channelID mychannel
configtxgen -profile TwoOrgChannel -outputCreateChannelTx ./config/assetschannel.tx -channelID assetschannel
```

## 生成组织关于通道的锚节点（主节点）交易
``` bash
configtxgen -profile TwoOrgChannel -outputAnchorPeersUpdate ./config/Org0MSPanchors.tx -channelID mychannel -asOrg Org0MSP
configtxgen -profile TwoOrgChannel -outputAnchorPeersUpdate ./config/Org1MSPanchors.tx -channelID mychannel -asOrg Org1MSP
configtxgen -profile TwoOrgChannel -outputAnchorPeersUpdate ./config/Org0MSPanchors.tx -channelID assetschannel -asOrg Org0MSP
configtxgen -profile TwoOrgChannel -outputAnchorPeersUpdate ./config/Org1MSPanchors.tx -channelID assetschannel -asOrg Org1MSP
```

#创建新通道只需要进行后续操作
## 创建通道
``` bash
peer channel create -o orderer.imocc.com:7050 -c mychannel -f /etc/hyperledger/config/mychannel.tx
peer channel create -o orderer.imocc.com:7050 -c assetschannel -f /etc/hyperledger/config/assetschannel.tx
```

## 加入通道
``` bash
peer channel join -b mychannel.block
peer channel join -b assetschannel.block
```

## 设置主节点
``` bash
peer channel update -o orderer.imocc.com:7050 -c mychannel -f /etc/hyperledger/config/Org1MSPanchors.tx
peer channel update -o orderer.imocc.com:7050 -c assetschannel -f /etc/hyperledger/config/Org1MSPanchors.tx
```

## 链码安装
``` bash
peer chaincode install -n badexample -v 1.0.0 -l golang -p github.com/chaincode/badexample
peer chaincode install -n assets -v 1.0.0 -l golang -p github.com/chaincode/assetsExchange

升级操作
peer chaincode install -n assets -v 1.0.1 -l golang -p github.com/chaincode/assetsExchange
```

## 链码实例化
``` bash
peer chaincode instantiate -o orderer.imocc.com:7050 -C mychannel -n badexample -l golang -v 1.0.0 -c '{"Args":["init"]}'
peer chaincode instantiate -o orderer.imocc.com:7050 -C assetschannel -n assets -l golang -v 1.0.0 -c '{"Args":["init"]}'

//加入背书策略
peer chaincode instantiate -o orderer.imocc.com:7050 -C assetschannel -n assets -l golang -v 1.0.0 -c '{"Args":["init"]}' -P "OR('org0MSP.member','org1MSP.admin')"

```

## 链码交互
``` bash
peer chaincode invoke -C assetschannel -n assets -c '{"Args":["userRegister", "user1", "user1"]}'
peer chaincode invoke -C assetschannel -n assets -c '{"Args":["AssetEnroll", "asset1", "asset1", "metadata", "user1"]}'
peer chaincode invoke -C assetschannel -n assets -c '{"Args":["userRegister", "user2", "user2"]}'
peer chaincode invoke -C assetschannel -n assets -c '{"Args":["AssetExchange", "user1", "asset1", "user2"]}'
peer chaincode invoke -C assetschannel -n assets -c '{"Args":["userDestroy", "user1"]}'
```

## 链码升级
``` bash
peer chaincode install -n assets -v 1.0.1 -l golang -p github.com/chaincode/assetsExchange
peer chaincode upgrade -C assetschannel -n assets -v 1.0.1 -c '{"Args":["init"]}'
peer chaincode upgrade -C assetschannel -n assets -v 1.0.1 -P "OR('org1MSP.admin')" -c '{"Args":["init"]}'
```



## 链码查询
``` bash
peer chaincode query -C assetschannel -n assets -c '{"Args":["queryUser", "user1"]}'
peer chaincode query -C assetschannel -n assets -c '{"Args":["queryAsset", "asset1"]}'
peer chaincode query -C assetschannel -n assets -c '{"Args":["queryUser", "user2"]}'
peer chaincode query -C assetschannel -n assets -c '{"Args":["queryAssetHistory", "asset1"]}'
peer chaincode query -C assetschannel -n assets -c '{"Args":["queryAssetHistory", "asset1", "all"]}'
```

## 命令行模式的背书策略

EXPR(E[,E...])
EXPR = OR AND
E = EXPR(E[,E...])
MSP.ROLE
MSP 组织名 org0MSP org1MSP
ROLE admin member

OR('org0MSP.member','org1MSP.admin')

理解:一笔交易要想有效必须要有org0的某一个用户的签名或者说是org1的admin的签名,只有当交易是有效的才能被发往order节点进行排序
