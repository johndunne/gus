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
		ListArgs: ListArgs{Size: 3},
	})
	assert.Equal(t, 3, len(users))

	// 2nd Page shorter than size
	users, err = us.List(ListUsersParams{
		ListArgs: ListArgs{Size: 4, Page: 1, OrderBy: "id", Direction: DirectionAsc},
	})
	ErrIf(t, err)
	assert.Equal(t, 2, len(users))

	// Order by id desc
	users, err = us.List(ListUsersParams{
		ListArgs: ListArgs{Size: 20, Page: 0, OrderBy: "id", Direction: DirectionDesc},
	})
	ErrIf(t, err)
	assert.Equal(t, int64(6), users[0].Id)

	// Order by id asc
	users, err = us.List(ListUsersParams{
		ListArgs: ListArgs{Size: 20, Page: 0, OrderBy: "id", Direction: DirectionAsc},
	})
	ErrIf(t, err)
	assert.Equal(t, int64(1), users[0].Id)
}

func TestUsers_Update(t *testing.T) {
	cp.Email = "update@mail.com"
	u, _, err := us.Create(cp)
	ErrIf(t, err)
	email := "donkey@kong.com"
	fname := "Donkey"
	phone := "0345345"
	up := UpdateUserParams{Id: &u.Id, Email: &email, FirstName: &fname, Phone: &phone}
	err = us.Update(up)
	ErrIf(t, err)
	u, _ = us.Get(u.Id)
	assert.Equal(t, *up.Email, u.Email)
	assert.Equal(t, *up.FirstName, u.FirstName)
	assert.Equal(t, *up.Phone, u.Phone)
	// untouched
	assert.Equal(t, cp.LastName, u.LastName)

	// Should not update to existing email
	u, _, err = us.Create(cp)
	up.Id = &u.Id
	err = us.Update(up)
	assert.Error(t, err)

	// Should not allow update of non-existing record
	id := int64(33453453)
	up.Id = &id
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

func TestUsers_Suspend(t *testing.T) {
	cp.Email = "suspend@mail.com"
	u, tempPassword, err := us.Create(cp)
	id := u.Id
	ErrIf(t, err)
	_, err = us.Authenticate(SignInParams{Email: cp.Email, Password: tempPassword})
	ErrIf(t, err)
	err = us.Suspend(id)
	_, err = us.Authenticate(SignInParams{Email: cp.Email, Password: tempPassword})
	assert.Error(t, err)
	err = us.Restore(id)
	ErrIf(t, err)
	_, err = us.Authenticate(SignInParams{Email: cp.Email, Password: tempPassword})
	ErrIf(t, err)
	ErrIf(t, orgsv.Suspend(cp.OrgId))
	_, err = us.Authenticate(SignInParams{Email: cp.Email, Password: tempPassword})
	assert.Error(t, err)
}
