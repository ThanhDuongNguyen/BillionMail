package batch_mail

import (
	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/public"
	"context"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

func (c *ControllerV1) GetTaskRecipientsByStatus(ctx context.Context, req *v1.GetTaskRecipientsByStatusReq) (res *v1.GetTaskRecipientsByStatusRes, err error) {
	g.Log().Printf(ctx, "Getting task recipients by status, TaskId: %d, Type: %s", req.TaskId, req.Type)

	res = &v1.GetTaskRecipientsByStatusRes{}

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

	// Set pagination defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	var total int
	var list []*v1.RecipientItem

	switch req.Type {
	case "delivered":
		total, list, err = queryDeliveredRecipients(ctx, req)
	case "bounced":
		total, list, err = queryBouncedRecipients(ctx, req)
	case "opened":
		total, list, err = queryOpenedRecipients(ctx, req)
	case "clicked":
		total, list, err = queryClickedRecipients(ctx, req)
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

	res.Data.Total = total
	res.Data.List = list
	if res.Data.List == nil {
		res.Data.List = make([]*v1.RecipientItem, 0)
	}

	res.SetSuccess(public.LangCtx(ctx, "Success"))
	return
}

func queryDeliveredRecipients(ctx context.Context, req *v1.GetTaskRecipientsByStatusReq) (int, []*v1.RecipientItem, error) {
	baseQuery := g.DB().Model("mailstat_send_mails sm").
		LeftJoin("mailstat_message_ids mi", "sm.postfix_message_id=mi.postfix_message_id").
		LeftJoin("recipient_info ri", "mi.message_id=ri.message_id").
		Where("ri.task_id", req.TaskId).
		Where("sm.status = 'sent' AND sm.dsn LIKE '2.%'").
		Where("mi.postfix_message_id IS NOT NULL")

	if req.Search != "" {
		baseQuery = baseQuery.Where("sm.recipient LIKE ?", "%"+req.Search+"%")
	}

	count, err := baseQuery.Count()
	if err != nil {
		return 0, nil, err
	}

	var results []struct {
		Recipient    string  `json:"recipient"`
		LogTime      int64   `json:"log_time"`
		MailProvider string  `json:"mail_provider"`
		Delay        float64 `json:"delay"`
	}

	err = baseQuery.Fields("sm.recipient, sm.log_time, sm.mail_provider").
		OrderDesc("sm.log_time_millis").
		Limit((req.Page-1)*req.PageSize, req.PageSize).
		Scan(&results)
	if err != nil {
		return 0, nil, err
	}

	list := make([]*v1.RecipientItem, 0, len(results))
	for _, r := range results {
		list = append(list, &v1.RecipientItem{
			Recipient:    r.Recipient,
			Time:         r.LogTime,
			MailProvider: r.MailProvider,
		})
	}

	return count, list, nil
}

func queryBouncedRecipients(ctx context.Context, req *v1.GetTaskRecipientsByStatusReq) (int, []*v1.RecipientItem, error) {
	baseQuery := g.DB().Model("mailstat_send_mails sm").
		LeftJoin("mailstat_message_ids mi", "sm.postfix_message_id=mi.postfix_message_id").
		LeftJoin("recipient_info ri", "mi.message_id=ri.message_id").
		Where("ri.task_id", req.TaskId).
		Where("sm.status", "bounced").
		Where("mi.postfix_message_id IS NOT NULL")

	if req.Search != "" {
		baseQuery = baseQuery.Where("sm.recipient LIKE ?", "%"+req.Search+"%")
	}

	count, err := baseQuery.Count()
	if err != nil {
		return 0, nil, err
	}

	var results []struct {
		Recipient    string `json:"recipient"`
		LogTime      int64  `json:"log_time"`
		MailProvider string `json:"mail_provider"`
	}

	err = baseQuery.Fields("sm.recipient, sm.log_time, sm.mail_provider").
		OrderDesc("sm.log_time_millis").
		Limit((req.Page-1)*req.PageSize, req.PageSize).
		Scan(&results)
	if err != nil {
		return 0, nil, err
	}

	list := make([]*v1.RecipientItem, 0, len(results))
	for _, r := range results {
		list = append(list, &v1.RecipientItem{
			Recipient:    r.Recipient,
			Time:         r.LogTime,
			MailProvider: r.MailProvider,
		})
	}

	return count, list, nil
}

func queryOpenedRecipients(ctx context.Context, req *v1.GetTaskRecipientsByStatusReq) (int, []*v1.RecipientItem, error) {
	// mailstat_opened has campaign_id which maps to task_id
	baseQuery := g.DB().Model("mailstat_opened o").
		Where("o.campaign_id", req.TaskId)

	if req.Search != "" {
		baseQuery = baseQuery.Where("o.recipient LIKE ?", "%"+req.Search+"%")
	}

	count, err := baseQuery.Count()
	if err != nil {
		// Fallback: try linking through message_id
		return queryOpenedRecipientsViaMessageId(ctx, req)
	}

	// Check if we have results using campaign_id
	if count == 0 {
		// Try fallback via message_id link
		fallbackCount, fallbackList, fallbackErr := queryOpenedRecipientsViaMessageId(ctx, req)
		if fallbackErr == nil && fallbackCount > 0 {
			return fallbackCount, fallbackList, nil
		}
	}

	var results []struct {
		Recipient string `json:"recipient"`
		LogTime   int64  `json:"log_time"`
	}

	err = baseQuery.Fields("o.recipient, o.log_time").
		OrderDesc("o.log_time_millis").
		Limit((req.Page-1)*req.PageSize, req.PageSize).
		Scan(&results)
	if err != nil {
		return 0, nil, err
	}

	list := make([]*v1.RecipientItem, 0, len(results))
	seen := make(map[string]bool)
	for _, r := range results {
		if seen[r.Recipient] {
			continue
		}
		seen[r.Recipient] = true
		list = append(list, &v1.RecipientItem{
			Recipient: r.Recipient,
			Time:      r.LogTime,
		})
	}

	return count, list, nil
}

func queryOpenedRecipientsViaMessageId(ctx context.Context, req *v1.GetTaskRecipientsByStatusReq) (int, []*v1.RecipientItem, error) {
	_ = ctx

	baseQuery := g.DB().Model("mailstat_opened o").
		InnerJoin("recipient_info ri", "o.message_id=ri.message_id").
		Where("ri.task_id", req.TaskId)

	if req.Search != "" {
		baseQuery = baseQuery.Where("o.recipient LIKE ?", "%"+req.Search+"%")
	}

	count, err := baseQuery.Count()
	if err != nil {
		return 0, nil, err
	}

	var results []struct {
		Recipient string `json:"recipient"`
		LogTime   int64  `json:"log_time"`
	}

	err = baseQuery.Fields("o.recipient, o.log_time").
		OrderDesc("o.log_time_millis").
		Limit((req.Page-1)*req.PageSize, req.PageSize).
		Scan(&results)
	if err != nil {
		return 0, nil, err
	}

	list := make([]*v1.RecipientItem, 0, len(results))
	for _, r := range results {
		list = append(list, &v1.RecipientItem{
			Recipient: r.Recipient,
			Time:      r.LogTime,
		})
	}

	return count, list, nil
}

func queryClickedRecipients(ctx context.Context, req *v1.GetTaskRecipientsByStatusReq) (int, []*v1.RecipientItem, error) {
	// mailstat_clicked has campaign_id which maps to task_id
	baseQuery := g.DB().Model("mailstat_clicked c").
		Where("c.campaign_id", req.TaskId)

	if req.Search != "" {
		baseQuery = baseQuery.Where("c.recipient LIKE ?", "%"+req.Search+"%")
	}

	count, err := baseQuery.Count()
	if err != nil {
		// Fallback: try linking through message_id
		return queryClickedRecipientsViaMessageId(ctx, req)
	}

	if count == 0 {
		fallbackCount, fallbackList, fallbackErr := queryClickedRecipientsViaMessageId(ctx, req)
		if fallbackErr == nil && fallbackCount > 0 {
			return fallbackCount, fallbackList, nil
		}
	}

	var results []struct {
		Recipient string `json:"recipient"`
		LogTime   int64  `json:"log_time"`
		Url       string `json:"url"`
	}

	err = baseQuery.Fields("c.recipient, c.log_time, c.url").
		OrderDesc("c.log_time_millis").
		Limit((req.Page-1)*req.PageSize, req.PageSize).
		Scan(&results)
	if err != nil {
		return 0, nil, err
	}

	list := make([]*v1.RecipientItem, 0, len(results))
	for _, r := range results {
		list = append(list, &v1.RecipientItem{
			Recipient: r.Recipient,
			Time:      r.LogTime,
			Url:       r.Url,
		})
	}

	return count, list, nil
}

func queryClickedRecipientsViaMessageId(ctx context.Context, req *v1.GetTaskRecipientsByStatusReq) (int, []*v1.RecipientItem, error) {
	_ = ctx

	baseQuery := g.DB().Model("mailstat_clicked c").
		InnerJoin("recipient_info ri", "c.message_id=ri.message_id").
		Where("ri.task_id", req.TaskId)

	if req.Search != "" {
		baseQuery = baseQuery.Where("c.recipient LIKE ?", "%"+req.Search+"%")
	}

	count, err := baseQuery.Count()
	if err != nil {
		return 0, nil, err
	}

	var results []struct {
		Recipient string `json:"recipient"`
		LogTime   int64  `json:"log_time"`
		Url       string `json:"url"`
	}

	err = baseQuery.Fields("c.recipient, c.log_time, c.url").
		OrderDesc("c.log_time_millis").
		Limit((req.Page-1)*req.PageSize, req.PageSize).
		Scan(&results)
	if err != nil {
		return 0, nil, err
	}

	list := make([]*v1.RecipientItem, 0, len(results))
	for _, r := range results {
		list = append(list, &v1.RecipientItem{
			Recipient: r.Recipient,
			Time:      r.LogTime,
			Url:       r.Url,
		})
	}

	return count, list, nil
}
