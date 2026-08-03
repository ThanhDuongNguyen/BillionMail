package batch_mail

import (
	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/public"
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

const maxExportRows = 50000

func (c *ControllerV1) ExportTaskRecipients(ctx context.Context, req *v1.ExportTaskRecipientsReq) (res *v1.ExportTaskRecipientsRes, err error) {
	g.Log().Printf(ctx, "Exporting task recipients, TaskId: %d, Type: %s", req.TaskId, req.Type)

	res = &v1.ExportTaskRecipientsRes{}

	// Verify task exists
	taskInfo, err := batch_mail.GetTaskInfo(ctx, req.TaskId)
	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to get task information: {}", err.Error())))
		return
	}

	if taskInfo == nil || taskInfo.Id == 0 {
		res.Code = 404
		res.SetError(gerror.New(public.LangCtx(ctx, "Task not found: {}", req.TaskId)))
		return
	}

	// Use a large page size to get all data (capped at maxExportRows)
	listReq := &v1.GetTaskRecipientsByStatusReq{
		TaskId:   req.TaskId,
		Type:     req.Type,
		Search:   req.Search,
		Page:     1,
		PageSize: maxExportRows,
	}

	var list []*v1.RecipientItem

	switch req.Type {
	case "delivered":
		_, list, err = queryDeliveredRecipients(ctx, listReq)
	case "bounced":
		_, list, err = queryBouncedRecipients(ctx, listReq)
	case "opened":
		_, list, err = queryOpenedRecipients(ctx, listReq)
	case "clicked":
		_, list, err = queryClickedRecipients(ctx, listReq)
	default:
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "Invalid status type: {}", req.Type)))
		return
	}

	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to query recipients: {}", err.Error())))
		return
	}

	// Generate CSV file
	var buf bytes.Buffer

	// Write UTF-8 BOM for Excel compatibility
	buf.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(&buf)

	// Headers
	headers := []string{"Email", "Time", "Mail Provider"}
	if req.Type == "clicked" {
		headers = append(headers, "URL")
	}
	writer.Write(headers)

	// Data rows
	for _, item := range list {
		row := []string{
			item.Recipient,
			time.Unix(item.Time, 0).Format("2006-01-02 15:04:05"),
			item.MailProvider,
		}
		if req.Type == "clicked" {
			row = append(row, item.Url)
		}
		writer.Write(row)
	}

	writer.Flush()
	if err = writer.Error(); err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to generate CSV file: {}", err.Error())))
		return
	}

	// Get the HTTP request from context and write file response directly
	r := g.RequestFromCtx(ctx)
	if r == nil {
		res.Code = 500
		res.SetError(gerror.New("Failed to get HTTP request from context"))
		return
	}

	filename := fmt.Sprintf("task_%d_%s_%s.csv", req.TaskId, req.Type, time.Now().Format("20060102_150405"))
	r.Response.Header().Set("Content-Type", "application/octet-stream")
	r.Response.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	r.Response.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
	r.Response.WriteExit(buf.Bytes())

	return
}
