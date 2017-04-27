package gus

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

var corg = CreateOrgParams{Name: "Trainers Inc."}

func TestOrgs_Create(t *testing.T) {
	u, err := orgsv.Create(corg)
	ErrIf(t, err)
	assert.Equal(t, u.Name, corg.Name)
	assert.True(t, u.Id > 0)

	// Get
	u, err = orgsv.Get(u.Id)
	ErrIf(t, err)
	assert.Equal(t, u.Name, corg.Name)

	orgs, err := orgsv.List(ListOrgsParams{})
	ErrIf(t, err)
	assert.Equal(t, 1, len(orgs))
	assert.Equal(t, corg.Name, orgs[0].Name)
}

func TestOrgs_Update(t *testing.T) {
	u, err := orgsv.Create(corg)
	ErrIf(t, err)
	name := "New Name"
	up := UpdateOrgParams{Id: &u.Id, Name: &name}
	err = orgsv.Update(up)
	ErrIf(t, err)
	u, _ = orgsv.Get(u.Id)
	assert.Equal(t, *up.Name, u.Name)

	// Should not allow update of non-existing record
	id := int64(33453453)
	up.Id = &id
	err = orgsv.Update(up)
	assert.Error(t, err)
}

func TestOrgs_Delete(t *testing.T) {
	u, err := orgsv.Create(corg)
	ErrIf(t, err)
	err = orgsv.Delete(u.Id)
	ErrIf(t, err)
	u, err = orgsv.Get(u.Id)
	assert.Nil(t, u)
	assert.Error(t, err)
}
