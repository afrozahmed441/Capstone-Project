package main 

import (
	"fmt"
	"strings"
	"path/filepath"
	"io/ioutil"
	"os"
	"time"
	"errors"
	"strconv"
	"encoding/json"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	cjson "github.com/TylerBrock/colorjson"
	ds "github.com/afrozahmed441/Capstone-Project/application/dataStructs"
	sign "github.com/afrozahmed441/Capstone-Project/application/sign"
)

func main() {

	err := os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	if err != nil {
		fmt.Printf("Error setting DISCOVERY_AS_LOCALHOST environemnt variable: %v", err)
		return
	}

	/// take in the org to connect
	var org string 
	fmt.Printf("Enter the orgnization name to connect with: ")
	fmt.Scanf("%s", &org)

	/// get the user from the org 
	var user string
	fmt.Printf("Enter the user name to connect with: ")
	fmt.Scanf("%s", &user)

	/// get the org wallet
	wallet, err := getOrgWallet(org)
	if err != nil {
		fmt.Println("ERROR : ", err)
		os.Exit(1)
	}
	
	/// check if user identity already exits 
	if !wallet.Exists(user) {
		err = populateWallet(wallet, user, org)
		if err != nil {
			fmt.Printf("Failed to populate wallet contents: %s\n", err)
			os.Exit(1)
		}
	}

	/// based on the org select the profile connection file 
	connection, err := connect(org, wallet, user)
	if err != nil {
		fmt.Println("ERROR : ", err)
		os.Exit(1)
	}
	
	/// get channel name 
	var channelName string
	fmt.Printf("Enter the name of the channel to connect: ")
	fmt.Scanf("%s", &channelName)

	channel, err := connection.GetNetwork(channelName)
	if err != nil {
		fmt.Println("ERROR : ", err)
		os.Exit(1)
	}

	/// get chaincode name
	var chaincodeName string 
	fmt.Printf("Enter the name of the chaincode name : ")
	fmt.Scanf("%s", &chaincodeName)	

	chaincode := channel.GetContract(chaincodeName)

	for {
		/// enter smart contract to invoke 
		var smartContract string 
		fmt.Printf("\nEnter the smart contract to invoke: ")
		fmt.Scanf("%s", &smartContract)

		args  := make([]string, 10)

		switch smartContract {
			/// testing sm
			case "InitPrivateDataCollection":
				fmt.Printf("Etner the collection name: ")
				fmt.Scanf("%s", &args[0])
				res, err := submitTransaction(chaincode, smartContract, org, args...)
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}
				fmt.Printf("Result: %v\n", string(res))
			
			case "AppointDoctor":
				fmt.Printf("Enter the doctor id to appoint: ")
				fmt.Scanf("%s", &args[0])
				_, err := submitTransaction(chaincode, smartContract, org, args...)
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}
				// fmt.Println(res)
				fmt.Println("Doctor Successfully Appointed!")
			
			//// Request Agreement Smart Contracts 
			case "CreateRequestAgreement":
				fmt.Printf("Enter the patient id to request patient data: ")
				fmt.Scanf("%s", &args[0])
				_, err := createRequestAgreement(chaincode, user, org, args[0])
				if err != nil {
					fmt.Printf("ERROR: %v", err)
				}
				// fmt.Println(res)
				fmt.Println("Request Agreement Created Successfully!")
			
			case "ValidateRequestAgreement":
				/// validate the digital signatures
				valid, err := verifyDigitalSignature(chaincode, org)
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
				}
				_, err = submitTransaction(chaincode, smartContract, org, strconv.FormatBool(valid))
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}
				// fmt.Println(res)
				fmt.Println("Request Agreement Validated Successfully!")

			case "ShareAssetData":
				_, err := submitTransaction(chaincode, smartContract, org)
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}
				// fmt.Println(res)
				fmt.Println("Requested Asset Data Shared Successfully!")
			
			/// data access request smart contracts 
			/// createDataAccessRequest
			case "CreateDataAccessRequest":
				fmt.Printf("Enter the patient id to request access patient data: ")
				fmt.Scanf("%s", &args[0])
				_, err := createDataAccessRequest(chaincode, user, org, args[0])
				if err != nil {
					fmt.Printf("ERROR: %v", err)
				}
				// fmt.Println(res)
				fmt.Println("Data Access Request Created Successfully!")

			/// validate access request agreement 
			case "ValidateAccessRequestAgreement":
				/// validate the digital signatures
				valid, err := verifyDSOnAccessRequest(chaincode, org)
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
				}
				/// invoke validate data access request smart contract 
				_, err = submitTransaction(chaincode, "ValidateDataAccessRequest", org, strconv.FormatBool(valid))
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}

				fmt.Println("Data Access Request Agreement Validated Successfully!")

			/// grantDataAccess 
			case "GrantDataAccess":

				/// invoke grant data access smart contract 
		        _, err = subTransactionWithOutArgs(chaincode, smartContract, org)
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}
				// fmt.Println(res)
				fmt.Println("Data Access Granted Successfully!")

			/// revokeDataAccess
			case "RevokeAccess":
				fmt.Printf("Enter the client id to revoke access from patient data: ")
				fmt.Scanf("%s", &args[0])
				_, err := submitTransaction(chaincode, smartContract, org, args...)
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}
				// fmt.Println(res)
				fmt.Println("Data Access Revoked Successfully!")
			
			/// read patient data
			case "GetPatientInfo", "GetDoctorInfo", "GetMedicalReports":
				res, err := evuTxn(chaincode, smartContract, org)
				if err != nil {
					fmt.Printf("ERROR : %v\n", err)
				}

				result, err :=  formatJSON(res)
					if err != nil {
						fmt.Printf("ERROR: %v\n", err)
					}

				fmt.Printf("Result: %v\n", string(result))
			
			case "GetPatientDataOrg", "GetPatientData", "GetDoctorDataOrg":
				res, err := evaluateTransaction(chaincode, smartContract, org)
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}
				
				result, err :=  formatJSON(res)
					if err != nil {
						fmt.Printf("ERROR: %v\n", err)
					}

				fmt.Printf("Result: %v\n", string(result))
			
			//// read smart contracts 
			case "ReadRequestAgreement", "ReadPatientData":
				fmt.Printf("Enter the patient id: ")
				fmt.Scanf("%s", &args[0])
				res, err := evaluateTransaction(chaincode, smartContract, org, args...)
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}

				result, err :=  formatJSON(res)
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
				}

				fmt.Printf("Result: %v\n", string(result))
			
			case "NotifyRequestAgreement", "NotifyDataAccessRequest", "ReadPatientsData":
				res, err := evuTxn(chaincode, smartContract, org)
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
					return 
				}

				result, err :=  formatJSON(res)
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
				}

				fmt.Printf("Result: %v\n", string(result))
			
			case "ReadDoctorPrivateData":
				fmt.Printf("Enter the doctor id: ")
				fmt.Scanf("%s", &args[0])
				res, err := evaluateTransaction(chaincode, smartContract, org, args...)
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}
				result, err :=  formatJSON(res)
				if err != nil {
					fmt.Printf("ERROR: %v\n", err)
				}

				fmt.Printf("Result: %v\n", string(result))

			case "RegisterPatient", "RegisterDoctor":
				_, err := submitTransactionWithTransient(chaincode, smartContract, org)
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}

				// fmt.Println(res)

				if smartContract == "RegisterPatient" {
					fmt.Println("Patient Registered Successfully!")
				}

				if smartContract == "RegisterDoctor" {
					fmt.Println("Doctor Registered Successfully!")
				}
			
			case "AddMedicalRecord":
				_, err := submitTransactionWithTransient(chaincode, smartContract, org)
				if err != nil {
					fmt.Println("ERROR: ", err)
					return 
				}

				// fmt.Println(res)

				fmt.Println("Medical Record Added Successfully!")
				
			case "Exit", "exit":
				os.Exit(0)
				return
			
			default: 
				fmt.Println("Enter valid smart contract name")
		}
	}
	
	connection.Close()

	return
}

