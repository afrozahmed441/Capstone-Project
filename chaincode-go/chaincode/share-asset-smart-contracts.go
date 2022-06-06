package chaincode

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"encoding/json"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const requestAgreementObjectType = "requestAgreement"

/// signatures 
type signatures struct {
	ClientSign string `json:"clientSign"`
	OrgSign string `json:"orgSign"`
}

func (s *signatures)assignSign(clientSign, orgSign string) error {
	if len(clientSign) == 0 || len(orgSign) == 0 {
		return fmt.Errorf("Signatures is not valid")
	}

	/// assign signatures
	s.ClientSign = clientSign
	s.OrgSign = orgSign
	
	return nil	
}

type metaData struct {
	Org string `json:"org"`
	User string `json:"user"`
	ClientID string `json:"id"`
}

/// request agreement 
type requestAgreement struct {
	MetaData metaData `json:"metaData"`
	PID  string `json:"pid"`
	HID  string `json:"hid"`
	DigitalSignatures signatures `json:"digitalSignatures"`
	Valid bool `json:"valid"`
}

func (ra *requestAgreement)assignData(pid, hid, clientSign, orgSign string) error {

	if len(pid) == 0 || len(hid) == 0 {
		return fmt.Errorf("request agreement data is not valid")
	}

	ra.PID = pid 
	ra.HID = hid
	
	err := ra.DigitalSignatures.assignSign(clientSign, orgSign)
	if err != nil {
		return fmt.Errorf("request agreement data is not valid: %v", err)
	}

	return nil
}

func (ra *requestAgreement)assignMetaData(org, user, clientID string) error {

	if len(org) == 0 || len(user) == 0 {
		return fmt.Errorf("request agreement data is not valid")
	}

	ra.MetaData.Org = org
	ra.MetaData.User = user
	ra.MetaData.ClientID = clientID

	return nil
}

func (ra *requestAgreement) verifyMetaData() error {
	if len(ra.MetaData.Org) == 0 || len(ra.MetaData.User) == 0 || len(ra.MetaData.ClientID) == 0 {
		return fmt.Errorf("Meta Data not assigned properly")
	}
	return nil
}

func (ra *requestAgreement) getClientID() (string, error) {
	if len(ra.MetaData.ClientID) == 0 {
		return "", fmt.Errorf("Client Id not found in meta data")
	}

	return ra.MetaData.ClientID, nil
}

/// create request agreement function 
/// creates an agreement to request the data from organization 
/// request agreement contains the patient ID and the hospital ID 
/*
  * PID patient ID
  * HID hospital ID
  * Digital Signatures 
*/

func (s *SmartContract) CreateRequestAgreement(ctx contractapi.TransactionContextInterface, pid string, docSign string, hospSign string, user string, org string) error {

	/// check if the client is doctor 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "doctor" {
		return fmt.Errorf("Only Doctor can create request agreement")
	}

	/// get client id from identity
	id, err := s.GetIdentityAttribute(ctx, "id")
	if err != nil {
		return fmt.Errorf("Error getting client id: %v", err)
	}


	/// check if there is already a request 
	_, err = s.ReadRequestAgreement(ctx, pid);
	if err == nil {
		return fmt.Errorf("Share Request Agreement for %v patient ID already exits", pid);
	}

	/// update doctor info 
	err = s.updateDocInfo(ctx, id, pid)
	if err != nil {
		return fmt.Errorf("Error while updating doctor data: %v", err)
	}

	/// check asset data already exists 
	assetData, err := s.ReadAssetPrivateData(ctx, pid)
	if assetData != nil {
		return fmt.Errorf("Data Already exists")
	}

	/// TODO access request for patient data

	// Get ID of submitting client identity
	clientID, err := getInvokedClientIdentity(ctx)
	if err != nil {
		return err
	}

	/// verify client org and peer org
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("Create Request Agreement cannot be performed: Error %v", err)
	}

	// Create agreeement that indicates which identity that is requesting data
	requestAgreeKey, err := ctx.GetStub().CreateCompositeKey(requestAgreementObjectType, []string{pid})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	var requestAgreementData requestAgreement
	err = requestAgreementData.assignMetaData(org, user, id)
	if err != nil {
		return fmt.Errorf("Cannot create request agreement: %v", err)
	}

	err = requestAgreementData.assignData(pid, clientID, docSign, hospSign)
	if err != nil {
		return fmt.Errorf("Cannot create request agreement: %v", err)
	}

	requestAgreementJSON, err := json.Marshal(requestAgreementData)
	if err != nil {
		return fmt.Errorf("Cannot marshal request agreement: %v", err)
	}

	log.Printf("createRequestAgreement Put: collection %v, ID %v, Key %v", org1AndOrg2PrivateCollection, pid, requestAgreeKey)
	err = ctx.GetStub().PutPrivateData(org1AndOrg2PrivateCollection, requestAgreeKey, requestAgreementJSON)
	if err != nil {
		return fmt.Errorf("failed to put asset bid: %v", err)
	}

	return nil

}

