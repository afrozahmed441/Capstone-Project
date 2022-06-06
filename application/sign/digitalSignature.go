package sign

import (
	"fmt"
	"crypto/ecdsa"
	"crypto/sha256"
	"encoding/base64"
	"path/filepath"
	"io/ioutil"
	"strings"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
)


/// get wallet content of user 
func getUserWalletContent(user string, org string, wallet *gateway.Wallet) (gateway.Identity, error) {

	/// check if the user exists or not 
	if !wallet.Exists(user) {
		err := populateWallet(wallet, user, org)
		if err != nil {
			return nil, fmt.Errorf("%v user does not exists: %v", user, err)
		}		
	}

	userWalletContent, err  := wallet.Get(user)
	if err != nil {
		return nil, fmt.Errorf("Cannot get the user wallet content: %v", err)
	}

	return userWalletContent, nil
}

/// function to sign the hash (digital signature)
func getSignFunc(userWalletContent gateway.Identity) (Sign, error){

	/// get private key from the user wallet 
    userPrivateKeyPEM := []byte(userWalletContent.(*gateway.X509Identity).Key())

	/// create a private key from PEM encoded data.
	privateKey, err := PrivateKeyFromPEM(userPrivateKeyPEM)
	if err != nil {
		return nil, fmt.Errorf("Cannot get private key: %v", err)
	}
	
	sign, err := NewPrivateKeySign(privateKey)
	if err != nil {
		return nil, fmt.Errorf("Cannot create signing function: %v", err)
	}

	return sign, nil
} 

/// function returns hash of the given data
func getHash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

/// digital signature of the data
func GetUserDigitalSignature(user string, org string, data []byte, wallet *gateway.Wallet) (string, error) {

	/// get user wallet content
	userWalletContent, err := getUserWalletContent(user, org, wallet)
	if err != nil {
		return "", err
	}

	/// create a signing function using private key
	signFunc, err := getSignFunc(userWalletContent)
	if err != nil {
		return "", err
	}

	/// get the hash of the data
	hash := getHash(data)

	dSign, err := signFunc(hash)
	if err != nil {
		return "", fmt.Errorf("Error while creating digital signature: %v", err)
	}

	encodedDigSign := base64.StdEncoding.EncodeToString(dSign)

	return encodedDigSign, nil
}

/// get public key from the certificate 
func getUserPublicKey(user string, org string, wallet *gateway.Wallet) (*ecdsa.PublicKey, error) {

	userWalletContent, err := getUserWalletContent(user, org, wallet)
	if err != nil {
		return nil, fmt.Errorf("Cannot get user public key: %v", err)
	}

	userCertificatePEM := []byte(userWalletContent.(*gateway.X509Identity).Certificate())
	
	cert, err := CertificateFromPEM(userCertificatePEM)
	if err != nil {
		return nil, fmt.Errorf("Cannot get user public key: %v", err)
	}

	publicKey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("Not a ECDSA public key")
	}

	return publicKey, nil
}


/// verify the digital signature 
func verifyDigitalSignature(publicKey *ecdsa.PublicKey, data []byte, digitalSignature string) (bool, error) {

	/// decode the digital signature back to bytes
	dSignBytes, err := base64.StdEncoding.DecodeString(digitalSignature)
	if err != nil {
		return false, fmt.Errorf("Cannot decode digital signature: %v", err)
	}

	/// hash of the data 
	hash := getHash(data)
	
	/// verfiy 
	res := ecdsa.VerifyASN1(publicKey, hash[:], dSignBytes)


	return res, nil
}


func Verify(user string, org string, data []byte, digitalSignature string, wallet *gateway.Wallet) (bool, error) {

	/// get user public key
	userPublicKey, err := getUserPublicKey(user, org, wallet)
	if err != nil {
		return false, fmt.Errorf("Cannot get user public key: %v", err)
	}

	/// verify 
	match, err := verifyDigitalSignature(userPublicKey, data, digitalSignature)
	if err != nil {
		return false, fmt.Errorf("Cannot verify the digital signature: %v", err)
	}

	return match, nil
}


/// util functions 
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
		return fmt.Errorf("keystore folder should have contain one file")
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

/// get org msp id from the org name 
func getOrgMSPID(org string) (string, error) {
	if len(org) == 0 {
		return "", fmt.Errorf("Org name is needed")
	}
	return (strings.ToUpper(string(org[0])) + string(org[1:]) + "MSP"), nil
}