/// function takes in org, wallet, user and connect to gateway 
/// and returns gateway
func connect(org string, wallet *gateway.Wallet, user string) (*gateway.Gateway, error) {
	
	ccpPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		org + ".example.com",
		"connection-" + org +".yaml",
	)


	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, user),
		gateway.WithTimeout(3000*time.Second),
	)

	if err != nil {
		return nil, fmt.Errorf("Failed to connect to gateway: %v", err)
	}

	return gw, nil
}


/// function to submit a transaction 
func submitTransaction(chaincode *gateway.Contract, smartContractName string, org string, args ...string) ([]byte, error) {

	switch smartContractName {
		case "InitPrivateDataCollection":
			if valid := validArgs(args, 1); !valid {
				return nil, fmt.Errorf("Error: parameters not valid")
			}
		case "AppointDoctor":
			if valid := validArgs(args, 1); !valid {
				return nil, fmt.Errorf("Error: parameters not valid")
			}
		case "ShareAssetData":
			res, err := subTransactionWithOutArgs(chaincode, smartContractName, org)
			if err != nil {
				return nil, fmt.Errorf("Failed to share asset data: %v", err)
			}
			return res, nil
		case "CreateRequestAgreement":
			if valid := validArgs(args, 4); !valid {
				return nil, fmt.Errorf("Error: parameters not valid")
			}
			args = append(args, org)
		case "ValidateRequestAgreement", "ValidateDataAccessRequest":
			if valid := validArgs(args, 1); !valid {
				return nil, fmt.Errorf("Error: parameters not valid")
			}
		case "CreateDataAccessRequest":
			if valid := validArgs(args, 3); !valid {
				return nil, fmt.Errorf("Error: parameters not valid")
			}
			args = append(args, org)
		case "GrantDataAccess":
			if valid := validArgs(args, 1); !valid {
				return nil, fmt.Errorf("Error: parameters not valid")
			}
		case "RevokeAccess":
			if valid := validArgs(args, 1); !valid {
				return nil, fmt.Errorf("Error: parameters not valid")
			}
		default: 
			return nil, fmt.Errorf("Error: invalid smart contract")		
	}

	var endorsingPeer string

	if org == "org1" {
		endorsingPeer = "peer0.org1.example.com:7051"
	} 
	if org == "org2" {
		endorsingPeer = "peer0.org2.example.com:9051"
	}

	tnx, err := chaincode.CreateTransaction(
		smartContractName,
		gateway.WithEndorsingPeers(endorsingPeer),
	)
	
	if err != nil {
		return nil, fmt.Errorf("Error while creating transaction: %v", err)
	}

	res, err := tnx.Submit(args...)
	if err != nil {
		return nil, fmt.Errorf("Error while submiting transaction: %v", err)
	}	

	return res, nil
}

