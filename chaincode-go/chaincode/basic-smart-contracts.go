package chaincode

import (
	"fmt"
	"log"
	"encoding/json"
	"encoding/base64"
	"strings"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/google/uuid"
)

/// common private data collection 
const org1AndOrg2PrivateCollection = "org1MSPorg2MSPPrivateCollection"

/// smart contract 
type SmartContract struct {
	contractapi.Contract
}

/// register patient function registers patient with given details and places the asset 
/// into the private data collection of the specific organization 
func (s *SmartContract) RegisterPatient(ctx contractapi.TransactionContextInterface) error {

	/// check if the client is patient 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return fmt.Errorf("Only Patient Can Register")
	}
	
	/// Take asset data from the transient map (input)
	transientMap, err := ctx.GetStub().GetTransient()
	/// check for errors 
	if err != nil {
		return fmt.Errorf("Error getting transient: %v", err)
	}

	/// access the asset data passed in the transient map 
	assetDataJSON, ok := transientMap["asset_data"]
	if !ok {
		return fmt.Errorf("Asset data not found in the transient map")
	}

	/// patient data 
	var assetData PatientInfo
	err = json.Unmarshal(assetDataJSON, &assetData)
	if err != nil {
		return fmt.Errorf("Error cannot unmarshal: %v", err)
	}

	/// add owner of patient data
	err = addOwner(ctx, &assetData)
	if err != nil {
		return fmt.Errorf("Error while inputing asset data: %v", err)
	}

	/// assign patient id 
    id, err := s.GetIdentityAttribute(ctx, "id")
	if err != nil {
		return fmt.Errorf("Cannot get id from client identity: %v", err)
	}

	err = assetData.setID(id)
	if err != nil {
		return fmt.Errorf("Error while registering patient: %v", err)
	}

	/// check input 
	err = checkValidData(assetData, 0)
	if err != nil {
		return err
	}

	/// get CollectionName
	orgCollectionName, errOrg := getOrgCollectionName(ctx)
	if errOrg != nil {
		return errOrg
	}
	
	/// check if asset already exists
	check := checkAssetAlreadyExists(ctx, orgCollectionName, assetData.ID)
	if check != nil {
		return check 
	}
		

	// verify client org and peer org 
	verify := verifyClientOrgMatchesPeerOrg(ctx)
	if verify != nil {
		return fmt.Errorf("Cannot execute the smart contract: Error %v", verify)
	}

	/// assign meta data 
	err = assetData.addMetaData(orgCollectionName)
	if err != nil {
		return fmt.Errorf("Error executing the smart contract: %v", err)
	}

	// marshal the asset data 
	assetPrivateData, err := json.Marshal(assetData)
	if err != nil {
		return fmt.Errorf("Failed to marshal asset data: %v", err)
	}

	// put the asset data into the private data collection of the org
	/// ERROR assetData 
	log.Printf("Put: collection %v, ID %v", orgCollectionName, assetData.ID)
	err = ctx.GetStub().PutPrivateData(orgCollectionName, assetData.ID, assetPrivateData)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}
	
    return nil
}

/// Function to Appointing Doctor to the patient 
func (s *SmartContract) AppointDoctor(ctx contractapi.TransactionContextInterface, id string) error {

	/// check client 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Cannot appoint doctor: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return fmt.Errorf("Cannot appoint doctor: client is not patient")
	}

	/// get client id 
	pid, err := s.GetIdentityAttribute(ctx, "id")
	if err != nil {
		return fmt.Errorf("Cannot appoint doctor: %v", err)
	}

	/// check if the patient is registered (whethere data is present in the collection or not)
	/// get patient info from the private data collection 
	patientData, err := s.ReadAssetPrivateData(ctx, pid)
	if err != nil {
		return err
	}

	doctorData, err := s.ReadDoctorPrivateData(ctx, id)
	if err != nil {
		return fmt.Errorf("Cannot appoint doctor: %v", err)
	}

	/// can also make it as array of doctor info pointers 
	err = patientData.addDoctorInfo(id)
	if err != nil {
		return err
	}

	/// update doctor PIDs 
	err = s.updateDocInfo(ctx, doctorData.ID, pid)
	if err != nil {
		return fmt.Errorf("Error while adding patient id: %v", err)
	}

	orgCollectionName, err := patientData.getMetaData()
	if err != nil {
		return err
	}

	assetPrivateData, err := json.Marshal(patientData)
	if err != nil {
		return fmt.Errorf("Failed to marshal asset data: %v", err)
	}

	// rewrite the patient data into the collection
	log.Printf("Put: collection %v, ID %v", orgCollectionName, patientData.ID)
	err = ctx.GetStub().PutPrivateData(orgCollectionName, patientData.ID, assetPrivateData)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}


	return nil
}

