package gus

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"fmt"
)

var cp = CreateUserParams{Email: "user@mail.com", OrgId: 1}

func TestUsers_Create(t *testing.T) {
	u, _, err := us.Create(cp)
	ErrIf(t, err)
	assert.Equal(t, u.Email, cp.Email)
	assert.True(t, u.Id > 0)

	// Should not allow create for existing email
	_, _, err = us.Create(cp)
	assert.Error(t, err)

	// Get
	u, err = us.Get(u.Id)
	ErrIf(t, err)
	assert.Equal(t, u.Email, cp.Email)

	// List
	users, err := us.List(ListUsersParams{OrgId:1})
	ErrIf(t, err)
	assert.Equal(t, 1, len(users))
	assert.Equal(t, int64(1), users[0].Id)
	assert.Equal(t, cp.Email, users[0].Email)

	i := 5
	for i > 0 {
		u, _, err = us.Create(CreateUserParams{Email:fmt.Sprintf("%d@mail.com", i) })
		ErrIf(t, err)
		i--
	}
	users, err = us.List(ListUsersParams{
		ListParams: ListParams{Size:3},
	})
	assert.Equal(t, 3, len(users))
	users, err = us.List(ListUsersParams{
		ListParams: ListParams{Size:2, Page:1},
	})
	ErrIf(t, err)
	assert.Equal(t, 2, len(users))
}

func TestUsers_Update(t *testing.T) {
	cp.Email = "update@mail.com"
	u, _, err := us.Create(cp)
	ErrIf(t, err)
	up := UpdateUserParams{Id: u.Id, Email: "donkey@kong.com", FirstName: "Donkey", LastName: "Kong", Phone: "0345345"}
	err = us.Update(up)
	ErrIf(t, err)
	u, _ = us.Get(u.Id)
	assert.Equal(t, up.Email, u.Email)
	assert.Equal(t, up.FirstName, u.FirstName)
	assert.Equal(t, up.LastName, u.LastName)
	assert.Equal(t, up.Phone, u.Phone)

	// Should not update to existing email
	u, _, err = us.Create(cp)
	up.Id = u.Id
	err = us.Update(up)
	assert.Error(t, err)

	// Should not allow update of non-existing record
	up.Id = 33453453
	err = us.Update(up)
	assert.Error(t, err)
}

func TestUsers_Delete(t *testing.T) {
	cp.Email = "delete@mail.com"
	u, _, err := us.Create(cp)
	ErrIf(t, err)
	err = us.Delete(u.Id)
	ErrIf(t, err)
	u, err = us.Get(u.Id)
	assert.Nil(t, u)
	assert.Error(t, err)
}