/// submit transaction with out any arguments
func subTransactionWithOutArgs(chaincode *gateway.Contract, smartContractName string, org string) ([]byte, error) {

	var endorsingPeer string

	if org == "org1" {
		endorsingPeer = "peer0.org1.example.com:7051"
	} 
	if org == "org2" {
		endorsingPeer = "peer0.org2.example.com:9051"
	}

	tnx, err := chaincode.CreateTransaction(
		smartContractName,
		gateway.WithEndorsingPeers(endorsingPeer),
	)
	
	if err != nil {
		return nil, fmt.Errorf("Error while creating transaction: %v", err)
	}

	res, err := tnx.Submit()
	if err != nil {
		return nil, fmt.Errorf("Error while submiting transaction: %v", err)
	}	

	return res, nil

}

/// evaluate transaction and return the result
func evaluateTransaction(chaincode *gateway.Contract, smartContractName string, org string, args ...string) ([]byte, error) {

	switch smartContractName{
		case "GetPatientDataOrg", "GetPatientData", "GetDoctorDataOrg":
			res, err := evuTxn(chaincode, smartContractName, org)
			if err != nil {
				return nil, fmt.Errorf("cannot execute smart contract: %v", err)
			}
			return res, nil
		case "ReadRequestAgreement", "ReadDataAccessRequest", "ReadPatientData":
			if valid := validArgs(args, 1); !valid {
				return nil, fmt.Errorf("Error: parameters not valid")
			}
		case "ReadDoctorPrivateData":
			if valid := validArgs(args, 1); !valid {
				return nil, fmt.Errorf("Error: parameters not valid")
			}
		case "GetIdentityAttribute":
			if valid := validArgs(args, 1); !valid {
				return nil, fmt.Errorf("Error: parameters not valid")
			}
		default: 
			return nil, fmt.Errorf("Error: invalid smart contract")	
	}

	var endorsingPeer string

	if org == "org1" {
		endorsingPeer = "peer0.org1.example.com:7051"
	} 
	if org == "org2" {
		endorsingPeer = "peer0.org2.example.com:9051"
	}

	tnx, err := chaincode.CreateTransaction(
		smartContractName,
		gateway.WithEndorsingPeers(endorsingPeer),
	)
	
	if err != nil {
		return nil, fmt.Errorf("Error while creating transaction: %v", err)
	}

	res, err := tnx.Submit(args...)
	if err != nil {
		return nil, fmt.Errorf("Error while submiting transaction: %v", err)
	}	

	return res, nil
}