/// Add medical report of the existing patient 
/// Func takes in the Patient ID and the medical record 
/// and add the medical record to the patient 
func (s *SmartContract) AddMedicalRecord(ctx contractapi.TransactionContextInterface, assetID string) error {

	/// check client identity 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "doctor" {
		return fmt.Errorf("Only Doctors Can add medical reports")
	}

	/// get id of doctor
	id, err := s.GetIdentityAttribute(ctx, "id")
	if err != nil {
		return fmt.Errorf("Error getting client id: %v", err)
	}

	/// get doctor data (if doctor not registered does not work)
	docData, err := s.ReadDoctorPrivateData(ctx, id)
	if err != nil {
		return fmt.Errorf("Cannot Read doctor data: %v", err)
	}

	/// check if doctor has the patient of id
	if !docData.checkPIDExists(assetID) {
		return fmt.Errorf("Cannot Add medical reports to this patient")
	}

	/// Take asset data from the transient map (input)
	transientMap, err := ctx.GetStub().GetTransient()
	/// check for errors 
	if err != nil {
		return fmt.Errorf("Error getting transient: %v", err)
	}

	/// access the asset data passed in the transient map 
	medicalDataJSON, ok := transientMap["medical_data"]
	if !ok {
		return fmt.Errorf("medical data not found in the transient map")
	}

	/// Medical info data 
	var medicalData MedicalInfo
	err = json.Unmarshal(medicalDataJSON, &medicalData)
	if err != nil {
		return fmt.Errorf("Error cannot unmarshal: %v", err)
	}

	/// Check if the Patient is present in the private data collection of the invoked peer org
	assetData, err := s.ReadAssetPrivateData(ctx, assetID)
	if err != nil {
		return err
	}

	/// issued by 
	medicalData.SetIssuedBy(id);

	/// if Patient is present in the private data collection 
	/// then add the medical records 
	assetData.addMedicalRecord(medicalData)

	/// get name of the collection stored in 
	orgCollectionName, err := assetData.getMetaData()
	if err != nil {
		return err
	}

	assetPrivateData, err := json.Marshal(assetData)
	if err != nil {
		return fmt.Errorf("Failed to marshal asset data: %v", err)
	}

	// rewrite the patient data into the collection
	log.Printf("Put: collection %v, ID %v", orgCollectionName, assetData.ID)
	err = ctx.GetStub().PutPrivateData(orgCollectionName, assetData.ID, assetPrivateData)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}


	return nil
}

