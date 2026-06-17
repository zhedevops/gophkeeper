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
		password, _ := cmd.Flags().GetString("password")

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
	loginCmd.Flags().StringP("password", "p", "", "Password")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// loginCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// loginCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