/// read request agreement function 
func (s *SmartContract) ReadRequestAgreement(ctx contractapi.TransactionContextInterface, assetID string) (*requestAgreement, error) {

	/// verify client org and peer org
	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("Read Request Agreement cannot be performed: Error %v", err)
	}
	
	log.Printf("ReadRequestAgreement: collection %v, ID %v", org1AndOrg2PrivateCollection, assetID)
	// composite key for requestAgreement of this asset
	requestAgreeKey, err := ctx.GetStub().CreateCompositeKey(requestAgreementObjectType, []string{assetID})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	// Get the identity from collection
	requestAgreementJSON, err := ctx.GetStub().GetPrivateData(org1AndOrg2PrivateCollection, requestAgreeKey) 
	if err != nil {
		return nil, fmt.Errorf("failed to read RequestAgreement: %v", err)
	}

	/// request agreement not found
	if requestAgreementJSON == nil {
		log.Printf("ReadRequestAgreement for %v does not exist", assetID)
		return nil, fmt.Errorf("ReadRequestAgreement for %v does not exist", assetID)
	}

	/// request agreement structure 
	var agreement requestAgreement
	err = json.Unmarshal(requestAgreementJSON, &agreement)
	if err != nil {
		return nil, fmt.Errorf("Cannot unmarshal request agreement: %v", err)
	}


	return &agreement, nil
}

/// read request agreement on the client id 
func (s *SmartContract) NotifyRequestAgreement(ctx contractapi.TransactionContextInterface) (*requestAgreement, error) {

	/// verify client org and peer org
	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("My Request Agreement smart contract cannot execute: Error %v", err)
	}
	
	/// get client type 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return nil, fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return nil, fmt.Errorf("Only patient can check if there are any request agreements.")
	}
	
	/// get client id
	id, err := s.GetIdentityAttribute(ctx, "id");
	if err != nil {
		return nil, fmt.Errorf("Cannot get client id: %v", err)
	}

	log.Printf("MyRequestAgreement: collection %v, ID %v", org1AndOrg2PrivateCollection, id)
	// composite key for requestAgreement of this asset
	requestAgreeKey, err := ctx.GetStub().CreateCompositeKey(requestAgreementObjectType, []string{id})
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key: %v", err)
	}

	// Get the identity from collection
	requestAgreementJSON, err := ctx.GetStub().GetPrivateData(org1AndOrg2PrivateCollection, requestAgreeKey) 
	if err != nil {
		return nil, fmt.Errorf("failed to read RequestAgreement: %v", err)
	}

	/// request agreement not found
	if requestAgreementJSON == nil {
		log.Printf("MyRequestAgreement for %v does not exist", id)
		return nil, fmt.Errorf("MyRequestAgreement for %v does not exist", id)
	}

	/// request agreement structure 
	var agreement requestAgreement
	err = json.Unmarshal(requestAgreementJSON, &agreement)
	if err != nil {
		return nil, fmt.Errorf("Cannot unmarshal request agreement: %v", err)
	}


	return &agreement, nil
}


