package database_initialization

import (
	"context"
	"github.com/gogf/gf/v2/frame/g"
	"time"
)

// defaultPermission defines a permission to seed
type defaultPermission struct {
	Name        string
	Description string
	Module      string
	Action      string
	Resource    string
}

// Default permissions for the system
var defaultPermissions = []defaultPermission{
	// Campaign module
	{"campaign:read:campaign", "View campaigns", "campaign", "read", "campaign"},
	{"campaign:create:campaign", "Create campaigns", "campaign", "create", "campaign"},
	{"campaign:update:campaign", "Update campaigns", "campaign", "update", "campaign"},
	{"campaign:delete:campaign", "Delete campaigns", "campaign", "delete", "campaign"},
	{"campaign:read:tracking", "View campaign tracking details", "campaign", "read", "tracking"},
	{"campaign:export:tracking", "Export campaign tracking data", "campaign", "export", "tracking"},

	// Contact module - group level (can select groups)
	{"contact:read:group", "View contact groups", "contact", "read", "group"},
	{"contact:create:group", "Create contact groups", "contact", "create", "group"},
	{"contact:update:group", "Update contact groups", "contact", "update", "group"},
	{"contact:delete:group", "Delete contact groups", "contact", "delete", "group"},

	// Contact module - subscriber level (can see individual contacts)
	{"contact:read:subscriber", "View subscribers detail", "contact", "read", "subscriber"},
	{"contact:create:subscriber", "Create subscribers", "contact", "create", "subscriber"},
	{"contact:update:subscriber", "Update subscribers", "contact", "update", "subscriber"},
	{"contact:delete:subscriber", "Delete subscribers", "contact", "delete", "subscriber"},
	{"contact:export:subscriber", "Export subscribers", "contact", "export", "subscriber"},

	// Domain module
	{"domain:read:domain", "View domains", "domain", "read", "domain"},
	{"domain:create:domain", "Create domains", "domain", "create", "domain"},
	{"domain:update:domain", "Update domains", "domain", "update", "domain"},
	{"domain:delete:domain", "Delete domains", "domain", "delete", "domain"},

	// Mailbox module
	{"mailbox:read:mailbox", "View mailboxes", "mailbox", "read", "mailbox"},
	{"mailbox:create:mailbox", "Create mailboxes", "mailbox", "create", "mailbox"},
	{"mailbox:update:mailbox", "Update mailboxes", "mailbox", "update", "mailbox"},
	{"mailbox:delete:mailbox", "Delete mailboxes", "mailbox", "delete", "mailbox"},

	// Template module
	{"template:read:template", "View email templates", "template", "read", "template"},
	{"template:create:template", "Create email templates", "template", "create", "template"},
	{"template:update:template", "Update email templates", "template", "update", "template"},
	{"template:delete:template", "Delete email templates", "template", "delete", "template"},

	// Settings module
	{"settings:read:settings", "View settings", "settings", "read", "settings"},
	{"settings:update:settings", "Update settings", "settings", "update", "settings"},

	// Overview/Dashboard
	{"overview:read:overview", "View dashboard overview", "overview", "read", "overview"},

	// Logs module
	{"logs:read:logs", "View operation logs", "logs", "read", "logs"},

	// SMTP/Relay module
	{"smtp:read:smtp", "View SMTP relay settings", "smtp", "read", "smtp"},
	{"smtp:update:smtp", "Update SMTP relay settings", "smtp", "update", "smtp"},
}

func init() {
	registerHandler(func() {
		seedDefaultPermissions()
	})
}

func seedDefaultPermissions() {
	ctx := context.Background()
	now := time.Now().Unix()

	for _, perm := range defaultPermissions {
		// Check if permission already exists
		count, err := g.DB().Model("permission").
			Where("module = ? AND action = ? AND resource = ?", perm.Module, perm.Action, perm.Resource).
			Count()
		if err != nil {
			g.Log().Error(ctx, "Failed to check permission:", perm.Name, err)
			continue
		}
		if count > 0 {
			continue
		}

		// Insert new permission
		_, err = g.DB().Model("permission").Data(g.Map{
			"permission_name": perm.Name,
			"description":     perm.Description,
			"module":          perm.Module,
			"action":          perm.Action,
			"resource":        perm.Resource,
			"status":          1,
			"create_time":     now,
			"update_time":     now,
		}).Insert()
		if err != nil {
			g.Log().Error(ctx, "Failed to insert permission:", perm.Name, err)
		}
	}

	g.Log().Info(ctx, "Default permissions seeded successfully")
}
