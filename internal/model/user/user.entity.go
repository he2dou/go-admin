package user

import (
	"context"

	"github.com/he2dou/go-admin/internal/model/util"
	"github.com/he2dou/go-admin/internal/schema"
	"github.com/he2dou/go-admin/internal/utils/structure"

	"gorm.io/gorm"
)

func GetUserDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return util.GetDBWithModel(ctx, defDB, new(User))
}

type SchemaUser schema.User

func (a SchemaUser) SetUser() *User {
	item := new(User)
	structure.Copy(a, item)
	return item
}

type User struct {
	util.Model
	UserName string  `gorm:"size:64;uniqueIndex;default:'';not null;"` // 用户名
	RealName string  `gorm:"size:64;index;default:'';"`                // 真实姓名
	Password string  `gorm:"size:40;default:'';"`                      // 密码
	Email    *string `gorm:"size:255;"`                                // 邮箱
	Phone    *string `gorm:"size:20;"`                                 // 手机号
	Status   int     `gorm:"index;default:0;"`                         // 状态(1:启用 2:停用)
	Creator  uint64  `gorm:""`                                         // 创建者
}

func (a User) GetUser() *schema.User {
	item := new(schema.User)
	structure.Copy(a, item)
	return item
}

type Users []*User

func (a Users) GetUsers() []*schema.User {
	list := make([]*schema.User, len(a))
	for i, item := range a {
		list[i] = item.GetUser()
	}
	return list
}
