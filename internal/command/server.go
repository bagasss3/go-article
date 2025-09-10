package command

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	handler "github.com/bagasss3/go-article/internal/api/http"
	"github.com/bagasss3/go-article/internal/config"
	"github.com/bagasss3/go-article/internal/infrastructure/cache"
	"github.com/bagasss3/go-article/internal/infrastructure/database"
	"github.com/bagasss3/go-article/internal/infrastructure/server"
	"github.com/bagasss3/go-article/internal/middleware"
	"github.com/bagasss3/go-article/internal/repository"
	"github.com/bagasss3/go-article/internal/service"
	"github.com/bagasss3/go-article/pkg/model"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"s"},
	Short:   "run server",
	Long:    "Start running the server",
	Run:     runServer,
}

func init() {
	RootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, args []string) {
	log.SetReportCaller(true)
	log.SetFormatter(&log.JSONFormatter{})

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize DBs
	database.InitDB()
	defer database.CloseDB()

	redisConn := database.NewRedisConn(config.RedisHost())
	defer redisConn.Close()

	// cache
	cacher := cache.NewRedisCache(redisConn)

	// Initialize Echo
	httpServer := server.NewHTTPServer()

	// Depedency injection
	corsMiddleware := middleware.ModuleCorsMiddleware(httpServer.Engine())
	corsMiddleware.Setup()

	articleRepository := repository.NewArticleRepository(database.PostgresDB, cacher)

	articleService := service.NewArticleService(articleRepository)

	registerHandlers(httpServer.Engine(), articleService)

	// Setup signal handling
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	// Start server
	go func() {
		log.WithField("addr", config.Port()).Info("Starting server")

		if err := httpServer.Start(config.Port()); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Server failed unexpectedly")
		}
	}()

	// Wait for interrupt signal
	<-signalCh
	log.Info("Received shutdown signal")

	// Gracefully shutdown the server
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := httpServer.Engine().Shutdown(shutdownCtx); err != nil {
		log.WithError(err).Error("Server shutdown failed")
	}

	log.Info("Server shutdown complete")
}

func registerHandlers(e *echo.Echo, articleSvc model.ArticleMethodService) {
	v1 := e.Group("/api/v1")

	handler.NewArticleHandler(articleSvc).Register(v1)
}
