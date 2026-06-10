package cmd

import (
	"context"
	"fmt"
	pb "gophkeeper/proto"

	"github.com/spf13/cobra"
)

var login string
var password string

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Регистрация нового пользователя",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("register called")
		//fmt.Println(login, password)

		client, conn, err := getClient()
		if err != nil {
			return err
		}
		defer conn.Close()
		// todo валидация пароля не менее 12 символов цифры, строчные и прописные, хотя бы 1 символ

		// 1. хранение данных в зашифрованном виде
		// 2. офлайн режим работы??? локально в sqllite сохранять данные те же что на сервере,
		//а в офлайн только читать

		// передавать токен после авторизации
		// md := metadata.New(map[string]string{"token": SecretToken})
		// ctx = metadata.NewOutgoingContext(ctx, md)
		resp, err := client.Register(
			context.Background(),
			pb.RegisterRequest_builder{
				Username: &login,
				Password: &password,
			}.Build(),
		)
		if err != nil {
			return err
		}
		fmt.Println("register result: ", resp)
		// сохранить токен в кеш а потом token := LoadToken()
		// md := metadata.New(map[string]string{
		//	"authorization": "Bearer " + token,
		//})
		//
		//ctx := metadata.NewOutgoingContext(context.Background(), md)
		err = SaveToken(resp.GetAccessToken())
		if err != nil {
			return err
		}

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
