package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"fmt"
)

var logger = shim.NewLogger("trance")

type SimpleChaincode struct {}

//药品
type Drug struct {
	ID string `json:"id"`
	OwnerID string `json:"ownerId"`
	Name string `json:"name"`
	Price float64 `json:"price"`
	Description string `json:"description"`
	Buyer string `json:"buyer"`
	Traces []Trace `json:"traces"`
}

//溯源消息
type Trace struct {
	TransID string `json:"transId"`
	Place string `json:"place"`
	TimeStamp string `json:"timeStamp"`
}



//链码初始化
func (c *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response  {
	logger.Info("#########  Trance Init ###########")
	return shim.Success(nil)
}

//链码交互的具体方法
func (c *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("########## Trance Invoke ###############")
	function,args := stub.GetFunctionAndParameters()
	switch function {
	case "drugInit":
		return c.drugInit(stub,args)
	case "trans":
		return c.trans(stub,args)
	case "buy":
		return c.buy(stub,args)
	case "query":
		return c.query(stub,args)
	case "queryHistoryForKey":
		return c.getHistoryForKey(stub,args)
	case "testRangeQuery":
		return c.testRangeQuery(stub,args)
	default:
		return shim.Error(fmt.Sprintf("unsupported function :%s ", function))
	}
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting SimpleChaincode: %s",err)
	}
}