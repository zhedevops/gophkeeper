package service

import (
	"context"
	"errors"
	"gophkeeper/internal/config"
	"gophkeeper/internal/mocks"
	"gophkeeper/internal/model"
	pb "gophkeeper/proto"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateUser(t *testing.T) {
	ctx := context.Background()
	login := "d51eae65"
	password := "dlf82a5xunr"
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepository(ctrl)
	cnf := config.GetConfig()
	m.EXPECT().CreateUser(ctx, login, gomock.Any()).DoAndReturn(func(ctx context.Context, username string, hash string) (int32, error) {
		require.NoError(t, CheckPassword(hash, password))
		return 1, nil
	})
	m.EXPECT().CreateUser(ctx, login, gomock.Any()).Return(int32(0), errors.New("user cannot create"))
	srv := NewService(m, cnf)

	t.Run("test ok", func(t *testing.T) {
		user, err := srv.CreateUser(ctx, login, password)
		assert.Nil(t, err)
		require.Equal(t, int32(1), user.ID)
		require.NotEmpty(t, user.AccessToken)
	})

	t.Run("test cannot create user", func(t *testing.T) {
		user, err := srv.CreateUser(ctx, login, password)
		assert.NotNil(t, err)
		assert.Equal(t, "user cannot create", err.Error())
		assert.Equal(t, model.User{}, user)
	})

	t.Run("test invalid hash", func(t *testing.T) {
		srv.hashFunc = func(_ string) (string, error) {
			return "", errors.New("hash failed")
		}
		user, err := srv.CreateUser(ctx, login, password)
		assert.NotNil(t, err)
		assert.Equal(t, "hash failed", err.Error())
		assert.Equal(t, model.User{}, user)
	})
}

func TestService_LoginUser(t *testing.T) {
	ctx := context.Background()
	login := "d51eae65"
	password := "dlf82a5xunr"
	passHash, err := HashPassword(password)
	require.NoError(t, err)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepository(ctrl)
	cnf := config.GetConfig()
	m.EXPECT().LoginUser(ctx, login).Return(int32(1), passHash, nil)
	m.EXPECT().LoginUser(ctx, login).Return(int32(0), "", errors.New("user not exists"))
	m.EXPECT().LoginUser(ctx, login).Return(int32(0), "", model.ErrInvalidCredentials)
	srv := NewService(m, cnf)

	t.Run("test ok", func(t *testing.T) {
		accessToken, err := srv.LoginUser(ctx, login, password)
		assert.Nil(t, err)
		require.NotEmpty(t, accessToken)
		user, _ := srv.ParseAuthToken(accessToken)
		require.Equal(t, int32(1), user.ID)
	})

	t.Run("test user not exists", func(t *testing.T) {
		accessToken, err := srv.LoginUser(ctx, login, password)
		assert.NotNil(t, err)
		assert.Equal(t, "user not exists", err.Error())
		require.Empty(t, accessToken)
	})

	t.Run("test wrong password", func(t *testing.T) {
		accessToken, err := srv.LoginUser(ctx, login, "wrongpassword")
		assert.NotNil(t, err)
		assert.Equal(t, model.ErrInvalidCredentials, err)
		require.Empty(t, accessToken)
	})
}

func TestService_SetVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepository(ctrl)
	cnf := config.GetConfig()
	srv := NewService(m, cnf)
	userID := int32(1)
	user := model.User{ID: userID}
	accessToken, err := srv.GetAuthToken(user)
	assert.Nil(t, err)
	assert.NotEmpty(t, accessToken)
	user.AccessToken = accessToken
	ctx := context.Background()
	ctx = srv.SetUser(ctx, user)
	key := "7dbf6d7e83e3396a18bd9bc99050a098da6f478fe294ac61d43d326481c12a7d"
	cnf.Security.MasterKey = key
	meta := "meta"
	filename := "filename"
	userdata := []byte("ciphertext")

	m.EXPECT().CreateVault(ctx, gomock.Any()).Return(int32(1), nil)
	m.EXPECT().CreateVault(ctx, gomock.Any()).Return(int32(0), model.ErrVaultAlreadyExists)

	t.Run("test ok", func(t *testing.T) {
		ID, err := srv.SetVault(ctx, pb.DataType(2), meta, filename, userdata)
		assert.Nil(t, err)
		require.Equal(t, int32(1), ID)
	})

	t.Run("test conflict", func(t *testing.T) {
		ID, err := srv.SetVault(ctx, pb.DataType(2), meta, filename, userdata)
		assert.NotNil(t, err)
		assert.Equal(t, model.ErrVaultAlreadyExists, err)
		require.Empty(t, ID)
	})

	t.Run("test user not exists", func(t *testing.T) {
		ctx := context.Background()
		ID, err := srv.SetVault(ctx, pb.DataType(2), meta, filename, userdata)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "user not found")
		require.Empty(t, ID)
	})

	t.Run("test wrong master key", func(t *testing.T) {
		cnf.Security.MasterKey = "aa50ec2616890da80f7d71a2a59cdd55"
		ID, err := srv.SetVault(ctx, pb.DataType(2), meta, filename, userdata)
		assert.NotNil(t, err)
		assert.Equal(t, "master key must be 32 bytes", err.Error())
		require.Empty(t, ID)
	})
}

