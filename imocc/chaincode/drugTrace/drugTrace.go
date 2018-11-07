package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	pb "github.com/hyperledger/fabric/protos/peer"
	"fmt"
	"strconv"
	"encoding/json"
	"time"
	"bytes"
	"strings"
)

var logger = shim.NewLogger("trance")

type SimpleChaincode struct {}

const (
	USER_NOT_EXIST      = 404
	PASSWORD_ERROR      = 403
	USER_ALREADY_EXIST  = 501
	DRUG_ALREADY_EXIST  = 502
	DRUG_DOES_NOT_EXIST = 503
	TRANS_MONEY_ERROR   = 504
	SUCCESS             = "success"
	ERROR               = "error"

)

//药品
type Drug struct {
	ID string `json:"id"`
	OwnerID string `json:"ownerId"`
	Name string `json:"name"`
	Price float64 `json:"price"`
	FileHash string `json:"file_hash"`
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



type DrugResult struct {
	Code int `json:"code"`
	Message string `json:"message"`
	Drug Drug `json:"drug"`
}

type NormalResult struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

func getErrorResult(reason int) []byte{
	logger.Info()
	result :=  NormalResult{reason,ERROR}
	byte,_ := json.Marshal(result)
	return byte
}


func getDrugSuccessResult(drug Drug) []byte{
	var drugResult DrugResult
	drugResult.Code = 0
	drugResult.Message = SUCCESS
	drugResult.Drug = drug
	reByte,_ := json.Marshal(drugResult)
	return reByte
}

func getNormalSuccessResult() []byte{
	var normalResult NormalResult
	normalResult.Code = 0
	normalResult.Message = SUCCESS
	reByte,_ := json.Marshal(normalResult)
	return reByte
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

func argsNumberErrorMsg(useMethod string,num int) pb.Response{
	msg := fmt.Sprintf("Incorrect number of args. Expecting %d", num)
	logger.Info(msg+","+useMethod)
	return shim.Error(msg+","+useMethod)
}

//==============================================================================================
//发布一种药品 args:ID | OwnerID | Name | Price | FileHash | Description
//==============================================================================================
func (c *SimpleChaincode) drugInit(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 6 {
		return argsNumberErrorMsg("function need args ID | OwnerID | Name | Price | FileHash | Description",6)
	}
	var drugID string
	var OwnerID,Name string
	var Price float64
	var FileHash string
	var Description string
	var err error
	drugID = args[0]
	OwnerID = args[1]
	Name = args[2]
	Price,err = strconv.ParseFloat(args[3],64)
	if err != nil {
		return shim.Error("Failed convert Price from string to float :"+err.Error())
	}
	FileHash = args[4]
	Description = args[5]
	var drug Drug
	drug.ID = drugID
	drug.OwnerID = OwnerID
	drug.Name = Name
	drug.Price = Price
	drug.FileHash = FileHash
	drug.Description = Description

	dBytes,err := stub.GetState(drugID)
	if err != nil {
		return shim.Error("Failed to get drug from state:"+err.Error())
	}
	if dBytes != nil {
		//TODO
		return shim.Success(getErrorResult(DRUG_ALREADY_EXIST))
	}
	bytes,err := json.Marshal(drug)
	if err != nil {
		return shim.Error("Failed to marshal drug:"+err.Error())
	}
	err = stub.PutState(drugID,bytes)
	if err != nil {
		return shim.Error("Failed to put drug to state:"+err.Error())
	}
	return shim.Success(getNormalSuccessResult())
}

//============================================
//追踪物流路径 args: drugID | agencyID:代理商用户ID | place
//============================================
func (c *SimpleChaincode) trans(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return argsNumberErrorMsg("function need args drugID | agencyID | place",3)
	}
	drugID := args[0]
	agencyID := args[1]
	place := args[2]
	dBytes,err := stub.GetState(drugID)
	if err != nil {
		return shim.Error("Failed to get drug from state:"+err.Error())
	}
	if dBytes == nil {
		return shim.Success(getErrorResult(DRUG_DOES_NOT_EXIST))
	}
	var drug Drug
	err = json.Unmarshal(dBytes,&drug)
	if err != nil {
		return shim.Error("Failed to unmarshal drug:"+err.Error())
	}
	var trace Trace
	trace.TransID = agencyID
	//获取当前时间戳
	currentTime := time.Unix(time.Now().UnixNano()/1e9, 0)
	trace.TimeStamp = currentTime.String()
	trace.Place = place
	drug.Traces = append(drug.Traces,trace)
	dBytes,err = json.Marshal(drug)
	if err != nil {
		return shim.Error("Failed to marshal drug :"+err.Error())
	}
	err = stub.PutState(drugID,dBytes)
	if err != nil {
		return shim.Error("Failed to put drug to state:"+err.Error())
	}
	//给提供溯源信息的代理商增加5积分
	trans := [][]byte{[]byte("getPoints"),[]byte(agencyID),[]byte("5")}
	response := stub.InvokeChaincode("register",trans,"mychannel")
	if response.Status != int32(200) {
		return shim.Error("Failed to get addComment :"+err.Error())
	}
	return shim.Success(getNormalSuccessResult())
}


//============================================
//消费者购买 args: drugID | buyerID | endPlace
//============================================
func (c *SimpleChaincode) buy(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return argsNumberErrorMsg("function need args drugID | buyerID | endPlace",+3)
	}
	drugID := args[0]
	buyerID := args[1]
	endPlace := args[2]

	//第一步:添加药品的溯源信息
	dBytes,err := stub.GetState(drugID)
	if err != nil {
		return shim.Error("Failed to get drug from state:"+err.Error())
	}
	if dBytes == nil {
		return shim.Success(getErrorResult(DRUG_DOES_NOT_EXIST))
	}
	var drug Drug
	err = json.Unmarshal(dBytes,&drug)
	if err != nil {
		return shim.Error("Failed to unmarshal drug :"+err.Error())
	}
	var trace Trace
	trace.TransID = buyerID
	currentTime := time.Unix(time.Now().UnixNano()/1e9, 0)
	trace.TimeStamp = currentTime.String()
	trace.Place = endPlace
	drug.Traces = append(drug.Traces,trace)
	drug.Buyer = buyerID
	drug.OwnerID = buyerID
	sellerID := drug.OwnerID
	dBytes,err = json.Marshal(drug)
	if err != nil {
		return shim.Error("Failed to drug:"+err.Error())
	}
	err = stub.PutState(drugID,dBytes)
	if err != nil {
		return shim.Error("Failed to put drug to state:"+err.Error())
	}
	//第二步:调用转账链码
	//获取药品的价格并转换成字符串
	price := strconv.FormatFloat(drug.Price,'f',10,64)
	transMoney := [][]byte{[]byte("transMoney"),[]byte(buyerID),[]byte(sellerID),[]byte(price)}
	response := stub.InvokeChaincode("register",transMoney,"mychannel")
	if response.Status != int32(200) {
		return shim.Success(getErrorResult(TRANS_MONEY_ERROR))
	}
	return shim.Success(getNormalSuccessResult())
}


