package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/mileusna/crontab"

	"github.com/fuzious/blockchainIndexer/connectionhelper"
	"github.com/fuzious/blockchainIndexer/token"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Database struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	From   string `json:"from,omitempty" bson:"from,omitempty"`
	To     string `json:"to,omitempty" bson:"to,omitempty"`
	Tokens string `json:"tokens,omitempty" bson:"tokens,omitempty"`
	TokenOwner string `json:"tokenOwner,omitempty" bson:"tokenOwner,omitempty"`
	Spender string `json:"spender,omitempty" bson:"spender,omitempty"`
}

type BlockLog struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	StartBlock string `json:"startblock,omitempty" bson:"from,omitempty"`
}

// LogTransfer ..
type LogTransfer struct {
	From   common.Address
	To     common.Address
	Tokens *big.Int
}

// LogApproval ..
type LogApproval struct {
	TokenOwner common.Address
	Spender    common.Address
	Tokens     *big.Int
}

func main() {
	ctab := crontab.New()
	dbClient , err :=connectionhelper.GetMongoClient()
	if err != nil {
		fmt.Println("Error",err);
		return
	}
	clientInstance = dbClient
	collection := clientInstance.Database("myFirstDatabase3").Collection("BlockLog")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	var initBlockLog BlockLog
	initBlockLog.StartBlock="11955000"
	result , _ :=collection.InsertOne(ctx,initBlockLog)
	fmt.Println("initBlockLog", result)

	//init cronjob for each minute
	ctab.MustAddJob("* * * * *", subscribeERC20) 
	
	router := mux.NewRouter()
	fmt.Println("HI")
	
	router.HandleFunc("/peekIndexedDB", peekIndexedDB ).Methods("GET")

	http.ListenAndServe(":12345",router)
}

func subscribeERC20() {	
	client, err := ethclient.Dial("https://rpc-mainnet.matic.network")
	if err!=nil {
		log.Fatal(err)
	}
	collection := clientInstance.Database("myFirstDatabase3").Collection("BlockLog")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	var blocklog BlockLog;
	if err = collection.FindOne(ctx, bson.M{}).Decode(&blocklog); err != nil {
		log.Fatal(err)
	}
	
	startblock, _ := strconv.Atoi(blocklog.StartBlock)
	toblock, err := client.HeaderByNumber(context.Background(), nil)
    if err != nil {
        log.Fatal(err)
    }

	toblockInt := toblock.Number.Int64()
	contractAddress := common.HexToAddress("0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174")
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(startblock)),
		ToBlock:   big.NewInt(toblockInt),
		Addresses: []common.Address{
			contractAddress,
		},
	}
	// id,_:=primitive.ObjectIDFromHex( blocklog.ID.String() )
	// fmt.Println("IDUPDATE",id)
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": blocklog.StartBlock},
		bson.D{
			{"$set", bson.D{{"startblock", strconv.Itoa( int(toblockInt))}}},
		},
	)
	fmt.Println("update id",result)

	logs, err := client.FilterLogs(context.Background(), query)
	if err != nil {
		log.Fatal(err)
	}

	contractAbi, err := abi.JSON(strings.NewReader(string(token.TokenABI)))
	if err != nil {
		log.Fatal(err)
	}

	logTransferSig := []byte("Transfer(address,address,uint256)")
	LogApprovalSig := []byte("Approval(address,address,uint256)")
	logTransferSigHash := crypto.Keccak256Hash(logTransferSig)
	logApprovalSigHash := crypto.Keccak256Hash(LogApprovalSig)
	collection = clientInstance.Database("myFirstDatabase3").Collection("Database")
	ctx, _ = context.WithTimeout(context.Background(), 10*time.Second)
	
	for _, vLog := range logs {
		fmt.Printf("Log Block Number: %d\n", vLog.BlockNumber)
		fmt.Printf("Log Index: %d\n", vLog.Index)
		fmt.Println("TransactionHash: ",vLog.TxHash)
		switch vLog.Topics[0].Hex() {
		case logTransferSigHash.Hex():
			fmt.Printf("Log Name: Transfer\n")

			var transferEvent LogTransfer
			err := contractAbi.UnpackIntoInterface(&transferEvent, "Transfer", vLog.Data)
			val , err := contractAbi.Unpack("Transfer",vLog.Data)
			fmt.Println(val)
			if err != nil {
				log.Fatal(err)
			}

			transferEvent.From = common.HexToAddress(vLog.Topics[1].Hex())
			transferEvent.To = common.HexToAddress(vLog.Topics[2].Hex())

			var database Database
			database.From = transferEvent.From.Hex()
			database.To = transferEvent.To.Hex()
			database.Tokens = transferEvent.Tokens.String()
			result, _ := collection.InsertOne(ctx, database)

			
			fmt.Printf("From: %s\n", transferEvent.From.Hex())
			fmt.Printf("To: %s\n", transferEvent.To.Hex())
			fmt.Printf("Tokens: %s\n", transferEvent.Tokens.String())
			fmt.Println("dbID ",result)
			
		case logApprovalSigHash.Hex():
			fmt.Printf("Log Name: Approval\n")

			var approvalEvent LogApproval
			val , err := contractAbi.Unpack("Approval",vLog.Data)
			fmt.Println(val)
			contractAbi.UnpackIntoInterface(&approvalEvent, "Approval", vLog.Data)

			if err != nil {
				log.Fatal(err)
			}

			approvalEvent.TokenOwner = common.HexToAddress(vLog.Topics[1].Hex())
			approvalEvent.Spender = common.HexToAddress(vLog.Topics[2].Hex())
			var database Database
			
			database.TokenOwner = approvalEvent.TokenOwner.Hex()
			database.Spender = approvalEvent.Spender.Hex()
			database.Tokens = approvalEvent.Tokens.String()
			
			result, _ := collection.InsertOne(ctx, database)
			
			fmt.Printf("Token Owner: %s\n", approvalEvent.TokenOwner.Hex())
			fmt.Printf("Spender: %s\n", approvalEvent.Spender.Hex())
			fmt.Printf("Tokens: %s\n", approvalEvent.Tokens.String())
			fmt.Println("dbId ", result)
		}

		fmt.Printf("\n\n")
	}
}

func peekIndexedDB(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type","application/json")
	var queryResult []Database
	collection := clientInstance.Database("myFirstDatabase3").Collection("Database")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	cursor, err := collection.Find(ctx,bson.M{})
	if err!=nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"`+err.Error()+`"}`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var database Database
		cursor.Decode(&database)
		queryResult = append(queryResult, database)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message":"`+err.Error()+`"}`))
		return
	}
	fmt.Println(queryResult)
	json.NewEncoder(response).Encode(queryResult)
}


var clientInstance *mongo.Client