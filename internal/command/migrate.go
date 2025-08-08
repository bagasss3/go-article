package command

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/bagasss3/go-article/internal/config"
	"github.com/bagasss3/go-article/internal/infrastructure/database"
	"github.com/pressly/goose"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "run migrate database",
	Long:  "Start migrate database",
	Run:   migration,
}

func init() {
	migrateCmd.PersistentFlags().String("direction", "up", "migration direction up/down")
	RootCmd.AddCommand(migrateCmd)
}

func migration(cmd *cobra.Command, args []string) {
	direction := cmd.Flag("direction").Value.String()

	err := goose.SetDialect("postgres")
	if err != nil {
		log.Error(err)
	}
	goose.SetTableName("schema_migrations")
	ctx := context.Background()

	db, err := database.InitDB(ctx, config.DBDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	var dir string = "./database/migrations"
	if direction == "up" {
		err = goose.Up(db, dir)
	} else {
		err = goose.Down(db, dir)
	}

	if err != nil {
		log.WithFields(log.Fields{
			"direction": direction}).
			Fatal("Failed to migrate database: ", err)
	}

	log.WithFields(log.Fields{
		"direction": direction,
	}).Info("Success applied migrations!")

}
