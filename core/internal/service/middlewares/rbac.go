package middlewares

import (
	"billionmail-core/internal/service/public"
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"

	"billionmail-core/internal/service/rbac"
)

// PathToRouteInfo converts path to module, action, and resource
func PathToRouteInfo(path string) (module, action, resource string) {
	// Map of path patterns to permission components
	// Format: path prefix -> module, default resource
	pathMappings := map[string]struct {
		module   string
		resource string
	}{
		"/api/campaign":        {"campaign", "campaign"},
		"/api/contact/group":   {"contact", "group"},
		"/api/contact":         {"contact", "subscriber"},
		"/api/subscribe_list":  {"contact", "group"},
		"/api/domains":         {"domain", "domain"},
		"/api/mail_boxes":      {"mailbox", "mailbox"},
		"/api/email_template":  {"template", "template"},
		"/api/settings":        {"settings", "settings"},
		"/api/overview":        {"overview", "overview"},
		"/api/operation_log":   {"logs", "logs"},
		"/api/relay":           {"smtp", "smtp"},
		"/api/batch_mail/tracking": {"campaign", "tracking"},
		"/api/batch_mail":      {"campaign", "campaign"},
		"/api/tags":            {"contact", "subscriber"},
	}

	// Check longest prefix first (more specific paths first)
	bestMatch := ""
	for prefix := range pathMappings {
		if strings.HasPrefix(path, prefix) && len(prefix) > len(bestMatch) {
			bestMatch = prefix
		}
	}

	if bestMatch == "" {
		return "", "", ""
	}

	module = pathMappings[bestMatch].module
	resource = pathMappings[bestMatch].resource

	// Determine action from path suffix or HTTP method context
	lowerPath := strings.ToLower(path)
	switch {
	case strings.Contains(lowerPath, "/create") || strings.Contains(lowerPath, "/add"):
		action = "create"
	case strings.Contains(lowerPath, "/update") || strings.Contains(lowerPath, "/edit") || strings.Contains(lowerPath, "/set_"):
		action = "update"
	case strings.Contains(lowerPath, "/delete") || strings.Contains(lowerPath, "/remove"):
		action = "delete"
	case strings.Contains(lowerPath, "/export"):
		action = "export"
	default:
		action = "read"
	}

	return
}

// RBACMiddleware handles permission verification for HTTP requests
type RBACMiddleware struct {
	PermissionService rbac.IPermission
}

// NewRBACMiddleware creates a new RBACMiddleware
func NewRBACMiddleware() *RBACMiddleware {
	return &RBACMiddleware{
		PermissionService: rbac.Permission(),
	}
}

// PermissionCheck checks if the current user has the required permission
func (m *RBACMiddleware) PermissionCheck(r *ghttp.Request) {
	// Skip permission check for authentication-related routes
	if r.URL.Path == "/api/login" ||
		r.URL.Path == "/api/refresh-token" ||
		r.URL.Path == "/api/get_validate_code" ||
		r.URL.Path == "/api/current-user" ||
		r.URL.Path == "/api/languages/set" ||
		r.URL.Path == "/api/languages/get" {
		r.Middleware.Next()
		return
	}

	// Skip RBAC management routes (they have their own admin check in controllers)
	if strings.HasPrefix(r.URL.Path, "/api/account/") ||
		strings.HasPrefix(r.URL.Path, "/api/role/") ||
		strings.HasPrefix(r.URL.Path, "/api/permission/") {
		r.Middleware.Next()
		return
	}

	// Skip common/shared APIs that all authenticated users need access to
	commonPaths := []string{
		"/api/settings/get_version",
		"/api/settings/get_language",
		"/api/domains/all",
		"/api/overview/info",
		"/api/files/",
		"/api/askai/",
		"/api/tags/",
	}
	for _, path := range commonPaths {
		if r.URL.Path == path || strings.HasPrefix(r.URL.Path, path) {
			r.Middleware.Next()
			return
		}
	}

	// Extract account ID from context
	accountIdVar := r.GetCtxVar("accountId")
	if accountIdVar == nil {
		r.Response.WriteJson(public.CodeMap[401])
		r.Exit()
		return
	}
	accountId := gconv.Int64(accountIdVar)

	// Get roles from context
	roles := r.GetCtxVar("roles", []string{}).Strings()

	// Check for admin role (has all permissions)
	for _, role := range roles {
		if role == "admin" {
			r.Middleware.Next()
			return
		}
	}

	// Extract module, action, and resource from request path
	module, action, resource := PathToRouteInfo(r.URL.Path)

	// If we couldn't determine the module, action, or resource, allow the request
	// (unknown routes are not protected by RBAC, they rely on JWT auth only)
	if module == "" || action == "" || resource == "" {
		r.Middleware.Next()
		return
	}

	// Check if user has the required permission
	hasPermission, err := m.PermissionService.Check(r.GetCtx(), accountId, module, action, resource)
	if err != nil {
		g.Log().Error(r.GetCtx(), "Permission check error:", err)
		r.Response.WriteJson(g.Map{
			"code":    500,
			"msg":     "Error checking permissions",
			"success": false,
		})
		r.Exit()
		return
	}

	if !hasPermission {
		r.Response.WriteJson(g.Map{
			"code":    403,
			"msg":     "Insufficient permissions",
			"success": false,
		})
		r.Exit()
		return
	}

	r.Middleware.Next()
}

// HasPermission checks if the current user has a specific permission
func HasPermission(ctx context.Context, module, action, resource string) bool {
	accountId := rbac.GetCurrentAccountId(ctx)
	if accountId == 0 {
		return false
	}

	// Get roles from context
	rolesVar := ctx.Value("roles")
	roles := []string{}
	if rolesVar != nil {
		roles = rolesVar.([]string)
	}

	// Check for admin role (has all permissions)
	for _, role := range roles {
		if role == "admin" {
			return true
		}
	}

	// Check specific permission
	permissionService := rbac.Permission()
	hasPermission, err := permissionService.Check(ctx, accountId, module, action, resource)
	if err != nil {
		g.Log().Error(ctx, "Permission check error:", err)
		return false
	}

	return hasPermission
}

// RequirePermission returns a middleware handler that checks for a specific permission
// RequirePermission middleware checks if user has required permission
func RequirePermission(module, action, resource string) ghttp.HandlerFunc {
	return func(r *ghttp.Request) {
		if !HasPermission(r.GetCtx(), module, action, resource) {
			r.Response.WriteJson(g.Map{
				"code": 403,
				"msg":  "Insufficient permissions",
			})
			r.Exit()
			return
		}
		r.Middleware.Next()
	}
}
