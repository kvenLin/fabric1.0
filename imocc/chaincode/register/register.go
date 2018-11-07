package main

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"fmt"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"encoding/json"
	"bytes"
	"strings"
)

var logger = shim.NewLogger("register")

//链码编写步骤:
// 1.定义一个链码的结构体对象,
// 2.实现当前对象的Init和Invoke方法,
// 3.在main函数中调用shim.Start(new(链码对象))启动链码

type SimpleChaincode struct {
}

const (
	USER_NOT_EXIST = 404
	PASSWORD_ERROR = 403
	USER_ALREADY_EXIST = 501
	SUCCESS = "success"
	ERROR = "error"

)

type User struct {
	ID string `json:"id"`
	Password string `json:"password"`
	Role string `json:"role"`
	Balance float64 `json:"balance"`
	Points int `json:"crePoint"`
}

type UserResult struct {
	Code int `json:"code"`
	Message string `json:"message"`
	User User `json:"user"`
}

type NormalResult struct {
	Code int `json:"code"`
	Message string `json:"message"`
}
//
//type TxInfo struct {
//	TxId string `json:"tx_id"`
//	Value string `json:"value"`
//	Timestamp Timestamp `json:"timestamp"`
//}
//
//type Timestamp struct {
//	seconds int64 `json:"seconds"`
//	nanos int64 `json:"nanos"`
//}
//
//type TxResult struct {
//	Code int `json:"code"`
//	Message string `json:"message"`
//	TxInfos []TxInfo `json:"tx_infos"`
//}
func getErrorResult(reason int) []byte{
	logger.Info()
	result :=  NormalResult{reason,ERROR}
	byte,_ := json.Marshal(result)
	return byte
}


func getUserSuccessResult(user User) []byte{
	var userResult UserResult
	userResult.Code = 0
	userResult.Message = SUCCESS
	userResult.User = user
	reByte,_ := json.Marshal(userResult)
	return reByte
}

func getNormalSuccessResult() []byte{
	var normalResult NormalResult
	normalResult.Code = 0
	normalResult.Message = SUCCESS
	reByte,_ := json.Marshal(normalResult)
	return reByte
}


//安装Chaincode
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Info("############# Register Init ##############")
	return shim.Success(nil)
}

//Invoke interface
func (t *SimpleChaincode)Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	logger.Info("################# Register Invoke ###############")
	function,args := stub.GetFunctionAndParameters()
	switch function {
	case "regist":
		return t.regist(stub,args)
	case "login":
		return t.login(stub,args)
	case "changePwd":
		return t.changePwd(stub,args)
	case "update":
		return t.update(stub,args)
	case "transMoney":
		return t.transMoney(stub,args)
	case "getPoints":
		return t.getPoints(stub,args)
	case "delete":
		return t.delete(stub,args)
	case "query":
		return t.query(stub,args)
	case "getHistoryForKey":
		return t.getHistoryForKey(stub,args)
	case "queryTxInfo":
		return t.queryTxInfo(stub,args)
	default:
		return shim.Error(fmt.Sprintf("unsupported function :%s ",function))
	}

	return shim.Success(nil)
}

func argsNumberErrorMsg(useMethod string,num int) pb.Response{
	msg := fmt.Sprintf("Incorrect number of args. Expecting %d", num)
	logger.Info(msg+","+useMethod)
	return shim.Error(msg+","+useMethod)
}

//==================
//创建账号 args:ID | Password | Role | Balance | Points
//==================
func (t *SimpleChaincode) regist(stub shim.ChaincodeStubInterface,args []string) pb.Response {
	logger.Info("regist args: ",args)
	if len(args) != 5{
		return argsNumberErrorMsg("useMethod: function on followed by 1 accountID and 4 value",5)
	}
	var accountID string
	var Password,Role string
	//var Balance,Points string //账户金额,积分
	var err error
	accountID = args[0]
	Password = args[1]
	Role = args[2]
	
	var user User
	user.ID = accountID
	user.Password = Password
	user.Role = Role
	Balance,_ := strconv.ParseFloat(args[3],64)
	user.Balance =  Balance
	Points,_ := strconv.Atoi(args[4])
	user.Points = Points
	uBytes,_ := json.Marshal(user)
	//get the state from ledger
	Avalbytes,err := stub.GetState(accountID)
	fmt.Println(string(Avalbytes))
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes != nil {
		return shim.Success(getErrorResult(USER_ALREADY_EXIST))
	}

	//Write the state back to the ledger
	err = stub.PutState(accountID, uBytes)
	if err != nil {
		shim.Error(err.Error())
	}
	return shim.Success(getNormalSuccessResult())

}


