package chaincode 

import (
	"fmt"
	"strconv"
	"time"
)

type ClientPersonalInfo struct {
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Age           int    `json:"age"`
	Gender 		  string `json:"gender"`
	Email         string `json:"email"`
	ContactNumber string `json:"contactNumber"`
	City          string `json:"city"`
	State         string `json:"state"`
	Country       string `json:"country"`
	Type 		  string `json:"type"`
}

type Date struct {
	Day    int  `json:"day"`
	Month  time.Month  `json:"month"` 
	Year   int  `json:"year"`
}


type MetaData struct {
	CollectionName  string `json:"collectionName"`
}

type MedicalInfo struct {
	Type string `json:"type"`
	MReport map[string]string `json:"mReport"`
	DateOfIssue  Date  `json:"dateOfIssue"`
	Owner string `json:"owner"`
	IssuedBy string `json:"issuedBy"`
}

type DoctorInfo struct {
	Meta MetaData `json:"meta"`
	ID string    `json:"did"`
	PersonalInfo ClientPersonalInfo `json:"personalInfo"`
	Specialization string  `json:"specialization"`
	HID string	`json:"hid"`
	PIDS []string `json:"pids"`
}


type PatientInfo struct {
	Meta MetaData `json:"meta"`
	ID  string  `json:"pid"`
	PersonalInfo ClientPersonalInfo `json:"personalInfo"`
	MedicalRecords []MedicalInfo  `json:"medicalRecords"`
	TreatedBy  []string    `json:"doctorInfo"`
	Owners  []string	`json:"owners"`	
}

/*
* ClientPersonalInfo 
*/

func (cpi *ClientPersonalInfo) SetInfo(age int, firstName, lastName, gender, email, contactNumber, city, state, country, Ctype string) {
	cpi.FirstName = firstName
	cpi.LastName = lastName
	cpi.Gender = gender
	cpi.Age = age
	cpi.Email = email
	cpi.ContactNumber = contactNumber
	cpi.City = city 
	cpi.State = state 
	cpi.Country = country 
	cpi.Type = Ctype
}

func (cpi *ClientPersonalInfo) validate() error {
	cpiMap := map[string]string {
		"firstName": cpi.FirstName,
		"lastName": cpi.LastName,
		"gender": cpi.Gender,
		"age": strconv.Itoa(cpi.Age), 
		"email": cpi.Email,
		"contactNumber": cpi.ContactNumber,
		"city": cpi.City, 
		"state": cpi.State, 
		"country": cpi.Country,
		"type": cpi.Type,
	} 

	for key, value := range cpiMap {
		if len(value) == 0 {
			return fmt.Errorf("%v field must be non-empty value", key)
		}
		/// for age field
		if key == "age" {
			if ageVal, _ := strconv.Atoi(value); ageVal <= 0 {
				return fmt.Errorf("%v field value is not valid", key)
			}
		}
	}
    
	return nil

}

func (cpi *ClientPersonalInfo) GetFullName() string {
	return (cpi.FirstName + cpi.LastName)
}

/**
 * PatientInfo 
*/

func (pi *PatientInfo) SetInfo(id string, personalInfo ClientPersonalInfo, records []MedicalInfo, treatedBy []string, owners []string) {
	pi.ID = id
	pi.PersonalInfo = personalInfo
	pi.MedicalRecords = records
	pi.TreatedBy = treatedBy 
	pi.Owners = owners
}

func (pi *PatientInfo) addMetaData(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("Collection Name undefined")
	}
	/// assign the meta data 
	pi.Meta.CollectionName = name

	return nil
}

func (pi *PatientInfo) getMetaData() (string, error) {

	if len(pi.Meta.CollectionName) == 0 { 
		return "", fmt.Errorf("Meta Data not found")
	}

	return pi.Meta.CollectionName, nil
}

/// function to set id for patient structure 
func (pi *PatientInfo) setID(id string) error {
	if len(id) == 0 {
		return fmt.Errorf("ID is not found")
	}

	/// assign ID to patient 
	pi.ID = id

	return nil
}

func validateDocData(doctorIDs []string) error {
	if len(doctorIDs) == 0 {
		return fmt.Errorf("Doctor ID not found")
	}
	return nil
}

func (pi *PatientInfo) validate(flag int) error {
	if len(pi.ID) == 0 {
		return fmt.Errorf("ID field must be non-empty value")
	}
	if err := pi.PersonalInfo.validate(); err != nil {
		return err
	}  

	if len(pi.Owners) == 0 {
		return fmt.Errorf("Owner field must be non-empty value")
	}

	if flag == 1 {
		
		if err := validateMedicalRecords(pi.MedicalRecords); err != nil {
			return err
		}
		
		if err := validateDocData(pi.TreatedBy); err != nil {
			return err
		}
	}

	return nil
}

