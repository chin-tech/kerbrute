package util

import (
	"encoding/hex"
	"fmt"
	"github.com/chin-tech/gokrb5/v8/messages"
)

func ASRepToHashcat(asrep messages.ASRep) (string, error) {
	return fmt.Sprintf("$krb5asrep$%d$%s@%s:%s$%s",
		asrep.EncPart.EType,
		asrep.CName.PrincipalNameString(),
		asrep.CRealm,
		hex.EncodeToString(asrep.EncPart.Cipher[:16]),
		hex.EncodeToString(asrep.EncPart.Cipher[16:])), nil
}

type CredType int

const (
	CredUnknown CredType = iota
	CredPassword
	CredHash
)

type SecureCredential struct {
	Type CredType
	Cred string
}

func NewPassword(pw string) SecureCredential {
	return SecureCredential{CredPassword, pw}
}

func NewHash(hash string) SecureCredential {
	return SecureCredential{CredHash, hash}
}

func (s SecureCredential) isPassword() bool { return s.Type == CredPassword }
func (s SecureCredential) isHash() bool     { return s.Type == CredHash }

func (s SecureCredential) HashBytes() ([]byte, error) {
	if s.isHash() {
		hashBytes, err := hex.DecodeString(s.Cred)
		if err != nil {
			return nil, err
		}
		return hashBytes, err
	}
	return nil, nil
}

type Combo struct {
	Username string
	Cred     SecureCredential
}
