package cmd

import (
	"errors"
	pb "gophkeeper/proto"

	"github.com/rs/zerolog/log"
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

		log.Info().Int32("vault id", vaultID).Msg("user data deleted")

		return nil
	},
}

func init() {
	vaultCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().Int32VarP(&vaultID, "id", "i", 0, "userdata ID")
}
