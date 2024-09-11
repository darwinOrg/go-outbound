package dgob

import (
	"encoding/json"
	"errors"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	outboundbot20191226 "github.com/alibabacloud-go/outboundbot-20191226/client"
	"github.com/alibabacloud-go/tea/tea"
	dgcoll "github.com/darwinOrg/go-common/collection"
	dgctx "github.com/darwinOrg/go-common/context"
	dgerr "github.com/darwinOrg/go-common/enums/error"
	"github.com/darwinOrg/go-common/model"
	"github.com/darwinOrg/go-common/utils"
	dglogger "github.com/darwinOrg/go-logger"
	"strings"
	"time"
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

type QueryJobWithResultRequest struct {
	InstanceId string
	JobGroupId string
	JobId      string
}

type QueryJobWithResultResponse struct {
	JobStatus                string                                `json:"jobStatus"`
	JobStatusName            string                                `json:"jobStatusName"`
	EndReason                string                                `json:"endReason"`
	FailureReason            string                                `json:"failureReason"`
	CallTime                 *time.Time                            `json:"callTime,omitempty"`
	CallDuration             int32                                 `json:"callDuration"`
	CallDurationDisplay      string                                `json:"callDurationDisplay"`
	CallStatus               string                                `json:"callStatus"`
	CallStatusName           string                                `json:"callStatusName"`
	HasAnswered              bool                                  `json:"hasAnswered"`
	HasHangUpByRejection     bool                                  `json:"hasHangUpByRejection"`
	HasReachedEndOfFlow      bool                                  `json:"hasReachedEndOfFlow"`
	HasLastPlaybackCompleted bool                                  `json:"hasLastPlaybackCompleted"`
	Extras                   []*model.KeyValuePair[string, string] `json:"extras"`
	RawResponse              string                                `json:"rawResponse"`
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
		InstanceId:   tea.String(req.InstanceId),
		ScenarioId:   tea.String(req.ScenarioId),
		JobGroupName: tea.String(req.JobGroupName),
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

	resp, err := obClient.AssignJobs(ajr)
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

func QueryJobWithResult(ctx *dgctx.DgContext, req *QueryJobWithResultRequest) (*QueryJobWithResultResponse, error) {
	qjwrr := &outboundbot20191226.QueryJobsWithResultRequest{
		InstanceId: tea.String(req.InstanceId),
		JobGroupId: tea.String(req.JobGroupId),
		QueryText:  tea.String(req.JobId),
		PageNumber: tea.Int32(1),
		PageSize:   tea.Int32(1),
	}

	resp, err := obClient.QueryJobsWithResult(qjwrr)
	if err != nil {
		recommend := extractRecommend(err)
		dglogger.Errorf(ctx, "outbound query jobs with result error | request: %+v | err: %v | recommend: %s", req, err, recommend)
		return nil, err
	}
	if resp == nil {
		return nil, dgerr.SYSTEM_ERROR
	}
	if *resp.StatusCode != 200 {
		dglogger.Errorf(ctx, "outbound query jobs with result error | request: %+v | response: %+v", qjwrr, resp)
		return nil, dgerr.NewDgError(int(*resp.StatusCode), *resp.Body.Message)
	}

	jobs := resp.Body.Jobs.List
	if len(jobs) == 0 {
		return nil, nil
	}
	job := jobs[0]

	jobResp := &QueryJobWithResultResponse{
		JobStatus:     tea.StringValue(job.Status),
		JobStatusName: tea.StringValue(job.StatusName),
		FailureReason: tea.StringValue(job.JobFailureReason),
		RawResponse:   utils.MustConvertBeanToJsonString(resp.Body),
	}

	if job.LatestTask != nil {
		jobResp.EndReason = tea.StringValue(job.LatestTask.TaskEndReason)
		callTime := time.UnixMilli(tea.Int64Value(job.LatestTask.CallTime))
		jobResp.CallTime = &callTime
		jobResp.CallDuration = tea.Int32Value(job.LatestTask.CallDuration)
		jobResp.CallDurationDisplay = tea.StringValue(job.LatestTask.CallDurationDisplay)
		jobResp.CallStatus = tea.StringValue(job.LatestTask.Status)
		jobResp.CallStatusName = tea.StringValue(job.LatestTask.StatusName)
		jobResp.HasAnswered = tea.BoolValue(job.LatestTask.HasAnswered)
		jobResp.HasHangUpByRejection = tea.BoolValue(job.LatestTask.HasHangUpByRejection)
		jobResp.HasReachedEndOfFlow = tea.BoolValue(job.LatestTask.HasReachedEndOfFlow)
		jobResp.HasLastPlaybackCompleted = tea.BoolValue(job.LatestTask.HasLastPlaybackCompleted)

		if len(job.LatestTask.Extras) > 0 {
			jobResp.Extras = dgcoll.MapToList(job.LatestTask.Extras, func(extra *outboundbot20191226.QueryJobsWithResultResponseBodyJobsListLatestTaskExtras) *model.KeyValuePair[string, string] {
				return &model.KeyValuePair[string, string]{
					Key:   tea.StringValue(extra.Key),
					Value: tea.StringValue(extra.Value),
				}
			})
		}
	}

	return jobResp, nil
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