//===========================
//查询药品信息 args: drugID
//============================
func (c *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return argsNumberErrorMsg("function need arg drugID",1)
	}
	drugID := args[0]
	dBytes,err := stub.GetState(drugID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for drugID " + drugID + "\"}"
		return shim.Error(jsonResp)
	}
	if dBytes == nil {
		jsonResp := "{\"Error\":\"Does not exist drugID " + drugID + "\"}"
		logger.Error(jsonResp)
		return shim.Success(getErrorResult(DRUG_DOES_NOT_EXIST))
	}
	jsonResp := string(dBytes)
	fmt.Sprintf("Query response:%s\n",jsonResp)
	var drug Drug
	err = json.Unmarshal(dBytes,&drug)
	if err != nil {
		return shim.Error("unmarshal error!")
	}
	return shim.Success(getDrugSuccessResult(drug))
}

//===========================
//通过key查看历史记录 args: drugID
//TODO,理解查出来的hash是什么
//===========================
func (c *SimpleChaincode) getHistoryForKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return argsNumberErrorMsg("function need arg drugID",1)
	}
	drugID := args[0]
	HisInterface, err := stub.GetHistoryForKey(drugID)
	if err != nil {
		return shim.Error("Failed to get hisInterface!")
	}
	historyStrings,err := getHistoryListResult(HisInterface)
	byteContent := strings.Join(historyStrings,"")
	if err != nil {
		return shim.Error("Failed to get historyListResult!")
	}
	return shim.Success([]byte(byteContent))
}

//==================================
//范围查询，args: 起始ID | 终止ID
//==================================
func (c *SimpleChaincode) testRangeQuery(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	startID := args[0]
	endID := args[1]
	resultsIterator, err := stub.GetStateByRange(startID, endID)
	if err != nil {
		return shim.Error("Query by Range failed")
	}
	services, err := getListResult(resultsIterator)
	if err != nil {
		return shim.Error("getListResult failed")
	}
	return shim.Success(services)
}


func getListResult(resultsIterator shim.StateQueryIteratorInterface) ([]byte, error) {

	defer resultsIterator.Close()
	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Printf("queryResult:\n%s\n", buffer.String())
	return buffer.Bytes(), nil
}


func getHistoryListResult(resultsIterator shim.HistoryQueryIteratorInterface) ([]string, error) {

	defer resultsIterator.Close()
	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	var responses []string
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		responses = append(responses, queryResponse.String())
		item, _ := json.Marshal(queryResponse)
		buffer.Write(item)
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Printf("queryResult:\n%s\n", buffer.String())
	return responses, nil
}






func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting SimpleChaincode: %s",err)
	}
}