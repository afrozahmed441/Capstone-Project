package chaincode

import (
	"fmt"
	"log"
	"strings"
	"strconv"
	"encoding/json"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type dataAccessRequest struct {
	MetaData metaData `json:"meta_data"`
	PatientID string `json:"patient_id"`
	ClientSign string `json:"client_sign"`
	Valid bool `json:"valid"`
}

func (dar *dataAccessRequest)assignData(pid, clientSign, user, org, id string) error {
	if len(pid) == 0 || len(clientSign) == 0 {
		return fmt.Errorf("Patient id and client digital signature are required")
	}

	if len(org) == 0 || len(user) == 0 || len(id) == 0 {
		return fmt.Errorf("Meta data is required")
	}

	dar.PatientID = pid 
	dar.ClientSign = clientSign

	dar.MetaData.Org = org 
	dar.MetaData.User = user 
	dar.MetaData.ClientID = id

	return nil
}

func (dar *dataAccessRequest) getClientID() (string, error) {
	if len(dar.MetaData.ClientID) == 0 {
		return "", fmt.Errorf("Client Id not found in meta data")
	}

	return dar.MetaData.ClientID, nil
}

const dataAccessRequestObjectType = "dataAccessRequest"

/// create data access request 
func (s *SmartContract) CreateDataAccessRequest(ctx contractapi.TransactionContextInterface, pid string, clientSign string, user string, org string) error {
	
	/// check if the client is doctor 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "doctor" {
		return fmt.Errorf("Only Doctor can create data access request")
	}

	/// get client id from identity 
	id, err := s.GetIdentityAttribute(ctx, "id")
	if err != nil {
		return fmt.Errorf("Error getting client id: %v", err)
	}

	/// verify client org and peer org
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("Create data access request cannot be performed: Error %v", err)
	}

	/// check client already has access to asset data 
	doctorData, err := s.ReadDoctorPrivateData(ctx, id)
	if err != nil {
		return fmt.Errorf("Error reading request client data: %v", err)
	}

	if doctorData.checkPIDExists(pid) {
		return fmt.Errorf("Client has already access to data")
	}

	requestAccessKey, err := ctx.GetStub().CreateCompositeKey(dataAccessRequestObjectType, []string{pid})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	var accessRequest dataAccessRequest
	err = accessRequest.assignData(pid, clientSign, user, org, id)
	if err != nil {
		return fmt.Errorf("Cannot create data access request: %v", err)
	}

	/// get org collection name 
	orgCollectionName, err := getOrgCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("Cannot create data access request: %v", err)
	}

	accessRequestJSON, err := json.Marshal(accessRequest)
	if err != nil {
		return fmt.Errorf("Cannot marshal data access request: %v", err)
	}

	log.Printf("createDataAccessRequest Put: collection %v, ID %v, Key %v", orgCollectionName, pid, requestAccessKey)
	err = ctx.GetStub().PutPrivateData(orgCollectionName, requestAccessKey, accessRequestJSON)
	if err != nil {
		return fmt.Errorf("failed to put asset bid: %v", err)
	}

	return nil

}

/// read data access request
func (s *SmartContract) ReadDataAccessRequest(ctx contractapi.TransactionContextInterface, pid string) (*dataAccessRequest, error) {
	
	// verify client org and peer org
	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("Read data access request cannot be performed: Error %v", err)
	}
	
	// composite key for dataAccessRequest of this asset
	requestAccessKey, err := ctx.GetStub().CreateCompositeKey(dataAccessRequestObjectType, []string{pid})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	/// get collection name 
	orgCollectionName, err := getOrgCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("Cannot create data access request: %v", err)
	}

	// Get the data access request from collection
	log.Printf("ReaddataAccessRequest: collection %v, ID %v", orgCollectionName, pid)
	dataAccessRequestJSON, err := ctx.GetStub().GetPrivateData(orgCollectionName, requestAccessKey) 
	if err != nil {
		return nil, fmt.Errorf("failed to read dataAccessRequest: %v", err)
	}

	/// data access request not found
	if dataAccessRequestJSON == nil {
		log.Printf("ReaddataAccessRequest for %v does not exist", pid)
		return nil, fmt.Errorf("ReaddataAccessRequest for %v does not exist", pid)
	}

	/// data access request structure 
	var request dataAccessRequest
	err = json.Unmarshal(dataAccessRequestJSON, &request)
	if err != nil {
		return nil, fmt.Errorf("Cannot unmarshal data access request: %v", err)
	}


	return &request, nil
}

