package command

import (
	log "github.com/sirupsen/logrus"

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

	database.InitDB()
	defer database.CloseDB()

	sqlDB, err := database.PostgresDB.DB()
	if err != nil {
		log.Fatalf("Failed to connect to db: %v", err)
	}

	var dir string = "./database/migrations"
	if direction == "up" {
		err = goose.Up(sqlDB, dir)
	} else {
		err = goose.Down(sqlDB, dir)
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
