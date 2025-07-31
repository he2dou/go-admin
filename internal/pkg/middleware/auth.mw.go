package middleware

import (
	"github.com/he2dou/go-admin/internal/config"
	"github.com/he2dou/go-admin/internal/pkg/auth"
	"github.com/he2dou/go-admin/internal/pkg/contextx"
	"github.com/he2dou/go-admin/internal/pkg/errors"
	"github.com/he2dou/go-admin/internal/pkg/ginx"
	"github.com/he2dou/go-admin/internal/pkg/logger"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func wrapUserAuthContext(c *gin.Context, userID uint64, userName string) {
	ctx := contextx.NewUserID(c.Request.Context(), userID)
	ctx = contextx.NewUserName(ctx, userName)
	ctx = logger.NewUserIDContext(ctx, userID)
	ctx = logger.NewUserNameContext(ctx, userName)
	c.Request = c.Request.WithContext(ctx)
}

// Valid user token (jwt)
func UserAuthMiddleware(a auth.Auther, skippers ...SkipperFunc) gin.HandlerFunc {
	if !config.C.JWTAuth.Enable {
		return func(c *gin.Context) {
			wrapUserAuthContext(c, config.C.Root.UserID, config.C.Root.UserName)
			c.Next()
		}
	}

	return func(c *gin.Context) {
		if SkipHandler(c, skippers...) {
			c.Next()
			return
		}

		tokenUserID, err := a.ParseUserID(c.Request.Context(), ginx.GetToken(c))
		if err != nil {
			if err == auth.ErrInvalidToken {
				if config.C.IsDebugMode() {
					wrapUserAuthContext(c, config.C.Root.UserID, config.C.Root.UserName)
					c.Next()
					return
				}
				ginx.ResError(c, errors.ErrInvalidToken)
				return
			}
			ginx.ResError(c, errors.WithStack(err))
			return
		}

		idx := strings.Index(tokenUserID, "-")
		if idx == -1 {
			ginx.ResError(c, errors.ErrInvalidToken)
			return
		}

		userID, _ := strconv.ParseUint(tokenUserID[:idx], 10, 64)
		wrapUserAuthContext(c, userID, tokenUserID[idx+1:])
		c.Next()
	}
}
