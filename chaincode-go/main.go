package main

import (
   "log"

   "github.com/hyperledger/fabric-contract-api-go/contractapi"
   "github.com/afrozahmed441/Capstone-Project/chaincode-go/chaincode"
)

func main() {
   assetChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-request-private-data chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-request-private-data chaincode: %v", err)
	}
}
