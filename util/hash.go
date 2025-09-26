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

type SecureCredential struct {
	password string
	hash     string
}

func (s *SecureCredential) GetHash() ([]byte, error) {
	hash, err := hex.DecodeString(s.hash)
	if err != nil {
		fmt.Printf("Bad hash format!: %v\n", err)
		return nil, err
	}
	return hash, err

}
