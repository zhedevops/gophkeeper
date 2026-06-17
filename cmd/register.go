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

		err := ValidatePassword(password)
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
	registerCmd.Flags().StringVarP(&password, "password", "p", "", "Password")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// registerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// registerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