/// eval transactions without any arguments 
func evuTxn(chaincode *gateway.Contract, smartContractName string, org string) ([]byte, error) {
	var endorsingPeer string

			if org == "org1" {
				endorsingPeer = "peer0.org1.example.com:7051"
			} 
			if org == "org2" {
				endorsingPeer = "peer0.org2.example.com:9051"
			}

			tnx, err := chaincode.CreateTransaction(
				smartContractName,
				gateway.WithEndorsingPeers(endorsingPeer),
			)
			
			if err != nil {
				return nil, fmt.Errorf("Error while creating transaction: %v", err)
			}
		
			res, err := tnx.Submit()
			if err != nil {
				return nil, fmt.Errorf("Error while submiting transaction: %v", err)
			}	
		
			return res, nil
}

/// submit transaction with transient data to network
func submitTransactionWithTransient(chaincode *gateway.Contract, smartContractName string, org string) ([]byte, error) {

	var id string
	var transientData map[string][]byte
	var err error

	if smartContractName == "AddMedicalRecord" {
		fmt.Printf("Enter patient id:  ")
		fmt.Scanf("%s", &id)
		transientData, err = createMedicalData(id)
	} else {
		transientData, err = getTransientData(smartContractName)
		if err != nil {
			return nil, fmt.Errorf("Error cannot get transient data: %v", err)
		}
	}

	var endorsingPeer string

	if org == "org1" {
		endorsingPeer = "peer0.org1.example.com:7051"
	} 
	if org == "org2" {
		endorsingPeer = "peer0.org2.example.com:9051"
	}

	tnx, err := chaincode.CreateTransaction(
		smartContractName,
		gateway.WithTransient(transientData),
		gateway.WithEndorsingPeers(endorsingPeer),
	)
	
	if err != nil {
		return nil, fmt.Errorf("Error while creating transaction: %v", err)
	}

	if (smartContractName == "AddMedicalRecord") {
		res, err := tnx.Submit(id)
		if err != nil {
			return nil, fmt.Errorf("Error while submiting transaction: %v", err)
		}	
		return res, nil
	}

	res, err := tnx.Submit()
	if err != nil {
		return nil, fmt.Errorf("Error while submiting transaction: %v", err)
	}	

	return res, nil
}

