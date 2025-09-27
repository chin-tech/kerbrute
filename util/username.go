package util

import (
	"errors"
	"strings"
)

func FormatUsername(username string) (user string, err error) {
	if username == "" {
		return "", errors.New("Bad username: blank")
	}
	parts := strings.Split(username, "@")
	if len(parts) > 2 {
		return "", errors.New("Bad username: too many @ signs")
	}
	return parts[0], nil
}

func FormatComboLine(combo string, isHash bool) (Combo, error) {
	parts := strings.SplitN(combo, ":", 2)
	var emptyData = Combo{"", SecureCredential{}}
	if len(parts) == 0 {
		err := errors.New("Bad format - missing ':'")
		return emptyData, err
	}
	user, err := FormatUsername(parts[0])
	if err != nil {
		return emptyData, err
	}
	pass := strings.Join(parts[1:], "")
	if pass == "" {
		err = errors.New("Password is blank")
		return emptyData, err
	}
	if isHash {
		return Combo{user, SecureCredential{CredHash, pass}}, err
	}

	return Combo{user, SecureCredential{CredPassword, pass}}, err

}
