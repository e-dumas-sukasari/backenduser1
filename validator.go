package backenduser1

import (
	"fmt"
	pasproj "github.com/e-dumas-sukasari/webpasetobackend"
)

func IsAdmin(Tokenstr, PublicKey string) bool {
	role, err := pasproj.DecodeGetRole(PublicKey, Tokenstr)
	if err != nil {
		fmt.Println("Error : " + err.Error())
	}
	if role != "admin" {
		return false
	}
	return true
}

func IsHR(TokenStr, Publickey string) bool {
	role, err := pasproj.DecodeGetRole(Publickey, TokenStr)
	if err != nil {
		fmt.Println("Error : " + err.Error())
	}
	if role != "HR" {
		return false
	}
	return true
}