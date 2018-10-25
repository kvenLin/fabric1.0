package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"bytes"
	"strconv"
	"math/rand"
	"time"
	"fmt"
)

type BadExampleCC struct {}
//链码的初始化
func (c *BadExampleCC) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (c *BadExampleCC) Invoke(stub shim.ChaincodeStubInterface) pb.Response{
	return shim.Success(bytes.NewBufferString(strconv.Itoa(int(rand.Int63n(time.Now().Unix())))).Bytes())
}

func main() {
	//启动链码
	err := shim.Start(new(BadExampleCC))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}