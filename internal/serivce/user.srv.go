package service

import (
	"context"
	"github.com/he2dou/go-admin/internal/model"
	"github.com/he2dou/go-admin/internal/pkg/errors"
	"github.com/he2dou/go-admin/internal/schema"
	"github.com/he2dou/go-admin/internal/utils/hash"
	"github.com/he2dou/go-admin/internal/utils/snowflake"
	"strconv"

	"github.com/casbin/casbin/v2"
	"github.com/google/wire"
)

var UserSet = wire.NewSet(wire.Struct(new(UserSrv), "*"))

type UserSrv struct {
	Enforcer     *casbin.SyncedEnforcer
	TransRepo    *model.TransRepo
	UserRepo     *model.UserRepo
	UserRoleRepo *model.UserRoleRepo
	RoleRepo     *model.RoleRepo
}

func (a *UserSrv) Query(ctx context.Context, params schema.UserQueryParam, opts ...schema.UserQueryOptions) (*schema.UserQueryResult, error) {
	return a.UserRepo.Query(ctx, params, opts...)
}

func (a *UserSrv) QueryShow(ctx context.Context, params schema.UserQueryParam, opts ...schema.UserQueryOptions) (*schema.UserShowQueryResult, error) {
	result, err := a.UserRepo.Query(ctx, params, opts...)
	if err != nil {
		return nil, err
	} else if result == nil {
		return nil, nil
	}

	userRoleResult, err := a.UserRoleRepo.Query(ctx, schema.UserRoleQueryParam{
		UserIDs: result.Data.GetIDs(),
	})
	if err != nil {
		return nil, err
	}

	roleResult, err := a.RoleRepo.Query(ctx, schema.RoleQueryParam{
		IDs: userRoleResult.Data.GetRoleIDs(),
	})
	if err != nil {
		return nil, err
	}

	return result.GetShowResult(userRoleResult.Data.GetUserIDMap(), roleResult.Data.GetMap()), nil
}

func (a *UserSrv) Get(ctx context.Context, id uint64, opts ...schema.UserQueryOptions) (*schema.User, error) {
	item, err := a.UserRepo.Get(ctx, id, opts...)
	if err != nil {
		return nil, err
	} else if item == nil {
		return nil, errors.ErrNotFound
	}

	userRoleResult, err := a.UserRoleRepo.Query(ctx, schema.UserRoleQueryParam{
		UserID: id,
	})
	if err != nil {
		return nil, err
	}
	item.UserRoles = userRoleResult.Data

	return item, nil
}

func (a *UserSrv) Create(ctx context.Context, item schema.User) (*schema.IDResult, error) {
	err := a.checkUserName(ctx, item)
	if err != nil {
		return nil, err
	}

	item.Password = hash.SHA1String(item.Password)
	item.ID = snowflake.MustID()
	err = a.TransRepo.Exec(ctx, func(ctx context.Context) error {
		for _, urItem := range item.UserRoles {
			urItem.ID = snowflake.MustID()
			urItem.UserID = item.ID
			err := a.UserRoleRepo.Create(ctx, *urItem)
			if err != nil {
				return err
			}
		}

		return a.UserRepo.Create(ctx, item)
	})
	if err != nil {
		return nil, err
	}

	for _, urItem := range item.UserRoles {
		a.Enforcer.AddRoleForUser(strconv.FormatUint(urItem.UserID, 10), strconv.FormatUint(urItem.RoleID, 10))
	}

	return schema.NewIDResult(item.ID), nil
}

func (a *UserSrv) checkUserName(ctx context.Context, item schema.User) error {
	if item.UserName == schema.GetRootUser().UserName {
		return errors.New400Response("user_name has been exists")
	}

	result, err := a.UserRepo.Query(ctx, schema.UserQueryParam{
		PaginationParam: schema.PaginationParam{OnlyCount: true},
		UserName:        item.UserName,
	})
	if err != nil {
		return err
	} else if result.PageResult.Total > 0 {
		return errors.New400Response("user_name has been exists")
	}
	return nil
}

func (a *UserSrv) Update(ctx context.Context, id uint64, item schema.User) error {
	oldItem, err := a.Get(ctx, id)
	if err != nil {
		return err
	} else if oldItem == nil {
		return errors.ErrNotFound
	} else if oldItem.UserName != item.UserName {
		err := a.checkUserName(ctx, item)
		if err != nil {
			return err
		}
	}

	if item.Password != "" {
		item.Password = hash.SHA1String(item.Password)
	} else {
		item.Password = oldItem.Password
	}

	item.ID = oldItem.ID
	item.Creator = oldItem.Creator
	item.CreatedAt = oldItem.CreatedAt

	addUserRoles, delUserRoles := a.compareUserRoles(ctx, oldItem.UserRoles, item.UserRoles)
	err = a.TransRepo.Exec(ctx, func(ctx context.Context) error {
		for _, aitem := range addUserRoles {
			aitem.ID = snowflake.MustID()
			aitem.UserID = id
			err := a.UserRoleRepo.Create(ctx, *aitem)
			if err != nil {
				return err
			}
		}

		for _, ritem := range delUserRoles {
			err := a.UserRoleRepo.Delete(ctx, ritem.ID)
			if err != nil {
				return err
			}
		}

		return a.UserRepo.Update(ctx, id, item)
	})
	if err != nil {
		return err
	}

	for _, aitem := range addUserRoles {
		a.Enforcer.AddRoleForUser(strconv.FormatUint(id, 10), strconv.FormatUint(aitem.RoleID, 10))
	}

	for _, ritem := range delUserRoles {
		a.Enforcer.DeleteRoleForUser(strconv.FormatUint(id, 10), strconv.FormatUint(ritem.RoleID, 10))
	}

	return nil
}

func (a *UserSrv) compareUserRoles(ctx context.Context, oldUserRoles, newUserRoles schema.UserRoles) (addList, delList schema.UserRoles) {
	mOldUserRoles := oldUserRoles.GetMap()
	mNewUserRoles := newUserRoles.GetMap()

	for k, item := range mNewUserRoles {
		if _, ok := mOldUserRoles[k]; ok {
			delete(mOldUserRoles, k)
			continue
		}
		addList = append(addList, item)
	}

	for _, item := range mOldUserRoles {
		delList = append(delList, item)
	}
	return
}

func (a *UserSrv) Delete(ctx context.Context, id uint64) error {
	oldItem, err := a.UserRepo.Get(ctx, id)
	if err != nil {
		return err
	} else if oldItem == nil {
		return errors.ErrNotFound
	}

	err = a.TransRepo.Exec(ctx, func(ctx context.Context) error {
		err := a.UserRoleRepo.DeleteByUserID(ctx, id)
		if err != nil {
			return err
		}

		return a.UserRepo.Delete(ctx, id)
	})
	if err != nil {
		return err
	}

	a.Enforcer.DeleteUser(strconv.FormatUint(id, 10))
	return nil
}

func (a *UserSrv) UpdateStatus(ctx context.Context, id uint64, status int) error {
	oldItem, err := a.Get(ctx, id)
	if err != nil {
		return err
	} else if oldItem == nil {
		return errors.ErrNotFound
	} else if oldItem.Status == status {
		return nil
	}

	err = a.UserRepo.UpdateStatus(ctx, id, status)
	if err != nil {
		return err
	}

	if status == 1 {
		for _, uritem := range oldItem.UserRoles {
			a.Enforcer.AddRoleForUser(strconv.FormatUint(id, 10), strconv.FormatUint(uritem.RoleID, 10))
		}
	} else {
		a.Enforcer.DeleteUser(strconv.FormatUint(id, 10))
	}

	return nil
}
