package app

import (
	"github.com/gin-gonic/gin"
	"github.com/he2dou/go-admin/internal/config"
	"github.com/he2dou/go-admin/internal/pkg/gzip"
	"github.com/he2dou/go-admin/internal/pkg/middleware"
	"github.com/he2dou/go-admin/internal/router"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/gin-swagger/swaggerFiles"
)

func InitGinEngine(r router.IRouter) *gin.Engine {
	gin.SetMode(config.App.RunMode)

	app := gin.New()
	app.NoMethod(middleware.NoMethodHandler())
	app.NoRoute(middleware.NoRouteHandler())

	prefixes := r.Prefixes()

	// Recover
	app.Use(middleware.RecoveryMiddleware())

	// Trace ID
	app.Use(middleware.TraceMiddleware(middleware.AllowPathPrefixNoSkipper(prefixes...)))

	// Copy body
	app.Use(middleware.CopyBodyMiddleware(middleware.AllowPathPrefixNoSkipper(prefixes...)))

	// Access logger
	app.Use(middleware.LoggerMiddleware(middleware.AllowPathPrefixNoSkipper(prefixes...)))

	// CORS
	if config.App.CORS.Enable {
		app.Use(middleware.CORSMiddleware())
	}

	// GZIP
	if config.App.GZIP.Enable {
		app.Use(gzip.Gzip(gzip.BestCompression,
			gzip.WithExcludedExtensions(config.App.GZIP.ExcludedExtentions),
			gzip.WithExcludedPaths(config.App.GZIP.ExcludedPaths),
		))
	}

	// Router register
	r.Register(app)

	// Swagger
	if config.App.Swagger {
		app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// Website
	if dir := config.App.WWW; dir != "" {
		app.Use(middleware.WWWMiddleware(dir, middleware.AllowPathPrefixSkipper(prefixes...)))
	}

	return app
}
