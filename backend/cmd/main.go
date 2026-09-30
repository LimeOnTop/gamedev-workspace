package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/LimeOnTop/gamedev-workspace/backend/cmd/config"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/apperr"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/controller"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/filestorage"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/mcp"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/middleware"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/repository"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/service"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/treecache"
	"github.com/LimeOnTop/gamedev-workspace/backend/internal/usecase"
)

const shutdownTimeout = 15 * time.Second

func main() {
	cfg := config.Load()

	apperr.Configure(cfg.DevMode)
	if !cfg.DevMode {
		gin.SetMode(gin.ReleaseMode)
	}

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		panic("open database: " + err.Error())
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxIdleTime(30 * time.Minute)
	defer db.Close()

	if err := waitForDatabase(db); err != nil {
		panic("connect database: " + err.Error())
	}

	var treeCache usecase.TreeCache = treecache.Noop{}
	if cfg.RedisURL != "" {
		opts, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			panic("parse REDIS_URL: " + err.Error())
		}
		client := redis.NewClient(opts)
		defer client.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := client.Ping(ctx).Err(); err != nil {
			log.Printf("redis unavailable, tree cache disabled: %v", err)
		} else {
			treeCache = treecache.NewStore(client, cfg.TreeCacheTTL)
			log.Printf("redis connected")
		}
		cancel()
	}

	files, err := filestorage.NewLocal(cfg.UploadDir)
	if err != nil {
		panic(err.Error())
	}

	nodeRepository := repository.NewNodeRepository(db)
	referenceRepository := repository.NewReferenceRepository(db)
	assetRepository := repository.NewAssetRepository(db)

	nodeService := service.NewNodeService(nodeRepository, referenceRepository, assetRepository, files, treeCache)
	referenceService := service.NewReferenceService(nodeRepository, referenceRepository, files, treeCache, cfg.MaxUploadSize)
	assetService := service.NewAssetService(assetRepository)

	nodeController := controller.NewNodeController(nodeService)
	referenceController := controller.NewReferenceController(referenceService, cfg.MaxUploadSize)
	assetController := controller.NewAssetController(assetService)
	mcpHandler := mcp.NewHandler(nodeService, referenceService, assetService, cfg.PublicURL)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.CORS())

	router.GET("/health", func(c *gin.Context) {
		if err := db.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// MCP streams are long-lived, so they are mounted outside the request logger.
	router.Any("/mcp", gin.WrapH(mcpHandler))

	logged := router.Group("/", middleware.Logger())
	api := logged.Group("/api")
	nodeController.Register(api)
	assetController.Register(api)
	referenceController.Register(api, logged)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("workspace service listening on %s (mcp: /mcp)", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	waitForShutdown(srv, errCh)
}

func waitForDatabase(db *sql.DB) error {
	var err error
	for attempt := 0; attempt < 30; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		err = db.PingContext(ctx)
		cancel()
		if err == nil {
			return nil
		}
		log.Printf("waiting for database: %v", err)
		time.Sleep(time.Second)
	}
	return err
}

func waitForShutdown(srv *http.Server, errCh <-chan error) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		log.Printf("shutdown signal received: %s", sig)
	case err := <-errCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic("http serve: " + err.Error())
		}
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown timed out, forcing stop: %v", err)
		srv.Close()
		return
	}
	log.Print("http server stopped")
}