/// function validates the request agreement based on the digital signatures
func (s *SmartContract) ValidateRequestAgreement(ctx contractapi.TransactionContextInterface, valid string) error {

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
		return fmt.Errorf("Validating request agreement failed: %v", err)
	}

	/// read the request agreement 
	agreement, err := s.ReadRequestAgreement(ctx, assetID)
	if err != nil {
		return fmt.Errorf("Cannot read request agreement: %v", err)
	}

	/// after verifying the digital signatures on the request agreement, assign the valid 
	/// field in the request agreement 
	agreement.Valid = check
	
	/// rewrite the request agreement
	requestAgreeKey, err := ctx.GetStub().CreateCompositeKey(requestAgreementObjectType, []string{assetID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	requestAgreementJSON, err := json.Marshal(agreement)
	if err != nil {
		return fmt.Errorf("Cannot marshal request agreement: %v", err)
	}

	log.Printf("ValidRequestAgreement Put: collection %v, ID %v, Key %v", org1AndOrg2PrivateCollection, assetID, requestAgreeKey)
	err = ctx.GetStub().PutPrivateData(org1AndOrg2PrivateCollection, requestAgreeKey, requestAgreementJSON)
	if err != nil {
		return fmt.Errorf("failed to put asset bid: %v", err)
	}

	return nil
}

/// delete request agreement 
func (s *SmartContract) DeleteRequestAgreement(ctx contractapi.TransactionContextInterface, assetID string) error {

	/// verify client org and peer org
	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("Read Request Agreement cannot be performed: Error %v", err)
	}

	// Delete the request agreement from the asset collection
	requestAgreeKey, err := ctx.GetStub().CreateCompositeKey(requestAgreementObjectType, []string{assetID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().DelPrivateData(org1AndOrg2PrivateCollection, requestAgreeKey)
	if err != nil {
		return err
	}

	return nil
}

/// share the asset data 
func (s *SmartContract) ShareAssetData(ctx contractapi.TransactionContextInterface) error {

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
		return fmt.Errorf("Validating request agreement failed: %v", err)
	}

	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("TransferAsset cannot be performed: Error %v", err)
	}

	/// read request agreement 
	agreement, err := s.ReadRequestAgreement(ctx, assetID)
	if err != nil {
		return fmt.Errorf("Error while reading request agreement: %v", err)
	}

	/// verify request agreement
	err = s.verifyRequestAgreement(ctx, agreement)
	if err != nil {
		return fmt.Errorf("Error while verfiy request agreement: %v", err)
	}

	assetData, err := s.ReadAssetPrivateData(ctx, assetID)
	if err != nil {
		return fmt.Errorf("Error reading asset data from the collection: %v", err)
	}

	/// add the ownership to the asset  
	assetData.addOwner(agreement.HID)

	/// update access to asset data 
	/// so the Request client has privilages for access and modifiying the asset data
	reqClientID, err := agreement.getClientID() 
	if err != nil {
		return fmt.Errorf("Error reading meta data of agreement")
	}
	assetData.addDoctorInfo(reqClientID)

	/// assign the meta data 
	assetData.addMetaData(org1AndOrg2PrivateCollection)

	/// write data into the common collection 
	assetJSONData, err := json.Marshal(assetData)
	if err != nil {
		return fmt.Errorf("Failed to marshal the asset data: %v", err)
	}
	
	log.Printf("Share Asset Put: collection %v, ID %v", org1AndOrg2PrivateCollection, assetID)
	err = ctx.GetStub().PutPrivateData(org1AndOrg2PrivateCollection, assetID, assetJSONData) //rewrite the asset
	if err != nil {
		return err
	}

	/// delete the data from the private collection of the organization 
	orgCollectionName, err := getOrgCollectionName(ctx)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelPrivateData(orgCollectionName, assetID)
	if err != nil {
		return err
	}

	/// delete the request agreement 
	/// after the sharing of asset data is done 
	err = s.DeleteRequestAgreement(ctx, assetID)
	if err != nil {
		return err
	} 

	return nil
}

/// verify request agreement function 
/// check : asset exists or not 
/// check : ownership of the asset 
/// check : HID in the request agreement
/// check : valid in the request agreement (which validates the digital signatures)
func (s *SmartContract) verifyRequestAgreement(ctx contractapi.TransactionContextInterface, agreement *requestAgreement) error {

	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return fmt.Errorf("Only Patient Can Register")
	}

	// / check if the asset exists 
	orgCollectionName, err := getOrgCollectionName(ctx)
	if err != nil {
		return err
	}

	err = checkAssetExistsInOwnerOrg(ctx, orgCollectionName, agreement.PID)
	if err != nil {
		return err
	}

	/// check if the owner is initating the sharing 
	clientID, err := getInvokedClientIdentity(ctx)
	if err != nil {
		return err
	}

	assetData, err := s.ReadAssetPrivateData(ctx, agreement.PID)
	if err != nil {
		return err
	}

	err = assetData.checkOwner(clientID)
	if err != nil {
		return err
	}

	/// check if the HID is valid or not 
	if agreement.HID == "" {
		return fmt.Errorf("HID not found in RequestAgreement for %v", agreement.PID)
	}

    if !agreement.Valid {
		return fmt.Errorf("Request Agreement is not valid request, digital signature falied to verify")
	}
	
	return nil
}

/// withdraw the requestagreement (cannel request agreement)

