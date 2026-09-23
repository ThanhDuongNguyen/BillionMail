package batch_mail

import (
	v1 "billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/model/entity"
	"billionmail-core/internal/service/mail_service"
	"billionmail-core/internal/service/public"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/util/grand"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const (
	// Attachment storage directory (relative to the core working dir, shared volume).
	attachmentStorageDir = "../conf/attachments"

	// Limits
	maxAttachmentCount     = 10
	maxAttachmentFileSize  = 15 * 1024 * 1024 // 15 MB per file
	maxAttachmentTotalSize = 25 * 1024 * 1024 // 25 MB total
)

// blockedAttachmentExts are executable/script extensions that are never allowed.
var blockedAttachmentExts = map[string]struct{}{
	".exe": {}, ".bat": {}, ".cmd": {}, ".com": {}, ".msi": {},
	".scr": {}, ".vbs": {}, ".js": {}, ".jse": {}, ".jar": {},
	".sh": {}, ".ps1": {}, ".psm1": {}, ".dll": {}, ".app": {},
	".pif": {}, ".hta": {}, ".wsf": {}, ".reg": {},
}

// attachmentDir returns the storage directory for a given task.
func attachmentDir(taskId int) string {
	return public.AbsPath(fmt.Sprintf("%s/%d", attachmentStorageDir, taskId))
}

// ValidateAttachmentInputs validates the raw attachment inputs before saving.
func ValidateAttachmentInputs(ctx context.Context, inputs []v1.TaskAttachment) error {
	if len(inputs) == 0 {
		return nil
	}
	if len(inputs) > maxAttachmentCount {
		return gerror.New(public.LangCtx(ctx, "Attachments cannot exceed {} files", maxAttachmentCount))
	}

	var total int64
	for _, in := range inputs {
		name := strings.TrimSpace(in.Filename)
		if name == "" {
			return gerror.New(public.LangCtx(ctx, "Attachment file name cannot be empty"))
		}

		ext := strings.ToLower(filepath.Ext(name))
		if _, blocked := blockedAttachmentExts[ext]; blocked {
			return gerror.New(public.LangCtx(ctx, "Attachment type {} is not allowed", ext))
		}

		// For newly uploaded files we validate the decoded size; for existing
		// files (edit mode, no Content) we trust the stored metadata.
		if in.Content != "" {
			raw, err := decodeBase64Content(in.Content)
			if err != nil {
				return gerror.New(public.LangCtx(ctx, "Attachment {} is not valid base64", name))
			}
			size := int64(len(raw))
			if size > maxAttachmentFileSize {
				return gerror.New(public.LangCtx(ctx, "Attachment {} exceeds the per-file size limit", name))
			}
			total += size
		} else {
			total += in.Size
		}
	}

	if total > maxAttachmentTotalSize {
		return gerror.New(public.LangCtx(ctx, "Total attachment size exceeds the limit"))
	}

	return nil
}

// decodeBase64Content decodes a base64 string, tolerating a data URI prefix.
func decodeBase64Content(content string) ([]byte, error) {
	if idx := strings.Index(content, ","); strings.HasPrefix(content, "data:") && idx != -1 {
		content = content[idx+1:]
	}
	content = strings.TrimSpace(content)
	return base64.StdEncoding.DecodeString(content)
}

// SaveTaskAttachments persists attachment binaries to disk for the given task
// and returns the metadata (without base64 content) to be stored in the DB.
//
// Inputs without a Content field but with a Path are treated as already-saved
// attachments (edit mode) and are kept as-is.
func SaveTaskAttachments(ctx context.Context, taskId int, inputs []v1.TaskAttachment) ([]entity.AttachmentMeta, error) {
	if len(inputs) == 0 {
		return []entity.AttachmentMeta{}, nil
	}

	dir := attachmentDir(taskId)
	if !public.FileExists(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, gerror.New(public.LangCtx(ctx, "Failed to create attachment directory: {}", err.Error()))
		}
	}

	items := make([]entity.AttachmentMeta, 0, len(inputs))
	for _, in := range inputs {
		// Existing attachment (edit mode): keep as-is.
		if in.Content == "" && in.Path != "" {
			items = append(items, entity.AttachmentMeta{
				Filename:    in.Filename,
				ContentType: in.ContentType,
				Size:        in.Size,
				Path:        in.Path,
			})
			continue
		}

		raw, err := decodeBase64Content(in.Content)
		if err != nil {
			return nil, gerror.New(public.LangCtx(ctx, "Attachment {} is not valid base64", in.Filename))
		}

		// Generate a safe, unique on-disk name; keep the original extension.
		ext := filepath.Ext(in.Filename)
		storedName := fmt.Sprintf("%d_%s%s", taskId, grand.S(12), ext)
		relPath := fmt.Sprintf("%s/%d/%s", attachmentStorageDir, taskId, storedName)
		absPath := filepath.Join(dir, storedName)

		if err := os.WriteFile(absPath, raw, 0644); err != nil {
			return nil, gerror.New(public.LangCtx(ctx, "Failed to save attachment: {}", err.Error()))
		}

		contentType := in.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		items = append(items, entity.AttachmentMeta{
			Filename:    in.Filename,
			ContentType: contentType,
			Size:        int64(len(raw)),
			Path:        relPath,
		})
	}

	return items, nil
}

// LoadAttachmentContent reads the binary content of a stored attachment from disk.
func LoadAttachmentContent(item entity.AttachmentMeta) ([]byte, error) {
	absPath := public.AbsPath(item.Path)
	return os.ReadFile(absPath)
}

// RemoveTaskAttachments deletes the whole attachment directory for a task.
func RemoveTaskAttachments(taskId int) {
	dir := attachmentDir(taskId)
	if public.FileExists(dir) {
		_ = os.RemoveAll(dir)
	}
}

// MarshalAttachments serializes attachment metadata for DB storage.
func MarshalAttachments(items []entity.AttachmentMeta) string {
	if len(items) == 0 {
		return "[]"
	}
	b, err := json.Marshal(items)
	if err != nil {
		g.Log().Warning(context.Background(), "marshal attachments failed:", err)
		return "[]"
	}
	return string(b)
}

// LoadTaskAttachmentsForSend reads a task's attachment files from disk and
// converts them into mail_service.Attachment values ready to be attached to an email.
// Attachments that fail to load are skipped (logged), so a single bad file does not
// abort the whole send.
func LoadTaskAttachmentsForSend(ctx context.Context, task *entity.EmailTask) []mail_service.Attachment {
	if task == nil || len(task.Attachments) == 0 {
		return nil
	}
	result := make([]mail_service.Attachment, 0, len(task.Attachments))
	for _, meta := range task.Attachments {
		data, err := LoadAttachmentContent(meta)
		if err != nil {
			g.Log().Warning(ctx, "failed to load attachment for task", task.Id, meta.Path, err)
			continue
		}
		result = append(result, mail_service.Attachment{
			Filename:    meta.Filename,
			ContentType: meta.ContentType,
			Data:        data,
		})
	}
	return result
}

// ParseAttachments deserializes stored attachment metadata.
func ParseAttachments(raw string) []entity.AttachmentMeta {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "[]" {
		return nil
	}
	var items []entity.AttachmentMeta
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		g.Log().Warning(context.Background(), "parse attachments failed:", err)
		return nil
	}
	return items
}
