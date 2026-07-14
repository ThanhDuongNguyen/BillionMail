package rbac

import (
	v1 "billionmail-core/api/rbac/v1"
	"billionmail-core/internal/model"
	"billionmail-core/internal/service/public"
	service "billionmail-core/internal/service/rbac"
	"context"
	"fmt"
	"github.com/gogf/gf/v2/errors/gerror"
)

// AccountList retrieves the account list
func (c *ControllerV1) AccountList(ctx context.Context, req *v1.AccountListReq) (res *v1.AccountListRes, err error) {
	res = &v1.AccountListRes{}

	// Only admin can list accounts
	if !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	accounts, total, err := service.Account().GetList(ctx, req.Page, req.PageSize, req.Username, req.Email, req.Status)
	if err != nil {
		err = fmt.Errorf("failed to get account list: %w", err)
		return
	}

	list := make([]v1.AccountInfoItem, 0, len(accounts))
	for _, acc := range accounts {
		roles, _ := service.Account().GetAccountRoles(ctx, acc.AccountId)
		roleNames := make([]string, 0, len(roles))
		for _, r := range roles {
			roleNames = append(roleNames, r.RoleName)
		}
		list = append(list, v1.AccountInfoItem{
			Id:         acc.AccountId,
			Username:   acc.Username,
			Email:      acc.Email,
			Status:     acc.Status,
			Language:   acc.Language,
			Roles:      roleNames,
			CreateTime: acc.CreateTime,
		})
	}

	res.Data.List = list
	res.Data.Total = total
	res.Data.Page = req.Page
	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

// AccountDetail retrieves account details
func (c *ControllerV1) AccountDetail(ctx context.Context, req *v1.AccountDetailReq) (res *v1.AccountDetailRes, err error) {
	res = &v1.AccountDetailRes{}

	if !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	account, err := service.Account().GetById(ctx, req.AccountId)
	if err != nil {
		err = fmt.Errorf("failed to get account: %w", err)
		return
	}

	res.Data.Account = v1.AccountInfoItem{
		Id:         account.AccountId,
		Username:   account.Username,
		Email:      account.Email,
		Status:     account.Status,
		Language:   account.Language,
		CreateTime: account.CreateTime,
	}

	// Get roles for this account
	roles, _ := service.Account().GetAccountRoles(ctx, req.AccountId)
	res.Data.Roles = make([]v1.RoleInfoItem, 0, len(roles))
	for _, r := range roles {
		res.Data.Roles = append(res.Data.Roles, v1.RoleInfoItem{
			Id:          r.RoleId,
			Name:        r.RoleName,
			Description: r.Description,
			Status:      r.Status,
			CreateTime:  r.CreateTime,
		})
	}

	// Get all roles
	allRoles, _ := service.Role().GetAll(ctx)
	res.Data.AllRoles = make([]v1.RoleInfoItem, 0, len(allRoles))
	for _, r := range allRoles {
		res.Data.AllRoles = append(res.Data.AllRoles, v1.RoleInfoItem{
			Id:          r.RoleId,
			Name:        r.RoleName,
			Description: r.Description,
			Status:      r.Status,
			CreateTime:  r.CreateTime,
		})
	}

	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

// AccountCreate creates a new account
func (c *ControllerV1) AccountCreate(ctx context.Context, req *v1.AccountCreateReq) (res *v1.AccountCreateRes, err error) {
	res = &v1.AccountCreateRes{}

	if !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	// Check if username exists
	exists, err := service.Account().UsernameExists(ctx, req.Username)
	if err != nil {
		err = fmt.Errorf("failed to check username: %w", err)
		return
	}
	if exists {
		err = gerror.New("Username already exists")
		return
	}

	// Create account
	accountId, err := service.Account().Create(ctx, &model.Account{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
		Status:   req.Status,
		Language: req.Lang,
	})
	if err != nil {
		err = fmt.Errorf("failed to create account: %w", err)
		return
	}

	// Assign roles
	if len(req.RoleIds) > 0 {
		err = service.Account().BindRoles(ctx, accountId, req.RoleIds)
		if err != nil {
			err = fmt.Errorf("failed to assign roles: %w", err)
			return
		}
	}

	res.Data.AccountId = accountId
	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

// AccountUpdate updates an account
func (c *ControllerV1) AccountUpdate(ctx context.Context, req *v1.AccountUpdateReq) (res *v1.AccountUpdateRes, err error) {
	res = &v1.AccountUpdateRes{}

	if !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	// Get existing account
	account, err := service.Account().GetById(ctx, req.AccountId)
	if err != nil {
		err = fmt.Errorf("failed to get account: %w", err)
		return
	}

	// Update fields
	if req.Username != "" {
		account.Username = req.Username
	}
	if req.Email != "" {
		account.Email = req.Email
	}
	if req.Status >= 0 {
		account.Status = req.Status
	}
	if req.Lang != "" {
		account.Language = req.Lang
	}

	err = service.Account().Update(ctx, account)
	if err != nil {
		err = fmt.Errorf("failed to update account: %w", err)
		return
	}

	// Update roles if provided
	if req.RoleIds != nil {
		err = service.Account().BindRoles(ctx, req.AccountId, req.RoleIds)
		if err != nil {
			err = fmt.Errorf("failed to update roles: %w", err)
			return
		}
	}

	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

// AccountPassword updates account password
func (c *ControllerV1) AccountPassword(ctx context.Context, req *v1.AccountPasswordReq) (res *v1.AccountPasswordRes, err error) {
	res = &v1.AccountPasswordRes{}

	currentAccountId := service.GetCurrentAccountId(ctx)

	// Users can change their own password; admin can change any password
	if currentAccountId != req.AccountId && !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	// Non-admin must verify old password
	if !isAdmin(ctx) {
		account, getErr := service.Account().GetById(ctx, req.AccountId)
		if getErr != nil {
			err = fmt.Errorf("failed to get account: %w", getErr)
			return
		}
		if !service.Account().VerifyPassword(account.Password, req.OldPassword) {
			err = gerror.New("Old password is incorrect")
			return
		}
	}

	err = service.Account().UpdatePassword(ctx, req.AccountId, req.NewPassword)
	if err != nil {
		err = fmt.Errorf("failed to update password: %w", err)
		return
	}

	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

// AccountDelete deletes an account
func (c *ControllerV1) AccountDelete(ctx context.Context, req *v1.AccountDeleteReq) (res *v1.AccountDeleteRes, err error) {
	res = &v1.AccountDeleteRes{}

	if !isAdmin(ctx) {
		err = gerror.New("Permission denied")
		return
	}

	// Prevent deleting the last admin
	isTargetAdmin, _ := service.Account().IsAdmin(ctx, req.AccountId)
	if isTargetAdmin {
		adminCount, _ := service.Account().CountAdmins(ctx)
		if adminCount <= 1 {
			err = gerror.New("Cannot delete the last admin account")
			return
		}
	}

	// Prevent self-deletion
	currentAccountId := service.GetCurrentAccountId(ctx)
	if currentAccountId == req.AccountId {
		err = gerror.New("Cannot delete your own account")
		return
	}

	err = service.Account().Delete(ctx, req.AccountId)
	if err != nil {
		err = fmt.Errorf("failed to delete account: %w", err)
		return
	}

	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

// isAdmin checks if the current user has admin role
func isAdmin(ctx context.Context) bool {
	roles := service.GetCurrentRoles(ctx)
	for _, role := range roles {
		if role == "admin" {
			return true
		}
	}
	return false
}