func (s *SmartContract) NotifyDataAccessRequest(ctx contractapi.TransactionContextInterface) (*dataAccessRequest, error) {
	
	/// check if the client is patient 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return nil, fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return nil, fmt.Errorf("Only patient can check for data access request")
	}

	/// get client id from identity 
	id, err := s.GetIdentityAttribute(ctx, "id")
	if err != nil {
		return nil, fmt.Errorf("Error getting client id: %v", err)
	}

	// verify client org and peer org
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("Read data access request cannot be performed: Error %v", err)
	}
	
	// composite key for dataAccessRequest of this asset
	requestAccessKey, err := ctx.GetStub().CreateCompositeKey(dataAccessRequestObjectType, []string{id})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	/// get collection name 
	orgCollectionName, err := getOrgCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("Cannot create data access request: %v", err)
	}

	// Get the data access request from collection
	log.Printf("ReaddataAccessRequest: collection %v, ID %v", orgCollectionName, id)
	dataAccessRequestJSON, err := ctx.GetStub().GetPrivateData(orgCollectionName, requestAccessKey) 
	if err != nil {
		return nil, fmt.Errorf("failed to read dataAccessRequest: %v", err)
	}

	/// data access request not found
	if dataAccessRequestJSON == nil {
		log.Printf("ReaddataAccessRequest for %v does not exist", id)
		return nil, fmt.Errorf("ReaddataAccessRequest for %v does not exist", id)
	}

	/// data access request structure 
	var request dataAccessRequest
	err = json.Unmarshal(dataAccessRequestJSON, &request)
	if err != nil {
		return nil, fmt.Errorf("Cannot unmarshal data access request: %v", err)
	}


	return &request, nil
}

/// delete data access request 
func (s *SmartContract) DeleteDataAccessRequest(ctx contractapi.TransactionContextInterface, pid string) error {

	/// verify client org and peer org
	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("Read data access request cannot be performed: Error %v", err)
	}

	/// get collection name 
	orgCollectionName, err := getOrgCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("Cannot create data access request: %v", err)
	}

	// Delete the data access request from the asset collection
	requestAccessKey, err := ctx.GetStub().CreateCompositeKey(dataAccessRequestObjectType, []string{pid})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().DelPrivateData(orgCollectionName, requestAccessKey)
	if err != nil {
		return err
	}

	return nil
}

/// validate data access request (digital signature validation)
func (s *SmartContract) ValidateDataAccessRequest(ctx contractapi.TransactionContextInterface, valid string) error {

	/// only patient 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return fmt.Errorf("Only Patient Can Register")
	}

	check, err := strconv.ParseBool(valid)
	if err != nil {
		return fmt.Errorf("Cannot convert to bool: %v", err)
	}

	/// get id 
	assetID, err := s.GetIdentityAttribute(ctx, "id")
	if err != nil {
		return fmt.Errorf("Validating data access request failed: %v", err)
	}

	/// read the data access request
	request, err := s.ReadDataAccessRequest(ctx, assetID)
	if err != nil {
		return fmt.Errorf("Cannot read data access request: %v", err)
	}

	/// after verifying the digital signature on the data access request, assign the valid field 
	request.Valid = check

	/// get collection name 
	orgCollection, err := getOrgCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to validate data access request: %v", err)
	}

	/// rewrite the data access request 
	requestAccessKey, err := ctx.GetStub().CreateCompositeKey(dataAccessRequestObjectType, []string{assetID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	dataAccessRequestJSON, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("Cannot marshal data access request: %v", err)
	}

	log.Printf("ValidateDataAccessRequest Put: collection %v, ID %v, Key %v", orgCollection, assetID, requestAccessKey)
	err = ctx.GetStub().PutPrivateData(orgCollection, requestAccessKey, dataAccessRequestJSON)
	if err != nil {
		return fmt.Errorf("failed to put asset bid: %v", err)
	}

	return nil
}

/// verify data access request 
func (s *SmartContract) VerifyDataAccessRequest(ctx contractapi.TransactionContextInterface, request *dataAccessRequest) error {

	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return fmt.Errorf("Only Patient Can Register")
	}

	/// check Patient id is valid in the request 
	if len(request.PatientID) == 0 {
		return fmt.Errorf("Patient Id not found in the data access request")
	}

	/// check valid patient ID 
	if !checkID(request.PatientID, "P") {
		return fmt.Errorf("Patient ID is not valid")
	}

	/// check if the owner is initating the verification data access request  
	clientID, err := getInvokedClientIdentity(ctx)
	if err != nil {
		return err
	}

	assetData, err := s.ReadAssetPrivateData(ctx, request.PatientID)
	if err != nil {
		return err
	}

	err = assetData.checkOwner(clientID)
	if err != nil {
		return err
	}

    if !request.Valid {
		return fmt.Errorf("data access request is not valid request, digital signature falied to verify")
	}
	
	return nil

}

