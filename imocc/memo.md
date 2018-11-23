# 笔记

### **整个网络搭建的学习笔记在Fabric1.0学习总结.docx文件中**

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

peer chaincode install -n register -v 1.0.0 -l golang -p github.com/chaincode/register
peer chaincode install -n drugTrace -v 1.0.0 -l golang -p github.com/chaincode/drugTrace


//升级操作
peer chaincode install -n assets -v 1.0.1 -l golang -p github.com/chaincode/assetsExchange
```

## 链码实例化
``` bash
peer chaincode instantiate -o orderer.imocc.com:7050 -C mychannel -n badexample -l golang -v 1.0.0 -c '{"Args":["init"]}'
peer chaincode instantiate -o orderer.imocc.com:7050 -C assetschannel -n assets -l golang -v 1.0.0 -c '{"Args":["init"]}'
peer chaincode instantiate -o orderer.imocc.com:7050 -C mychannel -n register -l golang -v 1.0.0 -c '{"Args":["init"]}'
peer chaincode instantiate -o orderer.imocc.com:7050 -C mychannel -n drugTrace -l golang -v 1.0.0 -c '{"Args":["init"]}'


//加入背书策略
peer chaincode instantiate -o orderer.imocc.com:7050 -C assetschannel -n assets -l golang -v 1.0.0 -c '{"Args":["init"]}' -P "OR('org0MSP.member','org1MSP.admin')"

```

## 链码交互
assetsExchange链码的交互
``` bash
peer chaincode invoke -C assetschannel -n assets -c '{"Args":["userRegister", "user1", "user1"]}'
peer chaincode invoke -C assetschannel -n assets -c '{"Args":["AssetEnroll", "asset1", "asset1", "metadata", "user1"]}'
peer chaincode invoke -C assetschannel -n assets -c '{"Args":["userRegister", "user2", "user2"]}'
peer chaincode invoke -C assetschannel -n assets -c '{"Args":["AssetExchange", "user1", "asset1", "user2"]}'
peer chaincode invoke -C assetschannel -n assets -c '{"Args":["userDestroy", "user1"]}'
```
register链码的交互
```bash
//创建用户,使用方法regist,用户id:201631063220,密码:123456,角色:user,账户:100,积分:0
peer chaincode invoke -C mychannel -n register -c '{"Args":["regist", "201631063220", "123456","user","100","0"]}'

//用户登录校验,使用方法login,用户id:201631063220,密码:123456
peer chaincode invoke -C mychannel -n register -c '{"Args":["login", "201631063220", "123456"]}'

//更改密码,使用changePwd,用户id:201631063220,旧密码:123456,新密码:123456789
peer chaincode invoke -C mychannel -n register -c '{"Args":["changePwd", "201631063220", "123456","123456789"]}'

//更新账户信息,使用方法update,用户id:201631063220,账户余额:200,积分:10
peer chaincode invoke -C mychannel -n register -c '{"Args":["update", "201631063220","200","10"]}'

//查询用户信息,使用方法query,用户id:201631063220
peer chaincode invoke -C mychannel -n register -c '{"Args":["query", "201631063220"]}'

//转账,使用方法transMoney,购买者id:201631063220,出售者id:201631063221,药品价格:10
peer chaincode invoke -C mychannel -n register -c '{"Args":["transMoney", "201631063220","201631063221","10"]}'

//增加用户积分,使用方法getPoints,用户id(即提供溯源信息用户的id):201631063221,积分值:10
peer chaincode invoke -C mychannel -n register -c '{"Args":["getPoints", "201631063221","10"]}'

//删除用户,使用方法delete,用户id:201631063221
peer chaincode invoke -C mychannel -n register -c '{"Args":["delete", "201631063221"]}'

//查询用户交易历史记录,使用方法getHistoryForKey,用户id:201631063220
peer chaincode invoke -C mychannel -n register -c '{"Args":["getHistoryForKey", "201631063220"]}'
```

drugTrace链码的交互
```bash
//药品上链的初始化,使用方法drugInit,药品id:1,拥有者id:201631063220,药品名称:testDrug,价格:10,文件hash值:dsagkjksahhjds,描述信息:this is a drug for test chaincode
peer chaincode invoke -C mychannel -n drugTrace -c '{"Args":["drugInit", "1","201631063220","testDrug","10","dsagkjksahhjds","this is a drug for test chaincode"]}'

peer chaincode invoke -C mychannel -n drugTrace -c '{"Args":["drugInit", "2","201631063221","testDrug2","10","dsagkjkdsdsaasd","this is another one drug for test chaincode"]}'

//代理商添加溯源信息,使用方法trans,药品id:1,代理商id:201631063221,地点:chengdu
peer chaincode invoke -C mychannel -n drugTrace -c '{"Args":["trans", "1","201631063221","chengdu"]}'

//购买药品,使用方法buy,药品id:1,购买者id:201631063222,购买地点:swpu
peer chaincode invoke -C mychannel -n drugTrace -c '{"Args":["buy", "1","201631063222","swpu"]}'

//查询药品信息,使用方法query,药品id:1
peer chaincode invoke -C mychannel -n drugTrace -c '{"Args":["query", "1"]}'

//通过key查看药品的历史记录,药品id:1
peer chaincode invoke -C mychannel -n drugTrace -c '{"Args":["queryHistoryForKey", "1"]}'

//范围查询,使用方法testRangeQuery,起始id:1,结束id:3 ######## 查询起始id为1到终止id为3之前的所有信息
peer chaincode invoke -C mychannel -n drugTrace -c '{"Args":["testRangeQuery", "1","3"]}'
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

## 链码调试
*注意:需要在chaincode目录下对应的链码路径使用命令*
```bash
//引入docker-compose的环境变量
export COMPOSE_HTTP_TIMEOUT=12000
//调试register
CORE_CHAINCODE_ID_NAME=register:1.0.0 CORE_PEER_ADDRESS=0.0.0.0:27051 CORE_CHAINCODE_LOGGING_LEVEL=DEBUG go run -tags=nopkcs11 register.go
//测试drugTrace
CORE_CHAINCODE_ID_NAME=drugTrace:1.0.0 CORE_PEER_ADDRESS=0.0.0.0:27051 CORE_CHAINCODE_LOGGING_LEVEL=DEBUG go run -tags=nopkcs11 drugTrace.go
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
