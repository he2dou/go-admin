package user

import (
	"context"

	"github.com/he2dou/go-admin/internal/model/util"
	"github.com/he2dou/go-admin/internal/schema"
	"github.com/he2dou/go-admin/internal/utils/structure"

	"gorm.io/gorm"
)

func GetUserRoleDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDBWithModel(ctx, defDB, new(UserRole))
}

type SchemaUserRole schema.UserRole

func (a SchemaUserRole) SetUserRole() *UserRole {
	item := new(UserRole)
	structure.Copy(a, item)
	return item
}

type UserRole struct {
	util.Model
	UserID uint64 `gorm:"index;default:0;"` // 用户内码
	RoleID uint64 `gorm:"index;default:0;"` // 角色内Set码
}

func (a UserRole) GetUserRole() *schema.UserRole {
	item := new(schema.UserRole)
	structure.Copy(a, item)
	return item
}

type UserRoles []*UserRole

func (a UserRoles) GetUserRoles() []*schema.UserRole {
	list := make([]*schema.UserRole, len(a))
	for i, item := range a {
		list[i] = item.GetUserRole()
	}
	return list
}
