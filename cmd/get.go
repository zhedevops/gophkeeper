package cmd

import (
	"errors"
	pb "gophkeeper/proto"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var vaultID int32

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Получение пользовательских данных",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, conn, err := getClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		ctx := cmd.Context()

		if vaultID == 0 {
			return errors.New("meta is required")
		}

		resp, err := client.GetVault(
			ctx,
			pb.GetVaultRequest_builder{
				Id: &vaultID,
			}.Build(),
		)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.Unavailable {
				// сервер недоступен
				vc, err := cacheRepo.Get(ctx, vaultID)
				if err != nil {
					return err
				}
				PrintRecordCache(vc)

				return nil
			}
			return err
		}

		PrintRecord(resp)

		return nil
	},
}

func init() {
	vaultCmd.AddCommand(getCmd)
	getCmd.Flags().Int32VarP(&vaultID, "id", "i", 0, "userdata ID")
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
