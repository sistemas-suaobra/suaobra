package server

import (
	"testing"

	"github.com/pocketbase/pocketbase/models"
	"github.com/pocketbase/pocketbase/models/schema"
	"github.com/pocketbase/pocketbase/tools/types"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActiveFromProperties(t *testing.T) {
	assert.True(t, activeFromProperties(nil))
	assert.True(t, activeFromProperties(map[string]any{"name": "Ana"}))
	assert.True(t, activeFromProperties(`{"name":"Ana"}`))
	assert.False(t, activeFromProperties(map[string]any{"active": false}))
	assert.False(t, activeFromProperties(`{"name":"Ana","active":false}`))
	assert.False(t, activeFromProperties(types.JsonRaw(`{"active":false}`)))
	assert.True(t, activeFromProperties(types.JsonRaw(`{"active":true}`)))
}

func testUserCollection(t *testing.T) *models.Collection {
	t.Helper()
	collection := &models.Collection{
		Name: "user",
		Type: models.CollectionTypeAuth,
	}
	collection.Schema = schema.NewSchema(
		&schema.SchemaField{Name: "manager", Type: schema.FieldTypeBool},
		&schema.SchemaField{Name: "properties", Type: schema.FieldTypeJson},
	)
	return collection
}

func TestLockUserPrivilegesOnAPIWrite_Create(t *testing.T) {
	record := models.NewRecord(testUserCollection(t))
	record.Set("manager", true)
	record.Set("properties", map[string]any{"name": "Ana", "is_admin": true})

	require.NoError(t, lockUserPrivilegesOnAPIWrite(record, nil))
	assert.False(t, record.GetBool("manager"))
	props, err := propertiesMap(record.Get("properties"))
	require.NoError(t, err)
	assert.Equal(t, "Ana", props["name"])
	assert.False(t, cast.ToBool(props["is_admin"]))
}

func TestLockUserPrivilegesOnAPIWrite_UpdatePreservesOriginal(t *testing.T) {
	collection := testUserCollection(t)

	original := models.NewRecord(collection)
	original.Set("manager", false)
	original.Set("properties", map[string]any{"name": "Ana", "is_admin": false, "phone": "11"})

	updated := models.NewRecord(collection)
	updated.Set("manager", true)
	updated.Set("properties", map[string]any{"name": "Ana", "is_admin": true, "phone": "22"})

	require.NoError(t, lockUserPrivilegesOnAPIWrite(updated, original))
	assert.False(t, updated.GetBool("manager"))
	props, err := propertiesMap(updated.Get("properties"))
	require.NoError(t, err)
	assert.Equal(t, "22", props["phone"])
	assert.False(t, cast.ToBool(props["is_admin"]))
}

func TestLockUserPrivilegesOnAPIWrite_KeepsExistingAdmin(t *testing.T) {
	collection := testUserCollection(t)

	original := models.NewRecord(collection)
	original.Set("manager", true)
	original.Set("properties", map[string]any{"is_admin": true, "photo": "a.jpg"})

	updated := models.NewRecord(collection)
	updated.Set("manager", false)
	updated.Set("properties", map[string]any{"is_admin": false, "photo": "b.jpg"})

	require.NoError(t, lockUserPrivilegesOnAPIWrite(updated, original))
	assert.True(t, updated.GetBool("manager"))
	props, err := propertiesMap(updated.Get("properties"))
	require.NoError(t, err)
	assert.Equal(t, "b.jpg", props["photo"])
	assert.True(t, cast.ToBool(props["is_admin"]))
}
