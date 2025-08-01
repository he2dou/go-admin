package model

import (
	"github.com/he2dou/go-admin/internal/config"
	"github.com/he2dou/go-admin/internal/model/menu"
	"github.com/he2dou/go-admin/internal/model/role"
	"github.com/he2dou/go-admin/internal/model/user"
	"github.com/he2dou/go-admin/internal/model/util"
	"strings"

	"github.com/google/wire"
	"gorm.io/gorm"
) // end

// RepoSet repo injection
var RepoSet = wire.NewSet(
	util.TransSet,
	menu.MenuActionResourceSet,
	menu.MenuActionSet,
	menu.MenuSet,
	role.RoleMenuSet,
	role.RoleSet,
	user.UserRoleSet,
	user.UserSet,
) // end

// Define repo type alias
type (
	TransRepo              = util.Trans
	MenuActionResourceRepo = menu.MenuActionResourceRepo
	MenuActionRepo         = menu.MenuActionRepo
	MenuRepo               = menu.MenuRepo
	RoleMenuRepo           = role.RoleMenuRepo
	RoleRepo               = role.RoleRepo
	UserRoleRepo           = user.UserRoleRepo
	UserRepo               = user.UserRepo
) // end

// Auto migration for given models
func AutoMigrate(db *gorm.DB) error {
	if dbType := config.App.Gorm.DBType; strings.ToLower(dbType) == "mysql" {
		db = db.Set("gorm:table_options", "ENGINE=InnoDB")
	}

	return db.AutoMigrate(
		new(menu.MenuActionResource),
		new(menu.MenuAction),
		new(menu.Menu),
		new(role.RoleMenu),
		new(role.Role),
		new(user.UserRole),
		new(user.User),
	) // end
}
