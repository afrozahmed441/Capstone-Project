package chaincode

import (
	"fmt"
	"strings"
	"log"
	"encoding/json"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)


func (s *SmartContract) InitAccessPatient(ctx contractapi.TransactionContextInterface, id string) error {

	clientID, err := getInvokedClientIdentity(ctx)
	if err != nil {
		return err
	}

	orgCollectionName, errOrg := getOrgCollectionName(ctx)
	if errOrg != nil {
		return errOrg
	}

	asset := PatientInfo{
		Meta: MetaData{CollectionName: orgCollectionName }, ID: id, PersonalInfo: ClientPersonalInfo{FirstName: "A", LastName: "B", Age: 21, Gender: "M", Email: "AB@gmail.com", ContactNumber: "123456789", City:"VJ", State:"AP", Country:"India", Type:"Patient"}, MedicalRecords: []MedicalInfo{}, TreatedBy: []string{}, Owners: []string{clientID}, 
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutPrivateData("Org1MSPPrivateCollection", asset.ID, assetJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	
	return nil
}

func (s *SmartContract) InitPatient(ctx contractapi.TransactionContextInterface) error {

	clientID, err := getInvokedClientIdentity(ctx)
	if err != nil {
		return err
	}

	orgCollectionName, errOrg := getOrgCollectionName(ctx)
	if errOrg != nil {
		return errOrg
	}

	for i := 1; i < 1502; i++ {
		id := fmt.Sprintf("%vP", i)

		asset := PatientInfo{
			Meta: MetaData{CollectionName: orgCollectionName }, ID: id, PersonalInfo: ClientPersonalInfo{FirstName: "A", LastName: "B", Age: 21, Gender: "M", Email: "AB@gmail.com", ContactNumber: "123456789", City:"VJ", State:"AP", Country:"India", Type:"Patient"}, MedicalRecords: []MedicalInfo{}, TreatedBy: []string{}, Owners: []string{clientID}, 
		}
	
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}
	
		err = ctx.GetStub().PutPrivateData(orgCollectionName, asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}


func (s *SmartContract) InitDoctor(ctx contractapi.TransactionContextInterface) error {
	clientID, err := getInvokedClientIdentity(ctx)
	if err != nil {
		return err
	}

	orgCollectionName, errOrg := getOrgCollectionName(ctx)
	if errOrg != nil {
		return errOrg
	}

	for i := 1; i < 1502; i++ {
		id := fmt.Sprintf("%vD", i)
		asset := DoctorInfo{
			Meta: MetaData{CollectionName: orgCollectionName }, ID: id, PersonalInfo: ClientPersonalInfo{FirstName: "C", LastName: "D", Age: 34, Gender: "F", Email: "CD@gmail.com", ContactNumber: "123456789", City:"VJ", State:"AP", Country:"India", Type:"Doctor"}, Specialization: "Heart", HID: clientID, PIDS: []string{}, 
		}

		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutPrivateData(orgCollectionName, asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}
	
	return nil
}


func (s *SmartContract) InitAccessDoctor(ctx contractapi.TransactionContextInterface, id string) error {

	clientID, err := getInvokedClientIdentity(ctx)
	if err != nil {
		return err
	}

	orgCollectionName, errOrg := getOrgCollectionName(ctx)
	if errOrg != nil {
		return errOrg
	}

	asset := DoctorInfo{
		Meta: MetaData{CollectionName: orgCollectionName }, ID: id, PersonalInfo: ClientPersonalInfo{FirstName: "C", LastName: "D", Age: 34, Gender: "F", Email: "CD@gmail.com", ContactNumber: "123456789", City:"VJ", State:"AP", Country:"India", Type:"Doctor"}, Specialization: "Heart", HID: clientID, PIDS: []string{}, 
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutPrivateData("Org1MSPPrivateCollection", asset.ID, assetJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	
	return nil
}


func (s *SmartContract) InitSharePatient(ctx contractapi.TransactionContextInterface, id string) error {

	clientID, err := getInvokedClientIdentity(ctx)
	if err != nil {
		return err
	}

	orgCollectionName, errOrg := getOrgCollectionName(ctx)
	if errOrg != nil {
		return errOrg
	}

	asset := PatientInfo{
		Meta: MetaData{CollectionName: orgCollectionName }, ID: id, PersonalInfo: ClientPersonalInfo{FirstName: "A", LastName: "B", Age: 21, Gender: "M", Email: "AB@gmail.com", ContactNumber: "123456789", City:"VJ", State:"AP", Country:"India", Type:"Patient"}, MedicalRecords: []MedicalInfo{}, TreatedBy: []string{}, Owners: []string{clientID}, 
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutPrivateData("Org1MSPPrivateCollection", asset.ID, assetJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	
	return nil
}

func (s *SmartContract) InitShareDoctor(ctx contractapi.TransactionContextInterface, id string) error {

	clientID, err := getInvokedClientIdentity(ctx)
	if err != nil {
		return err
	}

	orgCollectionName, errOrg := getOrgCollectionName(ctx)
	if errOrg != nil {
		return errOrg
	}

	asset := DoctorInfo{
		Meta: MetaData{CollectionName: orgCollectionName }, ID: id, PersonalInfo: ClientPersonalInfo{FirstName: "C", LastName: "D", Age: 34, Gender: "F", Email: "CD@gmail.com", ContactNumber: "123456789", City:"VJ", State:"AP", Country:"India", Type:"Doctor"}, Specialization: "Heart", HID: clientID, PIDS: []string{}, 
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutPrivateData("Org2MSPPrivateCollection", asset.ID, assetJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}
	
	return nil
}

/// create share request agreement 
func (s *SmartContract) InitShareRequestAgreements(ctx contractapi.TransactionContextInterface) error {
	
	/// check if the client is doctor 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "doctor" {
		return fmt.Errorf("Only Doctor can create request agreement")
	}

	for i := 1; i < 1502; i++ {

		id := fmt.Sprintf("%vD", i)
		pid := fmt.Sprintf("%vP", i)

		// Create agreeement that indicates which identity that is requesting data
		requestAgreeKey, err := ctx.GetStub().CreateCompositeKey(requestAgreementObjectType, []string{pid})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}

		var requestAgreementData requestAgreement
		err = requestAgreementData.assignMetaData("org2", "Alice", id)
		if err != nil {
			return fmt.Errorf("Cannot create request agreement: %v", err)
		}

		err = requestAgreementData.assignData(pid, 
			"x509: : CN=Alice,OU=client,0-Hyperledger, ST=North Carolina, C=US: :CN=ca.org2.example.com,0=org2.example.com, L=Hursley, ST=Hampshire, C=UK", 
			"MEQCIHBH4F+6VtHOeYGS9GrIGzzMVtLa+WcVpQnPz3ArTIr/AiBTHXW3b7 jFkhG1D2kFwIIpHk98vUzR1511Hv9Me9ixjg==", 
			"MEUCIQDDJs4tWf×G9qWg2HOuzoUrSmOeUaQDhA823DgQeJZDsQIgS5Szn+GIwq1WZnf3AbMpMsWbC6GBuJilLXB2s1rYFbo=")

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

	}
	return nil

}

func (s *SmartContract) InitShareRequestAgreementsValid(ctx contractapi.TransactionContextInterface) error {
	
	/// check if the client is doctor 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "doctor" {
		return fmt.Errorf("Only Doctor can create request agreement")
	}

	for i := 1; i < 1502; i++ {
		
		id := fmt.Sprintf("%vD", i)
		pid := fmt.Sprintf("%vP", i)

		// Create agreeement that indicates which identity that is requesting data
		requestAgreeKey, err := ctx.GetStub().CreateCompositeKey(requestAgreementObjectType, []string{pid})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}

		var requestAgreementData requestAgreement
		err = requestAgreementData.assignMetaData("org2", "Alice", id)
		if err != nil {
			return fmt.Errorf("Cannot create request agreement: %v", err)
		}

		err = requestAgreementData.assignData(pid, 
			"x509: : CN=Alice,OU=client,0-Hyperledger, ST=North Carolina, C=US: :CN=ca.org2.example.com,0=org2.example.com, L=Hursley, ST=Hampshire, C=UK", 
			"MEQCIHBH4F+6VtHOeYGS9GrIGzzMVtLa+WcVpQnPz3ArTIr/AiBTHXW3b7 jFkhG1D2kFwIIpHk98vUzR1511Hv9Me9ixjg==", 
			"MEUCIQDDJs4tWf×G9qWg2HOuzoUrSmOeUaQDhA823DgQeJZDsQIgS5Szn+GIwq1WZnf3AbMpMsWbC6GBuJilLXB2s1rYFbo=")

		if err != nil {
			return fmt.Errorf("Cannot create request agreement: %v", err)
		}

		requestAgreementData.Valid = true;

		requestAgreementJSON, err := json.Marshal(requestAgreementData)
		if err != nil {
			return fmt.Errorf("Cannot marshal request agreement: %v", err)
		}

		log.Printf("createRequestAgreement Put: collection %v, ID %v, Key %v", org1AndOrg2PrivateCollection, pid, requestAgreeKey)
		err = ctx.GetStub().PutPrivateData(org1AndOrg2PrivateCollection, requestAgreeKey, requestAgreementJSON)
		if err != nil {
			return fmt.Errorf("failed to put asset bid: %v", err)
		}
	}
	return nil

}

func (s *SmartContract) InitAccessRequestAgreements(ctx contractapi.TransactionContextInterface) error {
	/// check if the client is doctor 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "doctor" {
		return fmt.Errorf("Only Doctor can create data access request")
	}

	for i := 1; i < 1502; i++ {
		id := fmt.Sprintf("%vD", i)
		pid := fmt.Sprintf("%vP", i)

		requestAccessKey, err := ctx.GetStub().CreateCompositeKey(dataAccessRequestObjectType, []string{pid})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}

		var accessRequest dataAccessRequest
		err = accessRequest.assignData(pid, "MEQCIHBH4F+6VtHOeYGS9GrIGzzMVtLa+WcVpQnPz3ArTIr/AiBTHXW3b7 jFkhG1D2kFwIIpHk98vUzR1511Hv9Me9ixjg==",  "John", "org1", id)
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

	}

	return nil

}

func (s *SmartContract) InitAccessRequestAgreementsValid(ctx contractapi.TransactionContextInterface) error {
	/// check if the client is doctor 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "doctor" {
		return fmt.Errorf("Only Doctor can create data access request")
	}

	for i := 1; i < 1502; i++ {

		id := fmt.Sprintf("%vD", i)
		pid := fmt.Sprintf("%vP", i)

		requestAccessKey, err := ctx.GetStub().CreateCompositeKey(dataAccessRequestObjectType, []string{pid})
		if err != nil {
			return fmt.Errorf("failed to create composite key: %v", err)
		}

		var accessRequest dataAccessRequest
		err = accessRequest.assignData(pid, "MEQCIHBH4F+6VtHOeYGS9GrIGzzMVtLa+WcVpQnPz3ArTIr/AiBTHXW3b7 jFkhG1D2kFwIIpHk98vUzR1511Hv9Me9ixjg==",  "John", "org1", id)
		if err != nil {
			return fmt.Errorf("Cannot create data access request: %v", err)
		}

		accessRequest.Valid = true;

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
	}
	return nil

}

/// For Testing purpose 
/// Get all the data from the private data collection 
