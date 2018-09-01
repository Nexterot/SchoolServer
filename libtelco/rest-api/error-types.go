// error-types
package restapi

import (
	"encoding/json"

	"github.com/masyagin1998/SchoolServer/libtelco/log"
)

type marshalledErrors struct {
	logger            *log.Logger
	MalformedData     []byte
	InvalidData       []byte
	InvalidLoginData  []byte
	InvalidPage       []byte
	InvalidSystemType []byte
	InvalidToken      []byte
	InvalidDeviceInfo []byte
	WrongOldPassword  []byte
	SamePassword      []byte
}

func NewMarshalledErrors(logger *log.Logger) *marshalledErrors {
	// Malformed Data
	malformedData, err := json.Marshal(errorMnemocode{"malformed_data"})
	if err != nil {
		logger.Fatal("REST: Error occured when marshalling error mnemocode", "Error", err, "Mnemo", "malformed_data")
	}
	// Invalid data
	invalidData, err := json.Marshal(errorMnemocode{"invalid_data"})
	if err != nil {
		logger.Fatal("REST: Error occured when marshalling error mnemocode", "Error", err, "Mnemo", "invalid_data")
	}
	// Invalid Login Data
	invalidLoginData, err := json.Marshal(errorMnemocode{"invalid_login_data"})
	if err != nil {
		logger.Fatal("REST: Error occured when marshalling error mnemocode", "Error", err, "Mnemo", "invalid_login_data")
	}
	// Invalid Page on forum
	invalidPage, err := json.Marshal(errorMnemocode{"invalid_page"})
	if err != nil {
		logger.Fatal("REST: Error occured when marshalling error mnemocode", "Error", err, "Mnemo", "invalid_page")
	}
	// Invalid System type
	invalidSystemType, err := json.Marshal(errorMnemocode{"invalid_system_type (1 for ios, 2 for android)"})
	if err != nil {
		logger.Fatal("REST: Error occured when marshalling error mnemocode", "Error", err, "Mnemo", "invalid_system_type")
	}
	// Invalid Token
	invalidToken, err := json.Marshal(errorMnemocode{"empty_token"})
	if err != nil {
		logger.Fatal("REST: Error occured when marshalling error mnemocode", "Error", err, "Mnemo", "empty_token")
	}
	// Invalid Device Info
	invalidDeviceInfo, err := json.Marshal(errorMnemocode{"invalid_device_info"})
	if err != nil {
		logger.Fatal("REST: Error occured when marshalling error mnemocode", "Error", err, "Mnemo", "invalid_device_info")
	}
	// WrongOldPassword
	wrongOldPassword, err := json.Marshal(errorMnemocode{"wrong_old_password"})
	if err != nil {
		logger.Fatal("REST: Error occured when marshalling error mnemocode", "Error", err, "Mnemo", "wrong_old_password")
	}
	// SamePassword
	samePassword, err := json.Marshal(errorMnemocode{"same_password"})
	if err != nil {
		logger.Fatal("REST: Error occured when marshalling error mnemocode", "Error", err, "Mnemo", "same_password")
	}
	logger.Info("REST: Successfully marshalled error mnemocodes")
	return &marshalledErrors{
		MalformedData:     malformedData,
		InvalidData:       invalidData,
		InvalidLoginData:  invalidLoginData,
		InvalidPage:       invalidPage,
		InvalidSystemType: invalidSystemType,
		InvalidToken:      invalidToken,
		InvalidDeviceInfo: invalidDeviceInfo,
		WrongOldPassword:  wrongOldPassword,
		SamePassword:      samePassword,
	}
}

type errorMnemocode struct {
	Error string `json:"error"`
}
