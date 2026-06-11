package cmd

import (
	"errors"
	"fmt"
	"gophkeeper/internal/model"
	pb "gophkeeper/proto"
	"path/filepath"

	"github.com/spf13/cobra"
)

// MaxFileSize Максимальный размер хранимых данных в одной записи
const MaxFileSize = 5 * 1024 * 1024 // 5 MB

var (
	typeStr string
	meta    string
	data    string
	number  string
	holder  string
	expiry  string
	cvv     string
	file    string
)

// CreateOptions Тип с пользовательскими данными для хранения
type CreateOptions struct {
	Type string

	Meta string

	Login    string
	Password string

	Data string
	File string

	Number string
	Holder string
	Expiry string
	CVV    string
}

// CredentialsData Данные логин/пароль
type CredentialsData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// CardData Данные банковских карт
type CardData struct {
	Number string `json:"number"`
	Holder string `json:"holder"`
	Expiry string `json:"expiry"`
	CVV    string `json:"cvv"`
}

// createCmd represents the create command
var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Сохранение пользовательских данных",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, conn, err := getClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		ctx := cmd.Context()

		opts := CreateOptions{
			Type:     typeStr,
			Meta:     meta,
			Login:    login,
			Password: password,
			Data:     data,
			File:     file,
			Number:   number,
			Holder:   holder,
			Expiry:   expiry,
			CVV:      cvv,
		}

		if meta == "" {
			return errors.New("meta is required")
		}

		dt, payload, err := ParseData(opts)
		if err != nil {
			return err
		}

		var filename string
		if file != "" {
			filename = filepath.Base(file)
		}

		resp, err := client.SetVault(
			ctx,
			pb.SetVaultRequest_builder{
				Datatype: &dt,
				Meta:     &meta,
				Filename: &filename,
				Userdata: payload,
			}.Build(),
		)
		if err != nil {
			return err
		}
		fmt.Println("vault created with id: ", resp.GetId())

		vc := model.VaultCache{
			ID:       resp.GetId(),
			Datatype: int32(dt),
			Meta:     meta,
			Filename: filename,
			Userdata: payload,
		}
		err = cacheRepo.Save(ctx, vc)
		if err != nil {
			return err
		}

		fmt.Println("cache created with id: ", resp.GetId())
		return nil
	},
}

func init() {
	vaultCmd.AddCommand(createCmd)
	createCmd.Flags().StringVarP(&typeStr, "type", "t", "", "credentials|text|binary|card")
	createCmd.Flags().StringVarP(&meta, "meta", "m", "", "record metadata")
	createCmd.Flags().StringVarP(&login, "login", "l", "", "login")
	createCmd.Flags().StringVarP(&password, "password", "p", "", "password")
	createCmd.Flags().StringVarP(&data, "data", "d", "", "text data")
	createCmd.Flags().StringVarP(&number, "number", "n", "", "card number")
	createCmd.Flags().StringVarP(&holder, "holder", "o", "", "holder")
	createCmd.Flags().StringVarP(&expiry, "expiry", "e", "", "card expiry")
	createCmd.Flags().StringVarP(&cvv, "cvv", "c", "", "card cvv")
	createCmd.Flags().StringVarP(&file, "file", "f", "", "file path")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// createCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// createCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
