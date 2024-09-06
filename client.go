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
	"github.com/darwinOrg/go-common/model"
	"github.com/darwinOrg/go-common/utils"
	dglogger "github.com/darwinOrg/go-logger"
	"strings"
)

type OutBoundConfig struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
}

type CreateJobGroupRequest struct {
	InstanceId   string `json:"InstanceId"`
	ScenarioId   string `json:"ScenarioId"`
	JobGroupName string `json:"JobGroupName"`
}

type AssignJobsRequest struct {
	InstanceId string
	JobGroupId string
	Jobs       []*Job
}

type Job struct {
	Contacts []*Contact                            `json:"contacts"`
	Extras   []*model.KeyValuePair[string, string] `json:"extras"`
}

type Contact struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phonenumber"`
	ReferenceId string `json:"referenceId"`
	Honorific   string `json:"honorific,omitempty"`
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

func CreateJobGroup(ctx *dgctx.DgContext, req *CreateJobGroupRequest) (string, error) {
	resp, err := obClient.CreateJobGroup(&outboundbot20191226.CreateJobGroupRequest{
		InstanceId:   &req.InstanceId,
		ScenarioId:   &req.ScenarioId,
		JobGroupName: &req.JobGroupName,
	})
	if err != nil {
		recommend := extractRecommend(err)
		dglogger.Errorf(ctx, "outbound create job group error | request: %+v | err: %v | recommend: %s", req, err, recommend)
		return "", err
	}
	if resp == nil {
		return "", dgerr.SYSTEM_ERROR
	}
	if *resp.StatusCode != 200 {
		dglogger.Errorf(ctx, "outbound create job group error | request: %+v | response: %+v", req, resp)
		return "", dgerr.NewDgError(int(*resp.StatusCode), *resp.Body.Message)
	}

	return *resp.Body.JobGroup.JobGroupId, nil
}

func AssignJobs(ctx *dgctx.DgContext, req *AssignJobsRequest) ([]string, error) {
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
		dglogger.Errorf(ctx, "outbound assign jobs error | request: %+v | err: %v | recommend: %s", req, err, recommend)
		return nil, err
	}
	if resp == nil {
		return nil, dgerr.SYSTEM_ERROR
	}
	if *resp.StatusCode != 200 {
		dglogger.Errorf(ctx, "outbound assign jobs error | request: %+v | response: %+v", ajr, resp)
		return nil, dgerr.NewDgError(int(*resp.StatusCode), *resp.Body.Message)
	}

	jobIds := dgcoll.MapToList(resp.Body.JobsId, func(jobId *string) string { return tea.StringValue(jobId) })
	return jobIds, nil
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
