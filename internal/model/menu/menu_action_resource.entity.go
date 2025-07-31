package menu

import (
	"context"
	"github.com/he2dou/go-admin/internal/model/util"
	"github.com/he2dou/go-admin/internal/schema"
	"github.com/he2dou/go-admin/internal/utils/structure"
	"gorm.io/gorm"
)

func GetMenuActionResourceDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDBWithModel(ctx, defDB, new(MenuActionResource))
}

type SchemaMenuActionResource schema.MenuActionResource

func (a SchemaMenuActionResource) ToMenuActionResource() *MenuActionResource {
	item := new(MenuActionResource)
	structure.Copy(a, item)
	return item
}

type MenuActionResource struct {
	util.Model
	ActionID uint64 `gorm:"index;not null;"` // 菜单动作ID
	Method   string `gorm:"size:50;"`        // 资源请求方式(支持正则)
	Path     string `gorm:"size:255;"`       // 资源请求路径（支持/:id匹配）
}

func (a MenuActionResource) ToSchemaMenuActionResource() *schema.MenuActionResource {
	item := new(schema.MenuActionResource)
	structure.Copy(a, item)
	return item
}

type MenuActionResources []*MenuActionResource

func (a MenuActionResources) ToSchemaMenuActionResources() []*schema.MenuActionResource {
	list := make([]*schema.MenuActionResource, len(a))
	for i, item := range a {
		list[i] = item.ToSchemaMenuActionResource()
	}
	return list
}
