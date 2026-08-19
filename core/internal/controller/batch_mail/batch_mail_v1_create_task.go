package batch_mail

import (
	"billionmail-core/internal/consts"
	"billionmail-core/internal/service/batch_mail"
	"billionmail-core/internal/service/public"
	"context"
	"regexp"

	"github.com/gogf/gf/v2/errors/gerror"

	"billionmail-core/api/batch_mail/v1"
	"billionmail-core/internal/service/email_template"
)

func (c *ControllerV1) CreateTask(ctx context.Context, req *v1.CreateTaskReq) (res *v1.CreateTaskRes, err error) {
	res = &v1.CreateTaskRes{}

	if err = validateCreateTaskRequest(req); err != nil {
		res.SetError(err)
		return
	}

	// check template
	template, err := email_template.GetTemplate(ctx, req.TemplateId)
	if err != nil {
		res.Code = 500
		res.SetError(gerror.New(public.LangCtx(ctx, "Failed to get template")))
		return
	}
	if template == nil {
		res.Code = 400
		res.SetError(gerror.New(public.LangCtx(ctx, "Template not found")))
		return
	}

	addType := 0
	// create task and import recipients
	res.Data.Id, err = batch_mail.CreateTaskWithRecipients(ctx, req, addType)
	if err != nil {
		res.SetError(err)
		return
	}

	_ = public.WriteLog(ctx, public.LogParams{
		Type: consts.LOGTYPE.Task,
		Log:  "Create task :" + req.Subject + " successfully",
		Data: req,
	})

	res.SetSuccess(public.LangCtx(ctx, "Task created successfully"))
	return
}

func validateCreateTaskRequest(req *v1.CreateTaskReq) error {

	if req.GroupId <= 0 {
		return gerror.New("Must select a contact group")
	}

	if len(req.TagIds) > 0 {
		if req.TagLogic != "" && req.TagLogic != "AND" && req.TagLogic != "OR" {
			return gerror.New("Tag logic must be AND or OR")
		}

		if req.TagLogic == "" {
			req.TagLogic = "AND"
		}
	}

	if err := validateTaskVariables(req.Variables); err != nil {
		return err
	}

	return nil
}

// variableKeyRegex allows only alphanumeric characters and underscores for variable keys.
var variableKeyRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// reservedTaskKeys are built-in task fields that cannot be overridden by custom variables.
var reservedTaskKeys = map[string]struct{}{
	"Id": {}, "TaskName": {}, "Addresser": {}, "Subject": {},
	"FullName": {}, "RecipientCount": {}, "TaskProcess": {}, "Pause": {},
	"TemplateId": {}, "IsRecord": {}, "Unsubscribe": {}, "Threads": {},
	"TrackOpen": {}, "TrackClick": {}, "StartTime": {}, "CreateTime": {},
	"UpdateTime": {}, "Remark": {}, "Active": {},
}

const maxTaskVariables = 20

// validateTaskVariables validates custom task variables.
func validateTaskVariables(variables map[string]string) error {
	if len(variables) == 0 {
		return nil
	}
	if len(variables) > maxTaskVariables {
		return gerror.Newf("Task variables cannot exceed %d entries", maxTaskVariables)
	}
	for key := range variables {
		if !variableKeyRegex.MatchString(key) {
			return gerror.Newf("Variable key %q is invalid: only letters, digits, and underscores are allowed, and must start with a letter or underscore", key)
		}
		if _, reserved := reservedTaskKeys[key]; reserved {
			return gerror.Newf("Variable key %q conflicts with a built-in task field", key)
		}
	}
	return nil
}
