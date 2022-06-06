package dataStructs

import (
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


/**
 * PatientInfo 
*/

func (pi *PatientInfo) SetInfo(id string, personalInfo ClientPersonalInfo, records []MedicalInfo, treatedBy []string, owners []string) {
	pi.ID = id
	pi.PersonalInfo = personalInfo
	pi.MedicalRecords = records
	pi.TreatedBy = treatedBy 
	pi.Owners = owners
	pi.Meta = MetaData{CollectionName:""}
}


func (pi *PatientInfo) SetDefault(personalInfo ClientPersonalInfo) {
	pi.ID = ""
	pi.PersonalInfo = personalInfo
	pi.MedicalRecords = []MedicalInfo{}
	pi.TreatedBy = []string{}
	pi.Owners = []string{}
	pi.Meta = MetaData{CollectionName:""}
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

func (di *DoctorInfo) SetDefault(personalInfo ClientPersonalInfo, specialization string) {
	di.ID = ""
	di.PersonalInfo = personalInfo 
	di.Specialization = specialization
	di.HID = ""
	di.PIDS = []string{}
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

func (mr *MedicalInfo)SetDefault(reportType string) {
	mr.Type = reportType
	mr.MReport = map[string]string{}
	mr.DateOfIssue = Date{}
	mr.Owner = ""
	mr.IssuedBy = ""
}

func (d *Date) SetInfo(day int, month time.Month, year int) {
	d.Day = day 
	d.Month = month 
	d.Year = year
}

type signatures struct {
	ClientSign string `json:"clientSign"`
	OrgSign string `json:"orgSign"`
}

type MetaDataReq struct {
	Org string `json:"org"`
	User string `json:"user"`
	ClientID string `json:"id"`
}

/// request agreement data 
type RequestAgreement struct {
	MetaInfo MetaDataReq `json:"metaData"`
	PID  string `json:"pid"`
	HID  string `json:"hid"`
}

func (rd *RequestAgreement)SetInfo(metaInfo MetaDataReq, pid, hid string) {
	rd.PID = pid
	rd.HID = hid
	rd.MetaInfo = metaInfo
}

type RequestAgreementWithSign struct {
	MetaData MetaDataReq `json:"metaData"`
	PID  string `json:"pid"`
	HID  string `json:"hid"`
	DigitalSignatures signatures `json:"digitalSignatures"`
	Valid bool `json:"valid"`
}

func (r *RequestAgreementWithSign)GetPID() string {
	return r.PID
} 

func (r *RequestAgreementWithSign)GetHID() string {
	return r.HID
}

func (r *RequestAgreementWithSign)GetMetaInfo() MetaDataReq {
	return r.MetaData
}

func (r *RequestAgreementWithSign)GetClientDSign() string {
	return r.DigitalSignatures.ClientSign
} 

func (r *RequestAgreementWithSign)GetOrgDSign() string {
	return r.DigitalSignatures.OrgSign
} 

func (r *RequestAgreementWithSign)GetMetaDataUser() string {
	return r.MetaData.User
} 

func (r *RequestAgreementWithSign)GetMetaDataOrg() string {
	return r.MetaData.Org
} 

/// Data Access Request 
type DataAccessRequest struct {
	MetaData MetaDataReq `json:"meta_data"`
	PatientID string `json:"patient_id"`
	ClientSign string `json:"client_sign"`
	Valid bool `json:"valid"`
}

func (dar *DataAccessRequest) GetMetaInfo() MetaDataReq {
	return dar.MetaData
}

func (dar *DataAccessRequest) GetMetaDataUser() string {
	return dar.MetaData.User
}

func (dar *DataAccessRequest) GetMetaDataOrg() string {
	return dar.MetaData.Org
}

func (dar *DataAccessRequest) GetMetaDataID() string {
	return dar.MetaData.ClientID
}

func (dar *DataAccessRequest) GetClientDSign() string {
	return dar.ClientSign
}

func (dar *DataAccessRequest) GetPatientID() string {
	return dar.PatientID
}

/* data access request without digital signature */
type DataAccessReq struct {
	MetaData MetaDataReq `json:"meta_data"`
	PatientID string `json:"patient_id"`
}

func (r *DataAccessReq)SetInfo(metaInfo MetaDataReq, pid string) {
	r.MetaData = metaInfo
	r.PatientID = pid
}