/// add owner to the asset data 
func (pi *PatientInfo) addOwner(ID string) error {
	if err := pi.checkOwner(ID); err == nil {
		return fmt.Errorf("Owner already exists")
	}

	/// add owner id to the array 
	pi.Owners = append(pi.Owners, ID)

	return nil
}

/// check if the owner id exists or not 
func (pi *PatientInfo) checkOwner(ID string) (error) {
		
		for _, value := range pi.Owners {
			if value == ID {
				return nil
			}
		}

		return fmt.Errorf("error: invoked client identity does not own asset")
}

/// Add medical record to exists patient info
func (pi *PatientInfo) addMedicalRecord(medicalRecord MedicalInfo) error {
	// check if the record already exists in the patient data
	if err := pi.checkMedicalRecordAlreadyExists(medicalRecord); err != nil {
		return err
	}

	// add medical record to the patient data 
	pi.MedicalRecords = append(pi.MedicalRecords, medicalRecord)

	return nil
}

/// check if the medical record already exists 
func (pi *PatientInfo) checkMedicalRecordAlreadyExists(medicalRecord MedicalInfo) error {

	/// checking based on the type of recrod 
	/// TODO change to based on date later 
	for _, value := range pi.MedicalRecords {
		if value.Type == medicalRecord.Type {
			return fmt.Errorf("Medical Record Already Exists in the Patient Data")
		}
	}

	return nil
}

/// adding doctor info to the patient structure 
func (pi *PatientInfo) addDoctorInfo(doctorData string) error {

	/// check if doctor info already 
	if err := pi.checkDocInfoAlreadyExists(doctorData); err != nil {
		return fmt.Errorf("Error while adding doctor info: %v", err)
	}

	/// if not add the doctor info 
	pi.TreatedBy = append(pi.TreatedBy, doctorData)

	return nil
}

func (pi *PatientInfo) checkDocInfoAlreadyExists(doctorData string) error {

	for _, value := range pi.TreatedBy {
		if value == doctorData {
			return fmt.Errorf("Doctor Info already exists")
		}
	}

	return nil
}

/// PID method 
func (pi *PatientInfo) getPID() (string, error) {
	if len(pi.ID) == 0 {
		return "", fmt.Errorf("Patient ID not assigned")
	}

	return pi.ID, nil
}

/// return Doc info method 
func (pi *PatientInfo) getDocs() ([]string, error) {
	if len(pi.TreatedBy) == 0 {
		return nil, fmt.Errorf("Doctor not appointed")
	}
	return pi.TreatedBy, nil
}

/// return medical reports method 
/// can also add custome methods
func (pi *PatientInfo) getMedicalReports() ([]MedicalInfo, error) {
	if len(pi.MedicalRecords) == 0 {
		return nil, fmt.Errorf("Medical reports not available")
	}
	return pi.MedicalRecords, nil
}

/// remove doctor data 
func (pi *PatientInfo) removeAccess(idVal string) error {

	if err := pi.checkDocInfoAlreadyExists(idVal); err == nil {
		return fmt.Errorf("Client ID not found")
	}

	var index int
	for i, id := range pi.TreatedBy {
		if id == idVal {
			index = i
			break
		}
	}

	pi.TreatedBy = append(pi.TreatedBy[:index], pi.TreatedBy[index+1:]...)

	return nil
}



/**
* DoctorINfo 
*/

func (di *DoctorInfo) SetInfo(id string, personalInfo ClientPersonalInfo, specialization string, hid string, pids []string) {
	di.ID = id
	di.PersonalInfo = personalInfo 
	di.Specialization = specialization
	di.HID = hid
	di.PIDS = pids
}

func (di *DoctorInfo) addMetaData(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("Collection Name undefined")
	}
	/// assign the meta data 
	di.Meta.CollectionName = name

	return nil
}

func (di *DoctorInfo) getMetaData() (string, error) {

	if len(di.Meta.CollectionName) == 0 { 
		return "", fmt.Errorf("Meta Data not found")
	}

	return di.Meta.CollectionName, nil
}

func validateDoctorInfo(doctorInfo []DoctorInfo) error {
	for _, value := range doctorInfo {
		if err := value.validate(0); err != nil {
			return err
		}
	}
	return nil
}

/// validation of doctor info 
func (di *DoctorInfo) validate(flag int) error {
	if len(di.ID) == 0 {
		return fmt.Errorf("ID field must be non-empty value")
	}
	if err := di.PersonalInfo.validate(); err != nil {
		return err
	}
	if len(di.Specialization) == 0 {
		return fmt.Errorf("Specialization field must be non-empty value")
	}
	if len(di.HID) == 0 {
		return fmt.Errorf("HID field must be non-empty value")
	}
	if (flag == 1) {
		if len(di.PIDS) == 0 {
			return fmt.Errorf("PIDS field must be non-empty value")
		}
	}

	return nil
}

