package grpc

import (
	"context"
	"errors"
	"gophkeeper/internal/model"

	pb "gophkeeper/proto"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockService struct {
	user    model.User
	token   string
	vault   model.UserVault
	vaults  []model.UserVault
	vaultID int32
	err     error
}

func (m *mockService) CreateUser(ctx context.Context, username string, password string) (model.User, error) {
	return m.user, m.err
}

func (m *mockService) LoginUser(ctx context.Context, username string, password string) (string, error) {
	return m.token, m.err
}

func (m *mockService) SetVault(
	ctx context.Context,
	datatype pb.DataType,
	meta string,
	filename string,
	userdata []byte,
) (int32, error) {
	return m.vaultID, m.err
}

func (m *mockService) GetVault(ctx context.Context, ID int32) (model.UserVault, error) {
	return m.vault, m.err
}

func (m *mockService) ListVaults(ctx context.Context) ([]model.UserVault, error) {
	return m.vaults, m.err
}

func (m *mockService) DeleteVault(ctx context.Context, ID int32) error {
	return m.err
}

func TestRegister_Success(t *testing.T) {
	srv := &GophkeeperServiceServer{
		service: &mockService{user: model.User{ID: 1, AccessToken: "token"}},
	}

	username := "user"
	password := "pass"
	resp, err := srv.Register(
		context.Background(),
		pb.RegisterRequest_builder{
			Username: &username,
			Password: &password,
		}.Build(),
	)

	require.NoError(t, err)
	require.Equal(t, int32(1), resp.GetUserId())
	require.Equal(t, "token", resp.GetAccessToken())
}

func TestRegister_Error(t *testing.T) {
	srv := &GophkeeperServiceServer{
		service: &mockService{err: model.ErrUserAlreadyExists},
	}

	username := "user"
	password := "pass"
	resp, err := srv.Register(
		context.Background(),
		pb.RegisterRequest_builder{
			Username: &username,
			Password: &password,
		}.Build(),
	)

	require.ErrorIs(t, err, model.ErrUserAlreadyExists)
	require.Nil(t, resp)
}

func TestLogin_Success(t *testing.T) {
	srv := &GophkeeperServiceServer{
		service: &mockService{token: "token"},
	}

	username := "user"
	password := "pass"
	resp, err := srv.Login(
		context.Background(),
		pb.LoginRequest_builder{
			Username: &username,
			Password: &password,
		}.Build(),
	)

	require.NoError(t, err)
	require.Equal(t, "token", resp.GetAccessToken())
}

func TestLogin_Error(t *testing.T) {
	srv := &GophkeeperServiceServer{
		service: &mockService{err: model.ErrInvalidCredentials},
	}

	username := "user"
	password := "pass"
	resp, err := srv.Login(
		context.Background(),
		pb.LoginRequest_builder{
			Username: &username,
			Password: &password,
		}.Build(),
	)

	require.ErrorIs(t, err, model.ErrInvalidCredentials)
	require.Empty(t, resp.GetAccessToken())
}

func TestSetVault_Success(t *testing.T) {
	srv := &GophkeeperServiceServer{
		service: &mockService{vaultID: 1},
	}

	dt := pb.DataType_DATA_TYPE_TEXT
	meta := "meta"
	filename := ""
	resp, err := srv.SetVault(
		context.Background(),
		pb.SetVaultRequest_builder{
			Datatype: &dt,
			Meta:     &meta,
			Filename: &filename,
			Userdata: []byte("userdata"),
		}.Build(),
	)

	require.NoError(t, err)
	require.Equal(t, int32(1), resp.GetId())
}

func TestSetVault_Error(t *testing.T) {
	srv := &GophkeeperServiceServer{
		service: &mockService{err: model.ErrVaultAlreadyExists},
	}

	dt := pb.DataType_DATA_TYPE_TEXT
	meta := "meta"
	filename := ""
	resp, err := srv.SetVault(
		context.Background(),
		pb.SetVaultRequest_builder{
			Datatype: &dt,
			Meta:     &meta,
			Filename: &filename,
			Userdata: []byte("userdata"),
		}.Build(),
	)

	require.ErrorIs(t, err, model.ErrVaultAlreadyExists)
	require.Empty(t, resp.GetId())
}

func TestGetVault_Success(t *testing.T) {
	srv := &GophkeeperServiceServer{
		service: &mockService{
			vault: model.UserVault{
				ID:       1,
				Datatype: 2,
				Meta:     "meta",
				Userdata: []byte("userdata"),
			},
		},
	}

	vaultID := int32(1)
	resp, err := srv.GetVault(
		context.Background(),
		pb.GetVaultRequest_builder{Id: &vaultID}.Build(),
	)

	require.NoError(t, err)
	require.Equal(t, int32(1), resp.GetId())
	require.Equal(t, pb.DataType(2), resp.GetDatatype())
	require.Equal(t, "meta", resp.GetMeta())
	require.Equal(t, []byte("userdata"), resp.GetUserdata())
}

func TestGetVault_Error(t *testing.T) {
	srv := &GophkeeperServiceServer{
		service: &mockService{err: model.ErrVaultNotFound},
	}

	vaultID := int32(1)
	resp, err := srv.GetVault(
		context.Background(),
		pb.GetVaultRequest_builder{Id: &vaultID}.Build(),
	)

	require.ErrorIs(t, err, model.ErrVaultNotFound)
	require.Empty(t, resp.GetId())
}

func TestListVaults_Success(t *testing.T) {
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
	srv := &GophkeeperServiceServer{
		service: &mockService{vaults: uvs},
	}

	resp, err := srv.ListVaults(
		context.Background(),
		pb.ListVaultsRequest_builder{}.Build(),
	)

	require.NoError(t, err)
	require.Equal(t, 2, len(resp.GetItems()))
	for i, rec := range resp.GetItems() {
		require.Equal(t, uvs[i].ID, rec.GetId())
		require.Equal(t, pb.DataType(uvs[i].Datatype), rec.GetDatatype())
		require.Equal(t, uvs[i].Meta, rec.GetMeta())
	}

}

func TestListVaults_Error(t *testing.T) {
	expectedErr := errors.New("test error")
	srv := &GophkeeperServiceServer{
		service: &mockService{
			err: expectedErr,
		},
	}

	resp, err := srv.ListVaults(
		context.Background(),
		pb.ListVaultsRequest_builder{}.Build(),
	)

	require.ErrorIs(t, err, expectedErr)
	require.Empty(t, resp.GetItems())
}

func TestDeleteVault_Success(t *testing.T) {
	srv := &GophkeeperServiceServer{
		service: &mockService{},
	}

	vaultID := int32(1)
	_, err := srv.DeleteVault(
		context.Background(),
		pb.DeleteVaultRequest_builder{Id: &vaultID}.Build(),
	)

	require.NoError(t, err)
}

func TestDeleteVault_Error(t *testing.T) {
	srv := &GophkeeperServiceServer{
		service: &mockService{err: model.ErrVaultNotFound},
	}

	vaultID := int32(1)
	_, err := srv.DeleteVault(
		context.Background(),
		pb.DeleteVaultRequest_builder{Id: &vaultID}.Build(),
	)

	require.ErrorIs(t, err, model.ErrVaultNotFound)
}