func TestService_GetVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepository(ctrl)
	cnf := config.GetConfig()
	srv := NewService(m, cnf)
	userID := int32(1)
	user := model.User{ID: userID}
	accessToken, err := srv.GetAuthToken(user)
	assert.Nil(t, err)
	assert.NotEmpty(t, accessToken)
	user.AccessToken = accessToken
	ctx := context.Background()
	ctx = srv.SetUser(ctx, user)
	key := "7dbf6d7e83e3396a18bd9bc99050a098da6f478fe294ac61d43d326481c12a7d"
	cnf.Security.MasterKey = key
	meta := "meta"
	filename := "filename"
	k, err := getMasterKey(key)
	assert.Nil(t, err)
	userdata := []byte("ciphertext")
	encrypted, err := Encrypt(k, userdata)
	uv := model.UserVault{
		Datatype: 3,
		Meta:     meta,
		Filename: filename,
		Userdata: encrypted,
	}
	wrongUserdata := []byte("0x7DA9EA1E7B3995515841804DB2171CC836D4B5E39AD38C7B9CB4F096A5AA589D00B363DDC3C4B406")
	uvWrong := model.UserVault{
		Datatype: 3,
		Meta:     meta,
		Filename: filename,
		Userdata: wrongUserdata,
	}
	m.EXPECT().GetVault(ctx, int32(1), userID).Return(uv, nil)
	m.EXPECT().GetVault(ctx, int32(1), userID).Return(uvWrong, nil)
	m.EXPECT().GetVault(ctx, int32(1), userID).Return(model.UserVault{}, model.ErrVaultNotFound)
	m.EXPECT().GetVault(ctx, int32(1), userID).Return(uv, nil)

	t.Run("test ok", func(t *testing.T) {
		userV, err := srv.GetVault(ctx, 1)
		assert.Nil(t, err)
		require.Equal(t, meta, userV.Meta)
		require.Equal(t, userdata, userV.Userdata)
	})

	t.Run("test cipher error", func(t *testing.T) {
		userV, err := srv.GetVault(ctx, 1)
		assert.NotNil(t, err)
		assert.Equal(t, "cipher: message authentication failed", err.Error())
		require.Empty(t, userV)
	})

	t.Run("test conflict", func(t *testing.T) {
		userV, err := srv.GetVault(ctx, 1)
		assert.NotNil(t, err)
		assert.Equal(t, model.ErrVaultNotFound, err)
		require.Empty(t, userV)
	})

	t.Run("test user not exists", func(t *testing.T) {
		ctx := context.Background()
		userV, err := srv.GetVault(ctx, 1)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "user not found")
		require.Empty(t, userV)
	})

	t.Run("test wrong master key", func(t *testing.T) {
		cnf.Security.MasterKey = "aa50ec2616890da80f7d71a2a59cdd55"
		userV, err := srv.GetVault(ctx, 1)
		assert.NotNil(t, err)
		assert.Equal(t, "master key must be 32 bytes", err.Error())
		require.Empty(t, userV)
	})
}

func TestService_ListVaults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepository(ctrl)
	cnf := config.GetConfig()
	srv := NewService(m, cnf)
	userID := int32(1)
	user := model.User{ID: userID}
	accessToken, err := srv.GetAuthToken(user)
	assert.Nil(t, err)
	assert.NotEmpty(t, accessToken)
	user.AccessToken = accessToken
	ctx := context.Background()
	ctx = srv.SetUser(ctx, user)
	uvs := []model.UserVault{}
	uvs = append(uvs, model.UserVault{
		ID:       1,
		Datatype: 2,
		Meta:     "meta",
		Userdata: []byte("userdata"),
	})
	uvs = append(uvs, model.UserVault{
		ID:       2,
		Datatype: 3,
		Meta:     "meta2",
		Userdata: []byte("userdata2"),
	})

	m.EXPECT().ListVaults(ctx, userID).Return(uvs, nil)
	m.EXPECT().ListVaults(ctx, userID).Return([]model.UserVault{}, nil)

	t.Run("test ok", func(t *testing.T) {
		usvaults, err := srv.ListVaults(ctx)
		assert.Nil(t, err)
		require.Equal(t, 2, len(usvaults))
		for i, uv := range usvaults {
			require.Equal(t, uvs[i].ID, uv.ID)
			require.Equal(t, uvs[i].Datatype, uv.Datatype)
			require.Equal(t, uvs[i].Meta, uv.Meta)
			require.Equal(t, uvs[i].Userdata, uv.Userdata)
		}
	})

	t.Run("test user not exists", func(t *testing.T) {
		ctx := context.Background()
		usvaults, err := srv.ListVaults(ctx)
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "user not found")
		require.Empty(t, usvaults)
	})

	t.Run("test no vaults", func(t *testing.T) {
		usvaults, err := srv.ListVaults(ctx)
		assert.Nil(t, err)
		require.Empty(t, usvaults)
	})
}

func TestService_DeleteVault(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockRepository(ctrl)
	cnf := config.GetConfig()
	srv := NewService(m, cnf)
	userID := int32(1)
	user := model.User{ID: userID}
	accessToken, err := srv.GetAuthToken(user)
	assert.Nil(t, err)
	assert.NotEmpty(t, accessToken)
	user.AccessToken = accessToken
	ctx := context.Background()
	ctx = srv.SetUser(ctx, user)

	m.EXPECT().DeleteVault(ctx, int32(1), userID).Return(nil)
	m.EXPECT().DeleteVault(ctx, int32(1), userID).Return(model.ErrVaultNotFound)

	t.Run("test ok", func(t *testing.T) {
		err := srv.DeleteVault(ctx, int32(1))
		assert.Nil(t, err)
	})

	t.Run("test user not exists", func(t *testing.T) {
		ctx := context.Background()
		err := srv.DeleteVault(ctx, int32(1))
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "user not found")
	})

	t.Run("test no vaults", func(t *testing.T) {
		err := srv.DeleteVault(ctx, int32(1))
		assert.NotNil(t, err)
		assert.Equal(t, model.ErrVaultNotFound, err)
	})
}