/// register Doctor info 
func (s *SmartContract) RegisterDoctor(ctx contractapi.TransactionContextInterface) error {

	/// check if the client is doctor 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return fmt.Errorf("Error getting client identity: %v", err)
	}

	if strings.ToLower(client) != "doctor" {
		return fmt.Errorf("Only Doctors Can Register")
	}

	/// Take asset data from the transient map (input)
	transientMap, err := ctx.GetStub().GetTransient()
	/// check for errors 
	if err != nil {
		return fmt.Errorf("Error getting transient: %v", err)
	}

	/// access the asset data passed in the transient map 
	assetDataJSON, ok := transientMap["asset_data"]
	if !ok {
		return fmt.Errorf("Asset data not found in the transient map")
	}

	/// doctor data 
	var assetData DoctorInfo
	err = json.Unmarshal(assetDataJSON, &assetData)
	if err != nil {
		return fmt.Errorf("Error cannot unmarshal: %v", err)
	}

	/// Get the client ID 
	clientID, err := getInvokedClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("Error Registering Doctor: %v", err)
	}

	// verify client org and peer org 
	verify := verifyClientOrgMatchesPeerOrg(ctx)
	if verify != nil {
		return fmt.Errorf("Cannot execute the smart contract: Error %v", verify)
	}

	/// assign id to doctor
	DID, err := s.GetIdentityAttribute(ctx, "id")
	if err != nil {
		return fmt.Errorf("Cannot get id from client identity: %v",err)
	}
		
	/// store the doctor data 
	var doctorData DoctorInfo
	doctorData.SetInfo(DID, assetData.PersonalInfo, assetData.Specialization, clientID, []string{})

	/// validation of the doctor data
	err = checkValidDocInfo(doctorData, 0)
	if err != nil {
		return fmt.Errorf("Error Registering Doctor: %v", err)
	}

	/// Add doctor data to private data collection 
	orgCollectionName, err := getOrgCollectionName(ctx)
	if err != nil {
		return err
	}

	/// add meta data 
	err = doctorData.addMetaData(orgCollectionName)
	if err != nil {
		return fmt.Errorf("Error executing the smart contract: %v", err)
	}

	/// check if doctor data already exists
	err = checkAssetAlreadyExists(ctx, orgCollectionName, doctorData.ID)
	if err != nil {
		return err 
	}

	// marshal the doctor data 
	doctorPrivateData, err := json.Marshal(doctorData)
	if err != nil {
		return fmt.Errorf("Failed to marshal asset data: %v", err)
	}

	// put the doctor data into the private data collection of the org
	log.Printf("Put: collection %v, ID %v", orgCollectionName, doctorData.ID)
	err = ctx.GetStub().PutPrivateData(orgCollectionName, doctorData.ID, doctorPrivateData)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}

	return nil
}

/// update doc info 
func (s *SmartContract) updateDocInfo(ctx contractapi.TransactionContextInterface, id string, pid string) error {

	/// get doctor data
	doctorData, err := s.ReadDoctorPrivateData(ctx, id)
	if err != nil {
		return fmt.Errorf("Error while updating doctor info: %v", err)
	}

	err = doctorData.AddPID(pid)
	if err != nil {
		return fmt.Errorf("Error while adding patient id: %v", err)
	}

	orgCollectionName, err := getOrgCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("Error while updating doctor info : %v", err)
	}

	// marshal the doctor data 
	doctorPrivateData, err := json.Marshal(doctorData)
	if err != nil {
		return fmt.Errorf("Failed to marshal asset data: %v", err)
	}

	// put the doctor data into the private data collection of the org
	log.Printf("Put: collection %v, ID %v", orgCollectionName, doctorData.ID)
	err = ctx.GetStub().PutPrivateData(orgCollectionName, doctorData.ID, doctorPrivateData)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}

	return nil
}

/// get patient info, client must be patient
func (s *SmartContract) GetPatientInfo(ctx contractapi.TransactionContextInterface) (*PatientInfo, error) {

	/// check client identity 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return nil, fmt.Errorf("Cannot get Patient info: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return nil, fmt.Errorf("Cannot get Patient info: client is not patient")
	}

	/// get client id 
	id, err := s.GetIdentityAttribute(ctx, "id");
	if err != nil {
		return nil, fmt.Errorf("Cannot get Patient info: %v", err)
	}

	/// get data 
	patientInfo, err := s.ReadAssetPrivateData(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Cannot get Patient info: %v", err)
	}

	return patientInfo, nil
} 

