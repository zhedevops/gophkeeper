package client

import (
	"gophkeeper/internal/config"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	cnf := &config.Config{
		GRPCAddress: "localhost:3200",
		Security: config.Security{
			TLSCert: "../../testdata/server.crt",
		},
	}
	client, conn, err := New(cnf)

	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, conn)

	_ = conn.Close()
}
