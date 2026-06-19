package cmd

import (
	pb "gophkeeper/proto"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "Получить список сохранённых пользовательских данных",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, conn, err := getClient()
		if err != nil {
			return err
		}
		defer conn.Close()

		ctx := cmd.Context()

		resp, err := client.ListVaults(
			ctx,
			pb.ListVaultsRequest_builder{}.Build(),
		)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.Unavailable {
				// сервер недоступен
				vcs, err := cacheRepo.List(ctx)
				if err != nil {
					return err
				}
				PrintRecordsCache(vcs)

				return nil
			}
			return err
		}

		PrintRecords(resp)

		return nil
	},
}

func init() {
	vaultCmd.AddCommand(listCmd)
}