/// get doctor info of the patient client
func (s *SmartContract) GetDoctorInfo(ctx contractapi.TransactionContextInterface) (*Doctors, error) {

	/// check client identity 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return nil, fmt.Errorf("Cannot get Patient info: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return nil, fmt.Errorf("Cannot get Patient info: client is not patient")
	}

	/// get client id 
	id, err := s.GetIdentityAttribute(ctx, "id");
	if err != nil {
		return nil, fmt.Errorf("Cannot get Patient info: %v", err)
	}

	/// get data 
	patientInfo, err := s.ReadAssetPrivateData(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Cannot get Patient info: %v", err)
	}

	doctorIDs, err := patientInfo.getDocs()
	if err != nil {
		return nil, fmt.Errorf("Cannot get Doctor ids: %v", err)
	}

	doctorsData := []DoctorInfo{}
	for _, id := range doctorIDs {
		doctorData, err := s.ReadDoctorPrivateData(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("Cannot read doctor data: %v", err)
		}
		doctorsData = append(doctorsData, *doctorData)
	}

	docsInfo := &Doctors{
		Data: doctorsData,
	}

	return docsInfo, nil
} 

/// get patient medical info 
func (s *SmartContract) GetMedicalReports(ctx contractapi.TransactionContextInterface) (*MedicalRecords, error) {

	/// check client identity 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return nil, fmt.Errorf("Cannot get Patient info: %v", err)
	}

	if strings.ToLower(client) != "patient" {
		return nil, fmt.Errorf("Cannot get Patient info: client is not patient")
	}

	/// get client id 
	id, err := s.GetIdentityAttribute(ctx, "id");
	if err != nil {
		return nil, fmt.Errorf("Cannot get Patient info: %v", err)
	}

	/// get data 
	patientInfo, err := s.ReadAssetPrivateData(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Cannot get Patient info: %v", err)
	}

	medicalReports, err := patientInfo.getMedicalReports()
	if err != nil {
		return nil, fmt.Errorf("Cannot get medical info: %v", err)
	}

	medicalData := &MedicalRecords{
		Data: medicalReports,
	}
	
	return medicalData, nil
}

/// read all the patients data of a doctor 
func (s *SmartContract) ReadPatientsData(ctx contractapi.TransactionContextInterface) (*PatientsMainInfo, error) {
	/// get client identity 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return nil, fmt.Errorf("Cannot get Doctor info: %v", err)
	}

	if strings.ToLower(client) != "doctor" {
		return nil, fmt.Errorf("Cannot get Doctor info: client is not Doctor")
	}

	/// get client id 
	id, err := s.GetIdentityAttribute(ctx, "id");
	if err != nil {
		return nil, fmt.Errorf("Cannot get Doctor info: %v", err)
	}

	/// get doctor data
	doctorData, err := s.ReadDoctorPrivateData(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Doctor data not found: %v", err)
	}

	patientIDs, err := doctorData.getPatientIDs();
	if err != nil {
		return nil, fmt.Errorf("Error while reading doctor data: %v", err)
	}

	patientsData := []PatientMainInfo{}
	for _, id := range patientIDs {
		patientData, err  := s.ReadAssetPrivateData(ctx, id)
		if err != nil {
			return nil, fmt.Errorf("Error while reading patient data: %v", err)
		}
		
		patientMainData := getPatientMainInfo(*patientData);

		/// append patient data 
		patientsData = append(patientsData, patientMainData)
	}

	resultData := &PatientsMainInfo{
		Data: patientsData,
	}
	
	return resultData, nil
}

/// read specific patient data from doctor data
func (s *SmartContract) ReadPatientData(ctx contractapi.TransactionContextInterface, pid string) (*PatientMainInfo, error){

	/// get client identity 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return nil, fmt.Errorf("Cannot get Doctor info: %v", err)
	}

	if strings.ToLower(client) != "doctor" {
		return nil, fmt.Errorf("Cannot get Doctor info: client is not Doctor")
	}

	/// get client id 
	id, err := s.GetIdentityAttribute(ctx, "id");
	if err != nil {
		return nil, fmt.Errorf("Cannot get Doctor info: %v", err)
	}

	/// get doctor data
	doctorData, err := s.ReadDoctorPrivateData(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("Doctor data not found: %v", err)
	}

	/// check if patient is under doctor data 
	if !doctorData.checkPIDExists(pid) {
		return nil, fmt.Errorf("Cannot Read Patient Data of specified Patient id")
	}

	patientData, err := s.ReadAssetPrivateData(ctx, pid)
	if err != nil {
		return nil, fmt.Errorf("Cannot get Patient Data: %v", err)
	}

	patientMainData := getPatientMainInfo(*patientData);

	return &patientMainData, nil
}

