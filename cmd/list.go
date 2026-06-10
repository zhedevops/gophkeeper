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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