/// update pids in doctor info 
func (di *DoctorInfo)AddPID(id string) error {

	/// check if the patient id already exists
	if check := di.checkPIDExists(id); check {
		return fmt.Errorf("Patient ID already exists in doctor info")
	}

	/// add the patient id
	di.PIDS = append(di.PIDS, id)

	return nil
}

/// check for pids
func (di *DoctorInfo)checkPIDExists(id string) bool {

	for _, value := range di.PIDS {
		if value == id {
			return true
		}
	}

	return false
}

func (di *DoctorInfo) removePID(idVal string) error {

	if !di.checkPIDExists(idVal) {
		return fmt.Errorf("ID not found")
	}

	var index int
	for i, id := range di.PIDS {
		if id == idVal {
			index = i
			break
		}
	}

	di.PIDS = append(di.PIDS[:index], di.PIDS[index+1:]...)

	return nil
}

/// get doctor id 
func (di *DoctorInfo)getID() (string, error) {
	if len(di.ID) == 0 {
		return "", fmt.Errorf("ID not assigned")
	}
	return di.ID, nil
}

/// get patient ids of doctor 
func (di *DoctorInfo)getPatientIDs() ([]string, error) {
	if len(di.PIDS) == 0 {
		return []string{}, fmt.Errorf("Patient IDs not found")
	}
	return di.PIDS, nil
}

/// get Doctor ID 
func (di *DoctorInfo) getDID() (string, error) {
	if len(di.ID) == 0 {
		return "", fmt.Errorf("Doctor ID not assigned")
	}

	return di.ID, nil
}


/** 
* MedicalInfo
*/ 

func (mr *MedicalInfo) SetInfo(reportType string, report map[string]string, dateOfIssue Date, owner string, issuedBy string) {
	mr.Type = reportType
	mr.MReport = report 
	mr.DateOfIssue = dateOfIssue
	mr.Owner = owner
	mr.IssuedBy = issuedBy
}

/// 
func (mr *MedicalInfo) SetIssuedBy(id string) {
	mr.IssuedBy = id;
}

/// validate all the medical records of the Patient 
/// validation of medical records means no empty fields 
func validateMedicalRecords(medicalRecords []MedicalInfo) error {

	for _, value := range medicalRecords {
		if err := value.validate(); err != nil {
			return err
		}
	}

	return nil
}

/// validation of the medical info 
func (mr *MedicalInfo) validate() error {
	if len(mr.Type) == 0 {
		return fmt.Errorf("Type field must be non-empty value")
	}
	
	/// validate report (no empty values)
	for key, value := range mr.MReport {
		if len(value) == 0 {
			return fmt.Errorf("%v field must be non-empty value", key)
		}
	}

	if err := mr.DateOfIssue.validate(); err != nil {
		return err
	}
	
	if len(mr.Owner) == 0 {
		return fmt.Errorf("Owner field must be non-empty value")
	}

	return nil
}


/// refactor this function (check for the medical owner and the PID matches)
func (mr *MedicalInfo) checkOwner(ID string) error {
	if mr.Owner != ID {
		return fmt.Errorf("Owner Does not match the ID in the medical record")
	}

	return nil
}


/**
* Date 
*/

func (d *Date) SetInfo(day int, month time.Month, year int) {
	d.Day = day 
	d.Month = month 
	d.Year = year
}

func (d *Date) validate() error {
	if d.Day <= 0 {
		return fmt.Errorf("Day field value is not valid")
	}
	if d.Month <= 0 {
		return fmt.Errorf("Month field value is not valid")
	}
	if d.Year <= 0 {
		return fmt.Errorf("Year field value is not valid")
	}

	return nil
}


/// util structs 
type PatientMainInfo struct {
	ID  string  `json:"pid"`
	PersonalInfo ClientPersonalInfo `json:"personalInfo"`
	MedicalRecords []MedicalInfo  `json:"medicalRecords"`
}

func (pmi *PatientMainInfo) SetInfo(id string, personalInfo ClientPersonalInfo, records []MedicalInfo) {
	pmi.ID = id
	pmi.PersonalInfo = personalInfo
	pmi.MedicalRecords = records
}

type PatientsMainInfo struct {
	Data []PatientMainInfo `json:"data"`
}

type Patients struct {
	Data []PatientInfo `json:"data"`
}

type Doctors struct {
	Data []DoctorInfo `json:"data"`
}

type MedicalRecords struct {
	Data []MedicalInfo `json:"data"`
}