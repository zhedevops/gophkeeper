package cmd

import (
	"errors"
	"fmt"
	pb "gophkeeper/proto"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Удалить пользовательские данные",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, conn, err := getClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		ctx := cmd.Context()

		if vaultID == 0 {
			return errors.New("id is required")
		}

		_, err = client.DeleteVault(
			ctx,
			pb.DeleteVaultRequest_builder{
				Id: &vaultID,
			}.Build(),
		)
		if err != nil {
			return err
		}

		err = cacheRepo.Delete(ctx, vaultID)
		if err != nil {
			return err
		}
		fmt.Printf("vault %d deleted\n", vaultID)

		return nil
	},
}

func init() {
	vaultCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().Int32VarP(&vaultID, "id", "i", 0, "userdata ID")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
