package rbac

import (
	v1 "billionmail-core/api/rbac/v1"
	"billionmail-core/internal/service/public"
	service "billionmail-core/internal/service/rbac"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
)

// RoleList retrieves the role list
func (c *ControllerV1) RoleList(ctx context.Context, req *v1.RoleListReq) (res *v1.RoleListRes, err error) {
	res = &v1.RoleListRes{}

	if !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	roles, total, err := service.Role().GetList(ctx, req.Page, req.PageSize, req.Name, req.Status)
	if err != nil {
		err = fmt.Errorf("failed to get role list: %w", err)
		return
	}

	list := make([]v1.RoleInfoItem, 0, len(roles))
	for _, r := range roles {
		list = append(list, v1.RoleInfoItem{
			Id:          r.RoleId,
			Name:        r.RoleName,
			Description: r.Description,
			Status:      r.Status,
			CreateTime:  r.CreateTime,
		})
	}

	res.Data.List = list
	res.Data.Total = total
	res.Data.Page = req.Page
	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

// RoleDetail retrieves role details including permissions
func (c *ControllerV1) RoleDetail(ctx context.Context, req *v1.RoleDetailReq) (res *v1.RoleDetailRes, err error) {
	res = &v1.RoleDetailRes{}

	if !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	role, err := service.Role().GetById(ctx, req.RoleId)
	if err != nil {
		err = fmt.Errorf("failed to get role: %w", err)
		return
	}

	res.Data.Role = v1.RoleInfoItem{
		Id:          role.RoleId,
		Name:        role.RoleName,
		Description: role.Description,
		Status:      role.Status,
		CreateTime:  role.CreateTime,
	}

	// Get role permissions
	permissions, _ := service.Role().GetPermissions(ctx, req.RoleId)
	res.Data.Permissions = make([]v1.PermissionInfoItem, 0, len(permissions))
	for _, p := range permissions {
		res.Data.Permissions = append(res.Data.Permissions, v1.PermissionInfoItem{
			Id:          p.PermissionId,
			Name:        p.PermissionName,
			Description: p.Description,
			Module:      p.Module,
			Action:      p.Action,
			Resource:    p.Resource,
			Status:      p.Status,
		})
	}

	// Get all permissions
	allPermissions, _ := service.Permission().GetAll(ctx)
	res.Data.AllPermissions = make([]v1.PermissionInfoItem, 0, len(allPermissions))
	for _, p := range allPermissions {
		res.Data.AllPermissions = append(res.Data.AllPermissions, v1.PermissionInfoItem{
			Id:          p.PermissionId,
			Name:        p.PermissionName,
			Description: p.Description,
			Module:      p.Module,
			Action:      p.Action,
			Resource:    p.Resource,
			Status:      p.Status,
		})
	}

	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

// RoleCreate creates a new role
func (c *ControllerV1) RoleCreate(ctx context.Context, req *v1.RoleCreateReq) (res *v1.RoleCreateRes, err error) {
	res = &v1.RoleCreateRes{}

	if !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	// Check if role name exists
	exists, err := service.Role().NameExists(ctx, req.Name)
	if err != nil {
		err = fmt.Errorf("failed to check role name: %w", err)
		return
	}
	if exists {
		err = gerror.New("Role name already exists")
		return
	}

	roleId, err := service.Role().Create(ctx, req.Name, req.Description, req.Status)
	if err != nil {
		err = fmt.Errorf("failed to create role: %w", err)
		return
	}

	// Bind permissions if provided
	if len(req.PermissionIds) > 0 {
		err = service.Role().BindPermissions(ctx, roleId, req.PermissionIds)
		if err != nil {
			err = fmt.Errorf("failed to bind permissions: %w", err)
			return
		}
	}

	res.Data.RoleId = roleId
	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

// RoleUpdate updates a role
func (c *ControllerV1) RoleUpdate(ctx context.Context, req *v1.RoleUpdateReq) (res *v1.RoleUpdateRes, err error) {
	res = &v1.RoleUpdateRes{}

	if !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	// Prevent modifying admin role name
	role, err := service.Role().GetById(ctx, req.RoleId)
	if err != nil {
		err = fmt.Errorf("failed to get role: %w", err)
		return
	}
	if role.RoleName == "admin" && req.Name != "" && req.Name != "admin" {
		err = gerror.New("Cannot rename the admin role")
		return
	}

	err = service.Role().Update(ctx, req.RoleId, req.Name, req.Description, req.Status)
	if err != nil {
		err = fmt.Errorf("failed to update role: %w", err)
		return
	}

	// Update permissions if provided
	if req.PermissionIds != nil {
		err = service.Role().BindPermissions(ctx, req.RoleId, req.PermissionIds)
		if err != nil {
			err = fmt.Errorf("failed to update permissions: %w", err)
			return
		}
	}

	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

// RoleDelete deletes a role
func (c *ControllerV1) RoleDelete(ctx context.Context, req *v1.RoleDeleteReq) (res *v1.RoleDeleteRes, err error) {
	res = &v1.RoleDeleteRes{}

	if !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	// Prevent deleting admin role
	role, err := service.Role().GetById(ctx, req.RoleId)
	if err != nil {
		err = fmt.Errorf("failed to get role: %w", err)
		return
	}
	if role.RoleName == "admin" {
		err = gerror.New("Cannot delete the admin role")
		return
	}

	// Check if role has accounts
	hasAccounts, _ := service.Role().HasAccounts(ctx, req.RoleId)
	if hasAccounts {
		err = gerror.New("Cannot delete a role that has accounts assigned")
		return
	}

	err = service.Role().Delete(ctx, req.RoleId)
	if err != nil {
		err = fmt.Errorf("failed to delete role: %w", err)
		return
	}

	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}