/// grant request access to patient data 
func (s *SmartContract) GrantDataAccess(ctx contractapi.TransactionContextInterface) error {

	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return fmt.Errorf("Only Patient Can Register")
	}

	/// get id 
	assetID, err := s.GetIdentityAttribute(ctx, "id")
	if err != nil {
		return fmt.Errorf("Validating data access request failed: %v", err)
	}

	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("grant data access cannot be performed: Error %v", err)
	}

	/// read data access request 
	request, err := s.ReadDataAccessRequest(ctx, assetID)
	if err != nil {
		return fmt.Errorf("Error while reading data access request: %v", err)
	}

	/// verify data access request
	err = s.VerifyDataAccessRequest(ctx, request)
	if err != nil {
		return fmt.Errorf("Error while verfiy request agreement: %v", err)
	}

	assetData, err := s.ReadAssetPrivateData(ctx, assetID)
	if err != nil {
		return fmt.Errorf("Error reading asset data from the collection: %v", err)
	}

	/// update access to asset data 
	/// so the Request client has privilages for access and modifiying the asset data
	reqClientID, err := request.getClientID() 
	if err != nil {
		return fmt.Errorf("Error reading meta data of agreement")
	}

	err = assetData.addDoctorInfo(reqClientID)
	if err != nil {
		return fmt.Errorf("Cannot add data to patient: %v", err)
	}

	orgCollectionName, err := assetData.getMetaData()
	if err != nil {
		return err
	}

	assetDataJSON, err := json.Marshal(assetData)
	if err != nil {
		return fmt.Errorf("Failed to marshal asset data: %v", err)
	}

	// rewrite the patient data into the collection
	log.Printf("Put: collection %v, ID %v", orgCollectionName, assetID)
	err = ctx.GetStub().PutPrivateData(orgCollectionName, assetID, assetDataJSON)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}

	/// update doctor info 
	err = s.updateDocInfo(ctx, reqClientID, assetID)
	if err != nil {
		return fmt.Errorf("Error while adding patient id: %v", err)
	}

	/// delete the data access request 
	/// after the granting permission 
	err = s.DeleteDataAccessRequest(ctx, assetID)
	if err != nil {
		return err
	} 

	return nil
}

/// remove access to patient data (revoke access)
func (s *SmartContract) RevokeAccess(ctx contractapi.TransactionContextInterface, clientID string) error {
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return fmt.Errorf("Only Patient Can Register")
	}

	/// get id 
	assetID, err := s.GetIdentityAttribute(ctx, "id")
	if err != nil {
		return fmt.Errorf("Revoke data access request failed: %v", err)
	}

	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("Revoke data access cannot be performed: Error %v", err)
	}

	/// get asset data 
	assetData, err := s.ReadAssetPrivateData(ctx, assetID)
	if err != nil {
		return fmt.Errorf("Cannot read client data: %v", err)
	}

	err = assetData.removeAccess(clientID)
	if err != nil {
		return fmt.Errorf("Revoking access failed: %v", err)
	}

	/// update state database
	orgCollectionName, err := assetData.getMetaData()
	if err != nil {
		return err
	}

	assetDataJSON, err := json.Marshal(assetData)
	if err != nil {
		return fmt.Errorf("Failed to marshal asset data: %v", err)
	}

	// rewrite the patient data into the collection
	log.Printf("Put: collection %v, ID %v", orgCollectionName, assetID)
	err = ctx.GetStub().PutPrivateData(orgCollectionName, assetID, assetDataJSON)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}


	/// get doctor data 
	doctorData, err := s.ReadDoctorPrivateData(ctx, clientID)
	if err != nil {
		return fmt.Errorf("Cannot read client data: %v", err)
	}

	err = doctorData.removePID(assetID)
	if err != nil {
		return fmt.Errorf("Revoking access failed: %v", err)
	}

	orgCollectionName, err = doctorData.getMetaData()
	if err != nil {
		return err
	}

	doctorDataJSON, err := json.Marshal(doctorData)
	if err != nil {
		return fmt.Errorf("Failed to marshal asset data: %v", err)
	}

	// rewrite the patient data into the collection
	log.Printf("Put: collection %v, ID %v", orgCollectionName, clientID)
	err = ctx.GetStub().PutPrivateData(orgCollectionName, clientID, doctorDataJSON)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}
 
	return nil
}