///functions helps create the transaction data based on the smartcontract 
func getTransientData(smartContractName string) (map[string][]byte, error) {

	switch smartContractName {
		case "RegisterPatient", "RegisterDoctor":
			return createRegisterData(smartContractName)
		default: 
			return nil, fmt.Errorf("smart contract is invalid")
	}
}


/// get Org wallet function returns the wallet for the org
func getOrgWallet(org string) (*gateway.Wallet, error) {

	walletPath := org + "/wallet"
	wallet, err := gateway.NewFileSystemWallet(walletPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to create wallet: %s\n", err)
	}

	return wallet, nil
}

/// get org msp id from the org name 
func getOrgMSPID(org string) (string, error) {
	if len(org) == 0 {
		return "", fmt.Errorf("Org name is needed")
	}
	return (strings.ToUpper(string(org[0])) + string(org[1:]) + "MSP"), nil
}

/// populateWallet function inserts the identities (client certs) into wallet 
func populateWallet(wallet *gateway.Wallet, user string, org string) error {
	credPath := filepath.Join(
		"..",
		"..",
		"fabric-samples",
		"test-network",
		"organizations",
		"peerOrganizations",
		org + ".example.com",
		"users",
		user + "@" + org + ".example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return errors.New("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	orgMSPID, err := getOrgMSPID(org)
	if err  != nil {
		return fmt.Errorf("Error while populating the wallet: %v", err)
	}

	identity := gateway.NewX509Identity(orgMSPID, string(cert), string(key))

	err = wallet.Put(user, identity)
	if err != nil {
		return err
	}

	return nil
}

/// 
func createRequestAgreement(chaincode *gateway.Contract, user, org, id string) ([]byte, error) {

	/// get client id
	clientID, err := evuTxn(chaincode, "GetInvokedClientIdentity", org,)
	if err != nil {
		return nil, fmt.Errorf("Error getting client identity: %v", err)
	}	

	idAttr, err := evaluateTransaction(chaincode, "GetIdentityAttribute", org, "id")
	if err != nil {
		return nil, fmt.Errorf("Error getting client identity attribute: %v", err)
	}

	/// request data 
	data := ds.RequestAgreement{
		MetaInfo: ds.MetaDataReq{
			Org: org,
			User: user,
			ClientID: string(idAttr),
		},
		PID: id,
		HID: string(clientID),
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("Error cannot marshal data: %v", err)
	}

	/// get wallet 
	wallet, err := getOrgWallet(org)
	if err != nil {
		return nil, fmt.Errorf("Cannot get wallet: %v", err)
	}

	/// get digital signatures 
	clientDigitalSign, err := sign.GetUserDigitalSignature(user, org, dataBytes, wallet)
	if err != nil {
		return nil, fmt.Errorf("Cannot get digital signature of invoked client: %v", err)
	}

	hospDigitalSign, err := sign.GetUserDigitalSignature("Admin", org, dataBytes, wallet)
	if err != nil {
		return nil, fmt.Errorf("Cannot get digital signature of admin: %v", err)
	}

	/// invoke the create request agreement 
	res, err := submitTransaction(chaincode, "CreateRequestAgreement", org, id, clientDigitalSign, hospDigitalSign, user)
	if err != nil {
		return nil, fmt.Errorf("Cannot invoke create request agreement smart contract: %v", err)
	}

	return res, nil
}

/// 
func verifyDigitalSignature(chaincode *gateway.Contract, org string) (bool, error) {

	idBytes, err := evaluateTransaction(chaincode, "GetIdentityAttribute", org, "id")
	if err != nil {
		return false, fmt.Errorf("Failed to verify digital signature: %v", err)
	}

	/// convert byte to string 
	id := string(idBytes)

	data, err := evaluateTransaction(chaincode, "ReadRequestAgreement", org, id)
	if err != nil {
		return false, fmt.Errorf("Cannot verify digital signature: %v", err)
	}

	var requestAgreement ds.RequestAgreementWithSign
	err = json.Unmarshal(data, &requestAgreement)
	if err != nil {
		return false, fmt.Errorf("Cannot unmarshal the request agreement data: %v", err)
	}

	/// get request data 
	var requestData ds.RequestAgreement
	requestData.SetInfo(requestAgreement.GetMetaInfo(), requestAgreement.GetPID(), requestAgreement.GetHID())

	/// get digital signatures from request agreement
	clientDSign := requestAgreement.GetClientDSign()
	hospDSign := requestAgreement.GetOrgDSign()

	/// get meta data from request agreement 
	orgReq := requestAgreement.GetMetaDataOrg()
	userReq := requestAgreement.GetMetaDataUser()

	/// convert request data into bytes 
	dataByte, err := json.Marshal(requestData)
	if err != nil {
		return false, fmt.Errorf("Cannot marshal the request agreement data: %v", err)
	}

	/// get wallet 
	wallet, err := getOrgWallet(orgReq)
	if err != nil {
		return false, fmt.Errorf("Cannot get wallet: %v", err)
	}

	/// verify the digital signatures 
	checkClientDSign, err := sign.Verify(userReq, orgReq, dataByte, clientDSign, wallet)
	if err != nil {
		return false, fmt.Errorf("Client Digital Signature failed to verify: %v", err)
	}
	checkHospDSign, err := sign.Verify("Admin", orgReq, dataByte, hospDSign, wallet)
	if err != nil {
		return false, fmt.Errorf("Organization Digital Signature failed to verify: %v", err)
	}
	valid := checkClientDSign && checkHospDSign
	
	// fmt.Printf("Client : %v", checkClientDSign)
	// fmt.Printf("Admin : %v", checkHospDSign)

	return valid, nil
}

/// 
func createDataAccessRequest(chaincode *gateway.Contract, user, org, id string) ([]byte, error) {	

	idAttr, err := evaluateTransaction(chaincode, "GetIdentityAttribute", org, "id")
	if err != nil {
		return nil, fmt.Errorf("Error getting client identity attribute: %v", err)
	}

	/// get wallet 
	wallet, err := getOrgWallet(org)
	if err != nil {
		return nil, fmt.Errorf("Cannot get wallet: %v", err)
	}

	/// data access request data 
	dataAccessReq := ds.DataAccessReq{
		MetaData: ds.MetaDataReq {
			Org: org,
			User: user,
			ClientID: string(idAttr),
		},
		PatientID: id,
	}

	dataBytes, err := json.Marshal(dataAccessReq)
	if err != nil {
		return nil, fmt.Errorf("Error cannot marshal data: %v", err)
	}

	/// get digital signatures 
	clientDigitalSign, err := sign.GetUserDigitalSignature(user, org, dataBytes, wallet)
	if err != nil {
		return nil, fmt.Errorf("Cannot get digital signature of invoked client: %v", err)
	}

	/// invoke the create request agreement 
	res, err := submitTransaction(chaincode, "CreateDataAccessRequest", org, id, clientDigitalSign, user)
	if err != nil {
		return nil, fmt.Errorf("Cannot invoke create request agreement smart contract: %v", err)
	}

	return res, nil
}

/// 
func verifyDSOnAccessRequest(chaincode *gateway.Contract, org string) (bool, error) { 

	idBytes, err := evaluateTransaction(chaincode, "GetIdentityAttribute", org, "id")
	if err != nil {
		return false, fmt.Errorf("Failed to verify digital signature: %v", err)
	}

	/// convert byte to string 
	id := string(idBytes)

	data, err := evaluateTransaction(chaincode, "ReadDataAccessRequest", org, id)
	if err != nil {
		return false, fmt.Errorf("Cannot verify digital signature: %v", err)
	}

	var accessRequest ds.DataAccessRequest
	err = json.Unmarshal(data, &accessRequest)
	if err != nil {
		return false, fmt.Errorf("Cannot unmarshal the data access request data: %v", err)
	}

	var accessReq ds.DataAccessReq
	accessReq.SetInfo(accessRequest.GetMetaInfo(), accessRequest.GetPatientID())

	/// get digital signatures from request agreement
	clientDSign := accessRequest.GetClientDSign()

	/// get meta data from request agreement 
	orgReq := accessRequest.GetMetaDataOrg()
	userReq := accessRequest.GetMetaDataUser()

	/// get wallet 
	wallet, err := getOrgWallet(orgReq)
	if err != nil {
		return false, fmt.Errorf("Cannot get wallet: %v", err)
	}

	dataBytes, err := json.Marshal(accessReq)
	if err != nil {
		return false, fmt.Errorf("Cannot marshal the request agreement data: %v", err)
	}

	/// verify the digital signatures 
	checkClientDSign, err := sign.Verify(userReq, orgReq, dataBytes, clientDSign, wallet)
	if err != nil {
		return false, fmt.Errorf("Client Digital Signature failed to verify: %v", err)
	}

	return checkClientDSign, nil
}

/// 7. should include the shared data with DS ?


func createRegisterData(smartContractName string) (map[string][]byte, error) {

	var data map[string][]byte
	var firstName, lastName, gender, email, contactNumber, city, state, country string 
	var age int

	fmt.Printf("Enter first name: ")
	fmt.Scanf("%s", &firstName)
	fmt.Printf("Enter last name: ")
	fmt.Scanf("%s", &lastName)
	fmt.Printf("Enter age: ")
	fmt.Scanf("%d", &age)
	fmt.Printf("Enter gender: ")
	fmt.Scanf("%s", &gender)
	fmt.Printf("Enter email: ")
	fmt.Scanf("%s", &email)
	fmt.Printf("Enter contact number: ")
	fmt.Scanf("%s", &contactNumber)
	fmt.Printf("Enter city: ")
	fmt.Scanf("%s", &city)
	fmt.Printf("Enter state: ")
	fmt.Scanf("%s", &state)
	fmt.Printf("Enter country: ")
	fmt.Scanf("%s", &country)
	

	switch smartContractName {
	case "RegisterPatient":
		var clientPersonaldata ds.ClientPersonalInfo
		var clientType string = "Patient"
		clientPersonaldata.SetInfo(age, firstName, lastName, gender, email, contactNumber, city, state, country, clientType)
		var patientData ds.PatientInfo 
		patientData.SetDefault(clientPersonaldata)
		assetData, err := json.Marshal(patientData)
		if err != nil {
			return nil, fmt.Errorf("Cannot marshal the data")
		}
		data = map[string][]byte{
			"asset_data" : assetData,
		}
	
	case "RegisterDoctor":
		var clientPersonaldata ds.ClientPersonalInfo
		var clientType string = "Doctor"
		var spec string
		fmt.Printf("Enter specialization: ")
		fmt.Scanf("%s", &spec)
		clientPersonaldata.SetInfo(age, firstName, lastName, gender, email, contactNumber, city, state, country, clientType)
		var doctorData ds.DoctorInfo 
		doctorData.SetDefault(clientPersonaldata, spec)
		assetData, err := json.Marshal(doctorData)
		if err != nil {
			return nil, fmt.Errorf("Cannot marshal the data")
		}
		data = map[string][]byte{
			"asset_data" : assetData,
		}
		
	}

  return data, nil
}

/// function to add medical data to the patient 
func createMedicalData(id string) (map[string][]byte, error) {
	
	var data map[string][]byte
	var medicalData ds.MedicalInfo
	var mType string 
	fmt.Printf("Enter type of medical record: ")
	fmt.Scanf("%s", &mType)
	// medical data
	medicalRecord := createMedicalDataForm(mType)
	
	// date 
	var date ds.Date 
	year, month, day := time.Now().Date()
	date.SetInfo(day, month, year)

	// owner	
	owner := id

	medicalData.SetInfo(mType, medicalRecord, date, owner, "")
	assetData, err := json.Marshal(medicalData)
	if err != nil {
		return nil, fmt.Errorf("Cannot marshal the data")
	}
	data = map[string][]byte{
		"medical_data" : assetData,
	}
  
	return data, nil
}

func createMedicalDataForm(mType string) map[string]string {

	if strings.ToUpper(mType) == "CBC" {
		var hb, wbc, rbc, platelets, mcv, mch, mchc, mpv, neutrophils, lymphocyte, eosinophils, basophils string
		fmt.Printf("Enter hb: ")
		fmt.Scanf("%s", &hb)
		fmt.Printf("Enter wbc: ")
		fmt.Scanf("%s", &wbc)
		fmt.Printf("Enter rbc: ")
		fmt.Scanf("%s", &rbc)
		fmt.Printf("Enter platelets count: ")
		fmt.Scanf("%s", &platelets)
		fmt.Printf("Enter mcv: ")
		fmt.Scanf("%s", &mcv)
		fmt.Printf("Enter mch: ")
		fmt.Scanf("%s", &mch)
		fmt.Printf("Enter mchc: ")
		fmt.Scanf("%s", &mchc)
		fmt.Printf("Enter mpv: ")
		fmt.Scanf("%s", &mpv)
		fmt.Printf("Enter neutrophils: ")
		fmt.Scanf("%s", &neutrophils)
		fmt.Printf("Enter lymphocyte: ")
		fmt.Scanf("%s", &lymphocyte)
		fmt.Printf("Enter eosinophils: ")
		fmt.Scanf("%s", &eosinophils)
		fmt.Printf("Enter basophils: ")
		fmt.Scanf("%s", &basophils)

		keys := []string{"hb", "wbc", "rbc", "platelets", "mcv", "mch", "mchc", "mpv", "neutrophils", "lymphocyte", "eosinophils", "basophils"}
		values := []string{hb, wbc, rbc, platelets, mcv, mch, mchc, mpv, neutrophils, lymphocyte, eosinophils, basophils}

		res := getMap(keys, values)

		return res
	} 

	if strings.ToUpper(mType) == "RFT" {

		var creatinine, egfr, bun, na, k, cl, bicarb string 
		fmt.Printf("Enter Creatinine: ")
		fmt.Scanf("%s", &creatinine)
		fmt.Printf("Enter Egfr: ")
		fmt.Scanf("%s", &egfr)
		fmt.Printf("Enter bun: ")
		fmt.Scanf("%s", &bun)
		fmt.Printf("Enter na: ")
		fmt.Scanf("%s", &na)
		fmt.Printf("Enter k: ")
		fmt.Scanf("%s", &k)
		fmt.Printf("Enter cl: ")
		fmt.Scanf("%s", &cl)
		fmt.Printf("Enter bicarb: ")
		fmt.Scanf("%s", &bicarb)

		keys := []string{"creatinine", "egfr", "bun", "na", "k", "cl", "bicarb"}
		values := []string{creatinine, egfr, bun, na, k, cl, bicarb}

		res := getMap(keys, values)

		return res
	}

	return nil
}

/// util functions
func formatJSON(data []byte) ([]byte, error) {
	var jsonData map[string]interface{}
	json.Unmarshal(data, &jsonData)

	f := cjson.NewFormatter()
	f.Indent = 2

	prettyJSON, err := f.Marshal(jsonData)
	if err != nil {
		return nil, fmt.Errorf("Error while marshaling: %v", err)
	}

	return prettyJSON, nil
}

func getMap(keys []string, values []string) map[string]string {

	res := map[string]string{}
	i := 0
	for _, value := range keys {
		res[value] =  values[i]
		i++
	}

	return res
}

func validArgs(args []string, n int) bool {
	i := 0
	for _, value := range args {
		if len(value) != 0 {
			i++
		}
	}

	return (i == n)
}
