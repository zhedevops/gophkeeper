package cmd

import (
	pb "gophkeeper/proto"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Авторизация пользователя",
	RunE: func(cmd *cobra.Command, args []string) error {
		login, _ := cmd.Flags().GetString("login")
		password, err := readPassword()
		if err != nil {
			return err
		}

		client, conn, err := getClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		ctx := cmd.Context()

		resp, err := client.Login(
			ctx,
			pb.LoginRequest_builder{
				Username: &login,
				Password: &password,
			}.Build(),
		)
		if err != nil {
			return err
		}

		err = SaveToken(resp.GetAccessToken())
		if err != nil {
			return err
		}

		log.Info().Msg("user authenticated")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().StringP("login", "l", "", "Login")
}