/// read asset data from the private collection of the organization
func (s *SmartContract) ReadAssetPrivateData(ctx contractapi.TransactionContextInterface, assetID string) (*PatientInfo, error) {

	/// get CollectionName
	orgCollectionName, err := getOrgCollectionName(ctx)
	if err != nil {
		return nil, err
	}

	// verfiy client org
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("Cannot execute the smart contract: Error %v", err)
	}

	log.Printf("ReadAssetPrivateData: collection %v, ID %v", orgCollectionName, assetID)
	assetDataJSON, err := ctx.GetStub().GetPrivateData(orgCollectionName, assetID) //get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %v", err)
	}

	//No Asset found, return empty response
	if assetDataJSON == nil {
		log.Printf("%v does not exist in collection %v", assetID, orgCollectionName)
		/// read data from the common private data collection
		if assetData, errIn := s.ReadAssetData(ctx, assetID); errIn != nil {
			return nil, fmt.Errorf("failed to read asset %v", assetID)
		} else {
			return assetData, nil
		}
	}


	var assetData *PatientInfo
	err = json.Unmarshal(assetDataJSON, &assetData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}


	return assetData, nil
}

/// read asset data from the common private data collection 
func (s *SmartContract) ReadAssetData(ctx contractapi.TransactionContextInterface, assetID string) (*PatientInfo, error) {

	log.Printf("ReadAsset: collection %v, ID %v",  org1AndOrg2PrivateCollection, assetID)
	assetDataJSON, err := ctx.GetStub().GetPrivateData(org1AndOrg2PrivateCollection, assetID) //get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %v", err)
	}

	//No Asset found, return empty response
	if assetDataJSON == nil {
		log.Printf("%v does not exist in collection %v", assetID,  org1AndOrg2PrivateCollection)
		return nil, fmt.Errorf("asset %v not found", assetID)
	}

	var assetData *PatientInfo
	err = json.Unmarshal(assetDataJSON, &assetData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return assetData, nil

}

/// read doctor data from the private collection of the organization
func (s *SmartContract) ReadDoctorPrivateData(ctx contractapi.TransactionContextInterface, doctorID string) (*DoctorInfo, error) {

	/// get CollectionName
	orgCollectionName, err := getOrgCollectionName(ctx)
	if err != nil {
		return nil, err
	} 

	// verfiy client org
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("Cannot execute the smart contract: Error %v", err)
	}

	log.Printf("ReadAssetPrivateData: collection %v, ID %v", orgCollectionName, doctorID)
	doctorDataJSON, err := ctx.GetStub().GetPrivateData(orgCollectionName, doctorID) //get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %v", err)
	}

	//No Doctor info found, return empty response
	if doctorDataJSON == nil {
		log.Printf("%v does not exist in collection %v", doctorID, orgCollectionName)
		return nil, fmt.Errorf("%v does not exist in collection %v", doctorID, orgCollectionName)
	}

	var doctorData *DoctorInfo
	err = json.Unmarshal(doctorDataJSON, &doctorData)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}


	return doctorData, nil
}

/// query all the patient data in the private data collection of the specific org
func (s *SmartContract) GetPatientDataOrg(ctx contractapi.TransactionContextInterface) (*Patients, error) {

	/// access control 
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return nil, fmt.Errorf("Cannot get the client identity: %v", err)
	}

	if strings.ToLower(client) != "admin" {
		return nil, fmt.Errorf("Cannot execute the smart contract, only admin can")
	}

	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("Cannot execute the smart contract: Error %v", err)
	}

	/// get collection name, based on the client invoked 
	privateDataCollection, err := getOrgCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error get patient data org: %v", err)
	}

	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(privateDataCollection, "", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []PatientInfo{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset PatientInfo
		/// check the data is Patient data 
		if checkID(response.Key, "P"){
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
			}

			results = append(results, asset)
		}
	}

	patients := &Patients{
		Data: results, 
	}

	return patients, nil
}

