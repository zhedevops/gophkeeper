package cmd

import (
	"context"
	"fmt"
	"gophkeeper/internal/cache"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

var cacheRepo *cache.Repository

// vaultCmd represents the vault command
var vaultCmd = &cobra.Command{
	Use:   "vault",
	Short: "Пользовательские данные",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		token, err := LoadToken()
		if err != nil {
			return err
		}
		md := metadata.New(map[string]string{
			"authorization": "Bearer " + token,
		})

		ctx := metadata.NewOutgoingContext(context.Background(), md)

		cmd.SetContext(ctx)

		fmt.Println("init cache")
		if cacheRepo != nil {
			return nil
		}

		dbPath, err := CacheDBPath()
		if err != nil {
			return err
		}

		cacheRepo, err = cache.New(dbPath)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(vaultCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// vaultCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// vaultCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
