package menu

import (
	"context"
	"github.com/he2dou/go-admin/internal/model/util"
	"github.com/he2dou/go-admin/internal/schema"
	"github.com/he2dou/go-admin/internal/utils/structure"

	"gorm.io/gorm"
)

func GetMenuActionDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDBWithModel(ctx, defDB, new(MenuAction))
}

type SchemaMenuAction schema.MenuAction

func (a SchemaMenuAction) ToMenuAction() *MenuAction {
	item := new(MenuAction)
	structure.Copy(a, item)
	return item
}

type MenuAction struct {
	util.Model
	MenuID uint64 `gorm:"index;not null;"` // 菜单ID
	Code   string `gorm:"size:100;"`       // 动作编号
	Name   string `gorm:"size:100;"`       // 动作名称
}

func (a MenuAction) ToSchemaMenuAction() *schema.MenuAction {
	item := new(schema.MenuAction)
	structure.Copy(a, item)
	return item
}

type MenuActions []*MenuAction

func (a MenuActions) ToSchemaMenuActions() []*schema.MenuAction {
	list := make([]*schema.MenuAction, len(a))
	for i, item := range a {
		list[i] = item.ToSchemaMenuAction()
	}
	return list
}
