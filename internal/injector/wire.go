//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package injector

import (
	"github.com/google/wire"
	"github.com/he2dou/go-admin/internal/api"
	"github.com/he2dou/go-admin/internal/model/adapter"
	"github.com/he2dou/go-admin/internal/router"
	service "github.com/he2dou/go-admin/internal/serivce"
)

func BuildInjector() (*Injector, func(), error) {
	wire.Build(
		InitGormDB,
		model.RepoSet,
		InitAuth,
		InitCasbin,
		InitGinEngine,
		service.ServiceSet,
		api.APISet,
		router.RouterSet,
		adapter.CasbinAdapterSet,
		InjectorSet,
	)
	return new(Injector), nil, nil
}