/// query all the patient data in the shared private data collection
func (s *SmartContract) GetPatientData(ctx contractapi.TransactionContextInterface) (*Patients, error) {

	/// access control
	client, err := s.GetIdentityAttribute(ctx, "role")
	if err != nil {
		return nil, fmt.Errorf("Cannot get the client identity: %v", err)
	}

	if strings.ToLower(client) != "admin" {
		return nil, fmt.Errorf("Cannot execute the smart contract, only admin can")
	}

	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("Cannot execute the smart contract: Error %v", err)
	}

	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange("org1MSPorg2MSPPrivateCollection", "", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []PatientInfo{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset PatientInfo
		/// check the data is Patient data 
		if checkID(response.Key, "P"){
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
			}

			results = append(results, asset)
		}
	}

	patients := &Patients{
		Data: results,
	}

	return patients, nil
}

/// get doctor data from private collection of the org
func (s *SmartContract) GetDoctorDataOrg(ctx contractapi.TransactionContextInterface) (*Doctors, error) {

	err := verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return nil, fmt.Errorf("Cannot execute the smart contract: Error %v", err)
	}

	/// get collection name, based on the client invoked 
	privateDataCollection, err := getOrgCollectionName(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error get patient data org: %v", err)
	}

	/// get all the data from the private collection of the org1
	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(privateDataCollection, "", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []DoctorInfo{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset DoctorInfo
		/// check the data is Doctor data 
		if checkID(response.Key, "D"){
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
			}

			results = append(results, asset)
		}
	}

	doctors := &Doctors{
		Data: results,
	}

	return doctors, nil
}

/// get attributes from the identity cert 
func (s *SmartContract)GetIdentityAttribute(ctx contractapi.TransactionContextInterface, attr string) (string, error) {

	value, ok, err := ctx.GetClientIdentity().GetAttributeValue(attr)
	if err != nil {
		return "", fmt.Errorf("cannot get the attribute value from identity: %v", err)
	}

	if !ok {
		return "", fmt.Errorf("Cannot find the attribute specified")
	}

	return value, nil
}

/// get client identity 
func (s *SmartContract)GetInvokedClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

	ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read client ID: %v", err)
	}

	clientIDB64Decode, err := base64.StdEncoding.DecodeString(ID)
	if err != nil {
		return "", fmt.Errorf("Failed to base64 decode client ID: %v", err)
	}

	return string(clientIDB64Decode), nil
}

/// read asset data from the private collection of the organization using the name of the patient 
// func (s *SmartContract) ReadPatientDataByName(ctx contractapi.TransactionContextInterface, name string) (*PatientInfo, error) {

// 	err := verifyClientOrgMatchesPeerOrg(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("Cannot execute the smart contract: Error %v", err)
// 	}

// 	data, err := s.GetPatientDataOrg(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("Cannot read patient data: %v", err)
// 	}

// 	for _, value := range data {
// 		if value.PersonalInfo.GetFullName() == name {
// 			return value, nil
// 		}
// 	}

// 	data, err = s.GetPatientData(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("Cannot read patient data: %v", err)
// 	}

// 	for _, value := range data {
// 		if value.PersonalInfo.GetFullName() == name {
// 			return value, nil
// 		}
// 	}


// 	return nil, fmt.Errorf("Cannot find patient data")
// }

// /// read asset data from the private collection of the organization using the name of the doctor 
// func (s *SmartContract) ReadDoctorDataByName(ctx contractapi.TransactionContextInterface, name string) (*DoctorInfo, error) {

// 	err := verifyClientOrgMatchesPeerOrg(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("Cannot execute the smart contract: Error %v", err)
// 	}

// 	data, err := s.GetDoctorDataOrg(ctx)
// 	if err != nil {
// 		return nil, fmt.Errorf("Cannot read patient data: %v", err)
// 	}

