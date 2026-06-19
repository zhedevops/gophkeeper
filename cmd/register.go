package cmd

import (
	"errors"
	pb "gophkeeper/proto"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var login string
var password string

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Регистрация нового пользователя",
	RunE: func(cmd *cobra.Command, args []string) error {
		if login == "" {
			return errors.New("login id is required")
		}
		password, err := readPassword()
		if err != nil {
			return err
		}

		err = ValidatePassword(password)
		if err != nil {
			return err
		}

		client, conn, err := getClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		ctx := cmd.Context()

		resp, err := client.Register(
			ctx,
			pb.RegisterRequest_builder{
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

		log.Info().Msg("user successfully registered")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
	registerCmd.Flags().StringVarP(&login, "login", "l", "", "Login")
}