//================
//验证账号密码是否匹配,登录 args:ID|Password
//================
func (t *SimpleChaincode) login(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("login args:",args)
	if len(args) != 2 {
		return argsNumberErrorMsg("function need accountID and password",2)
	}
	accountID := args[0]
	password := args[1]
	//query the ledger
	bytes,err := stub.GetState(accountID)
	if err != nil {
		return shim.Error("Failed to get account: "+err.Error())
	}
	if bytes == nil {
		return shim.Success(getErrorResult(USER_NOT_EXIST))
	}

	var user User
	err = json.Unmarshal(bytes,&user)
	if err != nil {
		return shim.Error("Failed to unmarshal user"+err.Error())
	}
	if user.Password == password {
		return shim.Success(getNormalSuccessResult())
	}else {
		return shim.Success(getErrorResult(PASSWORD_ERROR))
	}
}

//==============================
//更改用户密码 args:ID| OldPassword |newPassword
//==============================
func (t *SimpleChaincode) changePwd(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("changing password args :",args)
	if len(args) != 3 {
		return argsNumberErrorMsg("function need ID,OldPassword,NewPassword",3)
	}
	userID := args[0]
	oldPassword := args[1]
	newPassword := args[2]
	var err error
	bytes,err := stub.GetState(userID)
	if err!=nil {
		return shim.Error("Failed get account:"+err.Error())
	}
	if bytes == nil {
		return shim.Success(getErrorResult(USER_NOT_EXIST))
	}
	var user User
	err = json.Unmarshal(bytes,&user)
	if err != nil {
		return shim.Error("Failed to unmarshal user:"+err.Error())
	}
	if user.Password == oldPassword {
		user.Password = newPassword
	}else {
		return shim.Success(getErrorResult(PASSWORD_ERROR))
	}
	Bytes,_ := json.Marshal(user)
	err = stub.PutState(user.ID,Bytes)
	if err != nil {
		return shim.Error("put to state error:"+err.Error())
	}
	return shim.Success(getNormalSuccessResult())

}
//==============================
//更新账户信息 args:ID| Balance | Points
//==============================
func (t *SimpleChaincode) update(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("update user Balance or Points args:",args)
	if len(args) != 3 {
		return argsNumberErrorMsg("function args need ID,Balance,Points",3)
	}
	userID := args[0]
	Balance,err := strconv.ParseFloat(args[1],64)
	if err != nil {
		return shim.Error("Failed to convert balance from string to float64:"+err.Error())
	}
	Points,err := strconv.Atoi(args[2])
	if err != nil{
		return shim.Error("Failed to convert points from string to int:"+err.Error())
	}
	bytes,err := stub.GetState(userID)
	if err != nil {
		return shim.Error("Failed get account:"+userID)
	}
	if bytes == nil {
		return shim.Success(getErrorResult(USER_NOT_EXIST))
	}
	var user User
	err = json.Unmarshal(bytes,&user)
	if err != nil {
		return shim.Error("unmarshal user fialed :"+err.Error())
	}
	user.Balance = Balance
	user.Points = Points
	Bytes,_ := json.Marshal(user)
	err = stub.PutState(userID,Bytes)
	if err != nil {
		return shim.Error("Failed to put user to state :"+err.Error())
	}
	return shim.Success(getNormalSuccessResult())
}

//================================
//查询账号 args:ID
//================================
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return argsNumberErrorMsg("function just need arg ID",1)
	}
	var ID string
	var err error
	ID = args[0]
	Avalbytes,err := stub.GetState(ID)
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get state for " + ID + "\"}"
		return shim.Error(jsonResp)
	}
	if Avalbytes == nil {
		return shim.Success(getErrorResult(USER_NOT_EXIST))
	}
	jsonResp := string(Avalbytes)
	fmt.Printf("Query Response:%s \n",jsonResp)
	var user User
	json.Unmarshal(Avalbytes,&user)
	return shim.Success(getUserSuccessResult(user))
}

