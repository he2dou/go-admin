package cmd

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/he2dou/go-admin/internal/config"
	"github.com/he2dou/go-admin/internal/injector"
	"github.com/he2dou/go-admin/internal/pkg/captcha"
	"github.com/he2dou/go-admin/internal/pkg/captcha/store"
	"github.com/he2dou/go-admin/internal/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/gops/agent"
)

type options struct {
	ConfigFile string
	ModelFile  string
	MenuFile   string
	WWWDir     string
	Version    string
}

type Option func(*options)

func SetConfigFile(s string) Option {
	return func(o *options) {
		o.ConfigFile = s
	}
}

func SetModelFile(s string) Option {
	return func(o *options) {
		o.ModelFile = s
	}
}

func SetWWWDir(s string) Option {
	return func(o *options) {
		o.WWWDir = s
	}
}

func SetMenuFile(s string) Option {
	return func(o *options) {
		o.MenuFile = s
	}
}

func SetVersion(s string) Option {
	return func(o *options) {
		o.Version = s
	}
}

func Init(ctx context.Context, opts ...Option) (func(), error) {
	var o options
	for _, opt := range opts {
		opt(&o)
	}

	config.MustLoad(o.ConfigFile)
	if v := o.ModelFile; v != "" {
		config.App.Casbin.Model = v
	}
	if v := o.WWWDir; v != "" {
		config.App.WWW = v
	}
	if v := o.MenuFile; v != "" {
		config.App.Menu.Data = v
	}
	config.PrintWithJSON()

	logger.WithContext(ctx).Printf("Start server,#run_mode %s,#version %s,#pid %d", config.App.RunMode, o.Version, os.Getpid())

	loggerCleanFunc, err := injector.InitLogger()
	if err != nil {
		return nil, err
	}

	monitorCleanFunc := InitMonitor(ctx)

	InitCaptcha()

	injector, injectorCleanFunc, err := injector.BuildInjector()
	if err != nil {
		return nil, err
	}

	if config.App.Menu.Enable && config.App.Menu.Data != "" {
		err = injector.MenuSrv.InitData(ctx, config.App.Menu.Data)
		if err != nil {
			return nil, err
		}
	}

	httpServerCleanFunc := InitHTTPServer(ctx, injector.Engine)

	return func() {
		httpServerCleanFunc()
		injectorCleanFunc()
		monitorCleanFunc()
		loggerCleanFunc()
	}, nil
}

func InitCaptcha() {
	cfg := config.App.Captcha
	if cfg.Store == "redis" {
		rc := config.App.Redis
		captcha.SetCustomStore(store.NewRedisStore(&redis.Options{
			Addr:     rc.Addr,
			Password: rc.Password,
			DB:       cfg.RedisDB,
		}, captcha.Expiration, logger.StandardLogger(), cfg.RedisPrefix))
	}
}

func InitMonitor(ctx context.Context) func() {
	if c := config.App.Monitor; c.Enable {
		// ShutdownCleanup set false to prevent automatically closes on os.Interrupt
		// and close agent manually before service shutting down
		err := agent.Listen(agent.Options{Addr: c.Addr, ConfigDir: c.ConfigDir, ShutdownCleanup: false})
		if err != nil {
			logger.WithContext(ctx).Errorf("Agent monitor error: %s", err.Error())
		}
		return func() {
			agent.Close()
		}
	}
	return func() {}
}

func InitHTTPServer(ctx context.Context, handler http.Handler) func() {
	cfg := config.App.HTTP
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  30 * time.Second,
	}

	go func() {
		logger.WithContext(ctx).Printf("HTTP server is running at %s.", addr)

		var err error
		if cfg.CertFile != "" && cfg.KeyFile != "" {
			srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			err = srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && err != http.ErrServerClosed {
			panic(err)
		}

	}()

	return func() {
		ctx, cancel := context.WithTimeout(ctx, time.Second*time.Duration(cfg.ShutdownTimeout))
		defer cancel()

		srv.SetKeepAlivesEnabled(false)
		if err := srv.Shutdown(ctx); err != nil {
			logger.WithContext(ctx).Errorf(err.Error())
		}
	}
}

func Run(ctx context.Context, opts ...Option) error {
	state := 1
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	cleanFunc, err := Init(ctx, opts...)
	if err != nil {
		return err
	}

EXIT:
	for {
		sig := <-sc
		logger.WithContext(ctx).Infof("Receive signal[%s]", sig.String())
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			state = 0
			break EXIT
		case syscall.SIGHUP:
		default:
			break EXIT
		}
	}

	cleanFunc()
	logger.WithContext(ctx).Infof("Server exit")
	time.Sleep(time.Second)
	os.Exit(state)
	return nil
}