// 	for _, value := range data {
// 		if value.PersonalInfo.GetFullName() == name {
// 			return value, nil
// 		}
// 	}

// 	return nil, fmt.Errorf("Cannot find Doctor data")
// }


/// util functions
/// check valid Patient Data 
func checkValidData(assetData PatientInfo, flag int) error {
	if err := assetData.validate(flag); err != nil {
		return err
	}
	return nil
} 

func checkValidDocInfo(docInfo DoctorInfo, flag int) error {
	if err := docInfo.validate(flag); err != nil {
		return err
	}
	return nil
}

func getPatientMainInfo(patientData PatientInfo) PatientMainInfo {
	var patientMainData PatientMainInfo
	patientMainData.SetInfo(patientData.ID, patientData.PersonalInfo, patientData.MedicalRecords)
	return patientMainData;
}

/// getOrgCollectionName Function, to return the private collection name of the 
/// specifi organization 
func getOrgCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID() 
	if err != nil {
		return "", fmt.Errorf("Failed to get the verfied identity of the client: %v", err)
	}
	
	/// making the org collection name from the identity of the client 
	orgCollectionName := clientMSPID + "PrivateCollection"

	return orgCollectionName, nil
}

/// check if the asset already exists in the world state 
/// function return error if asset is already present, else return nil
func checkAssetAlreadyExists(ctx contractapi.TransactionContextInterface, orgCollectionName, assetID string) error {
	asset, err := ctx.GetStub().GetPrivateData(orgCollectionName, assetID)
	if err != nil {
		return fmt.Errorf("Failed to get the asset: %v", err)
	} else if asset != nil {
		// fmt.Printf("Asset %v already exists", assetID)
		return fmt.Errorf("Asset %v already exists", assetID)
	}
	return nil
}

func checkAssetExistsInOwnerOrg(ctx contractapi.TransactionContextInterface, orgCollectionName, assetID string) error {
	assetData, err := ctx.GetStub().GetPrivateData(orgCollectionName, assetID)
	if err != nil {
		return fmt.Errorf("Failed to get the asset: %v", err)
	} 
	if assetData == nil {
		return fmt.Errorf("Asset %v does not exists in collection %v: %v", assetID, orgCollectionName, string(assetData))
	}
	
	return nil
}

/// get Client Identity of the the client invoked the smart contract 
/// function get thes invokedClientIdentity and return the identity (string)
func getInvokedClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

	ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read client ID: %v", err)
	}

	clientIDB64Decode, err := base64.StdEncoding.DecodeString(ID)
	if err != nil {
		return "", fmt.Errorf("Failed to base64 decode client ID: %v", err)
	}

	return string(clientIDB64Decode), nil
}

/// verify the client organization matches the peer organization
/// only the client from same organization can invoke the smart contract on the peers 
/// of same organization 
func verifyClientOrgMatchesPeerOrg(ctx contractapi.TransactionContextInterface) error {
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("Failed to get client MSPID: %v", err)
	}

	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the peer's MSPID: %v", err)
	}

	if clientMSPID != peerMSPID {
		return fmt.Errorf("client from org %v is not authorized to read or write private data from an org %v peer", clientMSPID, peerMSPID)
	}

	return nil
}

/// add owner to the asset data 
func addOwner(ctx contractapi.TransactionContextInterface, assetData *PatientInfo) error {

	/// get invoked client ID 
	clientID, err := getInvokedClientIdentity(ctx)
	if err != nil {
		return err
	}

	/// add the client ID to the asset data 
	assetData.addOwner(clientID)

	return nil
}

/// assign ID function 
/// generates unique ids using the UUID lib 
/// and based on the client type we assign an unique id
func assignID(clientType string) string {

	uid := uuid.New()
	var id string = uid.String() + string(clientType[0])	
	return id
}

/// check ID
func checkID(ID string, clientType string) bool {

	if string(ID[len(ID) - 1]) == clientType {
		return true
	} 
	 
	return false
}