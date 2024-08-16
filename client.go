package dgob

import (
	"encoding/json"
	"errors"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	outboundbot20191226 "github.com/alibabacloud-go/outboundbot-20191226/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	dgcoll "github.com/darwinOrg/go-common/collection"
	dgctx "github.com/darwinOrg/go-common/context"
	dgerr "github.com/darwinOrg/go-common/enums/error"
	"github.com/darwinOrg/go-common/utils"
	dglogger "github.com/darwinOrg/go-logger"
	"strings"
)

type OutBoundConfig struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
}

type StartJobRequest struct {
	InstanceId string
	JobGroupId string
	Jobs       []*Job
	ScenarioId string
	ScriptId   string
}

type AssignJobsRequest struct {
	InstanceId string
	JobGroupId string
	Jobs       []*Job
}

type Job struct {
	Contacts []*Contact      `json:"contacts"`
	Extras   []*KeyValuePair `json:"extras"`
}

type Contact struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phonenumber"`
	ReferenceId string `json:"referenceId"`
	Honorific   string `json:"honorific,omitempty"`
}

type KeyValuePair struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var obClient *outboundbot20191226.Client

func InitClient(cfg *OutBoundConfig) error {
	config := &openapi.Config{
		AccessKeyId:     tea.String(cfg.AccessKeyId),
		AccessKeySecret: tea.String(cfg.AccessKeySecret),
	}
	// Endpoint 请参考 https://api.aliyun.com/product/OutboundBot
	config.Endpoint = tea.String(cfg.Endpoint)
	var err error
	obClient, err = outboundbot20191226.NewClient(config)
	return err
}

func StartJob(ctx *dgctx.DgContext, req *StartJobRequest) error {
	sjr := &outboundbot20191226.StartJobRequest{
		InstanceId: tea.String(req.InstanceId),
		JobGroupId: tea.String(req.JobGroupId),
		JobJson:    tea.String(utils.MustConvertBeanToJsonString(req.Jobs)),
	}
	if req.ScenarioId != "" {
		sjr.ScenarioId = tea.String(req.ScenarioId)
	}
	if req.ScriptId != "" {
		sjr.ScriptId = tea.String(req.ScriptId)
	}

	resp, err := obClient.StartJob(sjr)
	if err != nil {
		recommend := extractRecommend(err)
		dglogger.Errorf(ctx, "outbound start job error | request: %+v | err: %v | recommend: %s", sjr, err, recommend)
		return err
	}
	if resp == nil {
		return dgerr.SYSTEM_ERROR
	}
	if *resp.StatusCode != 200 {
		dglogger.Errorf(ctx, "outbound start job error | request: %+v | response: %+v", sjr, resp)
		return dgerr.NewDgError(int(*resp.StatusCode), *resp.Body.Message)
	}

	return nil
}

func AssignJobs(ctx *dgctx.DgContext, req *AssignJobsRequest) error {
	ajr := &outboundbot20191226.AssignJobsRequest{
		InstanceId: tea.String(req.InstanceId),
		JobGroupId: tea.String(req.JobGroupId),
		JobsJson: dgcoll.MapToList(req.Jobs, func(job *Job) *string {
			return tea.String(utils.MustConvertBeanToJsonString(job))
		}),
	}

	resp, err := obClient.AssignJobsWithOptions(ajr, &util.RuntimeOptions{})
	if err != nil {
		recommend := extractRecommend(err)
		dglogger.Errorf(ctx, "outbound assign jobs error | request: %+v | err: %v | recommend: %s", ajr, err, recommend)
		return err
	}
	if resp == nil {
		return dgerr.SYSTEM_ERROR
	}
	if *resp.StatusCode != 200 {
		dglogger.Errorf(ctx, "outbound assign jobs error | request: %+v | response: %+v", ajr, resp)
		return dgerr.NewDgError(int(*resp.StatusCode), *resp.Body.Message)
	}

	return nil
}

func extractRecommend(err error) string {
	var tse = &tea.SDKError{}
	var _t *tea.SDKError
	if errors.As(err, &_t) {
		tse = _t
	}

	d := json.NewDecoder(strings.NewReader(tea.StringValue(tse.Data)))
	var data any
	_ = d.Decode(&data)

	if m, ok := data.(map[string]any); ok {
		if recommend, ok1 := m["Recommend"]; ok1 {
			if strRecommend, ok2 := recommend.(string); ok2 {
				return strRecommend
			}
		}
	}

	return ""
}
