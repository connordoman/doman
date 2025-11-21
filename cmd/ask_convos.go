package cmd

import (
	"fmt"

	"github.com/connordoman/doman/internal/data"
	"github.com/connordoman/doman/internal/txt"
	"github.com/spf13/cobra"
)

var AskConvosCommand = &cobra.Command{
	Use:   "convos",
	Short: "List all conversations",
	RunE:  runAskConvos,
}

func runAskConvos(cmd *cobra.Command, args []string) error {
	convos, err := data.ListConversations(10, 0)
	if err != nil {
		return err
	}
	for _, convo := range convos {
		createdAt := convo.CreatedAt.Local().Format("Jan 02, 2006 15:04:05 MST")
		fmt.Println(txt.Greyf("%s", createdAt), txt.Bluef("%s", convo.ID))
	}
	return nil
}
