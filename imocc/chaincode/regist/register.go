package regist

import (
	"github.com/hyperledger/fabric/core/chaincode/shim"
	"fmt"
	pb "github.com/hyperledger/fabric/protos/peer"
	"strconv"
	"encoding/json"
)

var logger = shim.NewLogger("register")

//链码编写步骤:
// 1.定义一个链码的结构体对象,
// 2.实现当前对象的Init和Invoke方法,
// 3.在main函数中调用shim.Start(new(链码对象))启动链码

type SimpleChaincode struct {
}


type User struct {
	ID string `json:"id"`
	Password string `json:"password"`
	Role string `json:"role"`
	Balance float64 `json:"balance"`
	Points int `json:"crePoint"`
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
		return shim.Error("this user already exist")
	}

	//Write the state back to the ledger
	err = stub.PutState(accountID, uBytes)
	if err != nil {
		shim.Error(err.Error())
	}
	return shim.Success([]byte(accountID+"帐号创建成功! "))

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
		return shim.Error("This account does not exists : "+accountID)
	}

	var user User
	err = json.Unmarshal(bytes,&user)
	if err != nil {
		return shim.Error("Failed to unmarshal user"+err.Error())
	}
	if user.Password == password {
		return shim.Success([]byte("correct password"))
	}else {
		return shim.Error("Wrong password!")
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
		return shim.Error("user does not exist:"+userID)
	}
	var user User
	err = json.Unmarshal(bytes,&user)
	if err != nil {
		return shim.Error("Failed to unmarshal user:"+err.Error())
	}
	if user.Password == oldPassword {
		user.Password = newPassword
	}else {
		return shim.Error("user password wrong!")
	}
	Bytes,_ := json.Marshal(user)
	err = stub.PutState(user.ID,Bytes)
	if err != nil {
		return shim.Error("put to state error:"+err.Error())
	}
	return shim.Success([]byte("更新成功!"))

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
		return shim.Error("this account does not exists:"+userID)
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
	return shim.Success([]byte("修改账户成功!"))
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
		jsonResp := "{\"Error\":\"Nil count for " + ID + "\"}"
		return shim.Error(jsonResp)
	}
	jsonResp := string(Avalbytes)
	fmt.Printf("Query Response:%s \n",jsonResp)
	return shim.Success(Avalbytes)
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
		shim.Error("account does not exist buyerID :"+buyerID)
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
	return shim.Success([]byte("转账成功!"))
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
		return shim.Error("user does not exist transerID:"+transerID)
	}
	err = json.Unmarshal(bytes,transer)
	if err != nil {
		return shim.Error("Failed unmarshal transer :"+err.Error())
	}
	transer.Points = transer.Points + Points
	Bytes,_ := json.Marshal(transer)
	err = stub.PutState(transerID,Bytes)
	if err != nil {
		return shim.Error("Failed to put transer to state:"+err.Error())
	}
	return shim.Success([]byte("增加积分成功"))
}

//================
//删除账号 args：userID
//================
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {

}
func (t *SimpleChaincode) getHistoryForKey(stub shim.ChaincodeStubInterface, args []string) pb.Response {

}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil{
		fmt.Sprintf("Error starting SimpleChaincode:%s",err)
	}
}