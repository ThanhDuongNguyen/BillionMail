package rbac

import (
	v1 "billionmail-core/api/rbac/v1"
	"billionmail-core/internal/service/public"
	service "billionmail-core/internal/service/rbac"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
)

// PermissionList retrieves the permission list
func (c *ControllerV1) PermissionList(ctx context.Context, req *v1.PermissionListReq) (res *v1.PermissionListRes, err error) {
	res = &v1.PermissionListRes{}

	if !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	permissions, total, err := service.Permission().GetList(ctx, req.Page, req.PageSize, req.Module, req.Action, req.Status)
	if err != nil {
		err = fmt.Errorf("failed to get permission list: %w", err)
		return
	}

	list := make([]v1.PermissionInfoItem, 0, len(permissions))
	for _, p := range permissions {
		list = append(list, v1.PermissionInfoItem{
			Id:          p.PermissionId,
			Name:        p.PermissionName,
			Description: p.Description,
			Module:      p.Module,
			Action:      p.Action,
			Resource:    p.Resource,
			Status:      p.Status,
			CreateTime:  p.CreateTime,
			UpdateTime:  p.UpdateTime,
		})
	}

	res.Data.List = list
	res.Data.Total = total
	res.Data.Page = req.Page
	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}