//==============================
//转账 args:buyerID| sellerID | Price
//==============================
func (t *SimpleChaincode) transMoney(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("transMoney args:",args)
	if len(args) != 3 {
		return argsNumberErrorMsg("function args need buyerID,sellerID,price",3)
	}
	buyerID := args[0]
	sellerID := args[1]
	price,err := strconv.ParseFloat(args[2],64)
	if err != nil {
		return shim.Error("Failed to convert price from string to float: "+err.Error())
	}
	var buyer,seller User
	//query the ledger
	bytes,err := stub.GetState(buyerID)
	if err != nil {
		return shim.Error("Failed get buyer :"+err.Error())
	}
	if bytes == nil {
		return shim.Error("account does not exist buyerID :"+buyerID)
	}
	err = json.Unmarshal(bytes,&buyer)
	if err != nil {
		return shim.Error("Failed to unmarshal buyer:"+err.Error())
	}
	Bytes,err := stub.GetState(sellerID)
	if err != nil {
		return shim.Error("Failed to get seller :"+err.Error())
	}
	if Bytes == nil {
		return shim.Error("account does not exist sellerID:"+sellerID)
	}
	err = json.Unmarshal(Bytes,&seller)
	if err != nil {
		return shim.Error("Failed to unmarshal seller:"+err.Error())
	}
	buyer.Balance = buyer.Balance - price
	Bytes,_ = json.Marshal(buyer)
	err = stub.PutState(buyerID,Bytes)
	if err != nil {
		return shim.Error("Failed to put buyer to state :"+err.Error())
	}
	seller.Balance = seller.Balance + price
	Bytes,_ = json.Marshal(seller)
	err = stub.PutState(sellerID,Bytes)
	if err != nil {
		return shim.Error("Failed to put seller to state :"+err.Error())
	}
	return shim.Success(getNormalSuccessResult())
}

//==============================
//增加积分 args:transerID | Points
//==============================
func (t *SimpleChaincode) getPoints(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("getPoints args :",args)
	if len(args) != 2 {
		return argsNumberErrorMsg("function args need transerID and Points",2)
	}
	transerID := args[0]
	Points,err := strconv.Atoi(args[1])

	if err != nil {
		return shim.Error("Failed convert point from string to float:"+err.Error())
	}
	var transer User
	bytes,err := stub.GetState(transerID)
	if err != nil {
		return shim.Error("Failed to user from state :"+err.Error())
	}
	if bytes == nil {
		return shim.Success(getErrorResult(USER_NOT_EXIST))
	}
	err = json.Unmarshal(bytes,&transer)
	if err != nil {
		return shim.Error("Failed unmarshal transer :"+err.Error())
	}
	transer.Points = transer.Points + Points
	Bytes,_ := json.Marshal(transer)
	err = stub.PutState(transerID,Bytes)
	if err != nil {
		return shim.Error("Failed to put transer to state:"+err.Error())
	}
	return shim.Success(getNormalSuccessResult())
}

//================
//删除账号 args：userID
//================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("delete user args:",args)
	if len(args) != 1 {
		return argsNumberErrorMsg("function need arg userID ",1)
	}
	userID := args[0]
	bytes,err := stub.GetState(userID)
	if err != nil {
		return shim.Error("Failed to get user from state:"+err.Error())
	}
	if bytes == nil {
		return shim.Success(getErrorResult(USER_NOT_EXIST))
	}
	var user User
	user =  User{}
	Bytes,err := json.Marshal(user)
	if err != nil {
		return shim.Error("Failed to marshal user:"+err.Error())
	}
	err = stub.PutState(userID,Bytes)
	if err != nil {
		return shim.Error("Failed to put user to state :"+err.Error())
	}
	err = stub.DelState(userID)
	if err != nil {
		return shim.Error("Failed to delete user :"+err.Error())
	}
	return shim.Success(getNormalSuccessResult())
}

//==================================
//查看历史消息 args: userID
//==================================
func (t *SimpleChaincode) getHistoryForKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return argsNumberErrorMsg("function need arg userID ",1)
	}
	var userID string
	userID = args[0]
	HisInterface,err := stub.GetHistoryForKey(userID)
	fmt.Println(HisInterface)
	historyStrings,err := getHistoryListResult(HisInterface)
	byteContent := strings.Join(historyStrings,"")
	if err != nil {
		return shim.Error("Failed to get user history:"+err.Error())
	}
	return shim.Success([]byte(byteContent))
}

//=================================
//查看单笔交易信息 args: txID
//===================================
func (t *SimpleChaincode) queryTxInfo(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return argsNumberErrorMsg("function need arg tx_id",1)
	}
	var txId string
	txId = args[0]
	infoByte,err := stub.GetState(txId)
	if err != nil {
		return shim.Error("Failed to get txInfo from state:"+err.Error())
	}
	return shim.Success(infoByte)
}

func getHistoryListResult(resultsIterator shim.HistoryQueryIteratorInterface)([]string,error){
	defer resultsIterator.Close()
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
		//item, _ := json.Marshal(queryResponse)
		logger.Info("queryResponse:",queryResponse)
		//logger.Info("history item:",item)
		//buffer.Write(item)
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")
	fmt.Printf("queryResult:\n%s\n", buffer.String())
	//return buffer.Bytes(), nil
	return responses,nil
}


func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil{
		fmt.Sprintf("Error starting SimpleChaincode:%s",err)
	}
}