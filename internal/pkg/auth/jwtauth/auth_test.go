package jwtauth

import (
	"context"
	"github.com/he2dou/go-admin/internal/pkg/auth/jwtauth/store/buntdb"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	store, err := buntdb.NewStore(":memory:")
	assert.Nil(t, err)

	jwtAuth := New(store)

	defer jwtAuth.Release()

	ctx := context.Background()
	userID := "test"
	token, err := jwtAuth.GenerateToken(ctx, userID)
	assert.Nil(t, err)
	assert.NotNil(t, token)

	id, err := jwtAuth.ParseUserID(ctx, token.GetAccessToken())
	assert.Nil(t, err)
	assert.Equal(t, userID, id)

	err = jwtAuth.DestroyToken(ctx, token.GetAccessToken())
	assert.Nil(t, err)

	id, err = jwtAuth.ParseUserID(ctx, token.GetAccessToken())
	assert.NotNil(t, err)
	assert.EqualError(t, err, "invalid token")
	assert.Empty(t, id)
}
