package gus

import (
	"math/rand"
	"time"
	"encoding/json"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
)

type Role int64
type OrgType int64

type User struct {
	Id        int64 `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
	OrgId     int64 `json:"org_id"`
	Updated   time.Time `json:"updated"`
	Created   time.Time `json:"created"`
	Role      Role `json:"role"`
	Suspended bool `json:"suspended"`
}

type UserWithClaims struct {
	*User
	*Claims
}

type Claims struct {
	Role         Role `json:"role"`
	OrgId        int64 `json:"org_id"`
	OrgSuspended bool `json:"org_suspended"`
}

type UserWithToken struct {
	User
	Token string `json:"token"`
}

func RandStringBytesMask(length int) string {
	b := make([]byte, length)
	for i := 0; i < length; {
		if idx := int(rand.Int63() & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i++
		}
	}
	return string(b)
}

// ApplyUpdates will apply updates to an 'original' struct and update fields based on an 'updates' struct
// The 'updates' struct should have point fields and should also serialize to and from json the same as the
// Intended destination fields.
func ApplyUpdates(original interface{}, updates interface{}) error {
	p, err := json.Marshal(updates)
	if err != nil {
		return err
	}
	json.Unmarshal(p, original)
	return nil
}
