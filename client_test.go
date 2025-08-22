package dgob_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	dgctx "github.com/darwinOrg/go-common/context"
	"github.com/darwinOrg/go-common/model"
	"github.com/darwinOrg/go-common/utils"
	dglogger "github.com/darwinOrg/go-logger"
	dgob "github.com/darwinOrg/go-outbound"
)

func initClient() {
	err := dgob.InitClient(&dgob.OutBoundConfig{
		AccessKeyId:     os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"),
		AccessKeySecret: os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"),
		Endpoint:        "outboundbot.cn-shanghai.aliyuncs.com",
	})
	if err != nil {
		panic(err)
	}
}

func TestQueryScript(t *testing.T) {
	initClient()
	ctx := dgctx.SimpleDgContext()

	resp, err := dgob.QueryScript(ctx, &dgob.QueryScriptRequest{
		InstanceId: os.Getenv("INSTANCE_ID"),
		ScriptId:   os.Getenv("SCRIPT_ID"),
	})
	if err != nil {
		panic(err)
	}

	dglogger.Infof(ctx, "resp: %s", utils.MustConvertBeanToJsonStringPretty(resp))
}

func TestCreateJobGroup(t *testing.T) {
	initClient()
	ctx := dgctx.SimpleDgContext()

	jobGroupId, err := dgob.CreateJobGroup(ctx, &dgob.CreateJobGroupRequest{
		InstanceId:   os.Getenv("INSTANCE_ID"),
		ScenarioId:   os.Getenv("SCENARIO_ID"),
		JobGroupName: fmt.Sprintf("测试任务名称_%d", time.Now().UnixMilli()),
	})
	if err != nil {
		panic(err)
	}

	dglogger.Infof(ctx, "jobGroupId: %s", jobGroupId)
}

func TestModifyJobGroup(t *testing.T) {
	initClient()
	ctx := dgctx.SimpleDgContext()

	err := dgob.ModifyJobGroup(ctx, &dgob.ModifyJobGroupRequest{
		InstanceId:     os.Getenv("INSTANCE_ID"),
		JobGroupId:     os.Getenv("JOB_GROUP_ID"),
		JobGroupName:   "面试通知_不带岗位v1",
		JobGroupStatus: "Draft",
		MinConcurrency: 1,
	})
	if err != nil {
		panic(err)
	}
}

func TestQueryJobGroup(t *testing.T) {
	initClient()
	ctx := dgctx.SimpleDgContext()

	resp, err := dgob.QueryJobGroup(ctx, &dgob.QueryJobGroupRequest{
		InstanceId: os.Getenv("INSTANCE_ID"),
		JobGroupId: os.Getenv("JOB_GROUP_ID"),
	})
	if err != nil {
		panic(err)
	}

	dglogger.Infof(ctx, "outbound query job group resp: %s", utils.MustConvertBeanToJsonStringPretty(resp))
}

func TestDeleteJobGroup(t *testing.T) {
	initClient()
	ctx := dgctx.SimpleDgContext()

	err := dgob.DeleteJobGroup(ctx, &dgob.QueryJobGroupRequest{
		InstanceId: os.Getenv("INSTANCE_ID"),
		JobGroupId: os.Getenv("JOB_GROUP_ID"),
	})
	if err != nil {
		panic(err)
	}
}

func TestAssignJobs(t *testing.T) {
	initClient()
	ctx := dgctx.SimpleDgContext()

	jobIds, err := dgob.AssignJobs(ctx, &dgob.AssignJobsRequest{
		InstanceId: os.Getenv("INSTANCE_ID"),
		JobGroupId: os.Getenv("JOB_GROUP_ID"),
		Jobs: []*dgob.Job{
			{
				Contacts: []*dgob.Contact{
					{
						Name:        "飘歌",
						PhoneNumber: "15901431753",
						ReferenceId: "01",
					},
				},
				Extras: []*model.KeyValuePair[string, string]{
					{
						Key:   "companyName",
						Value: "腾讯",
					},
					{
						Key:   "expiredAt",
						Value: time.Now().Add(time.Hour * 24 * 3).Format("2006-01-02 15:04:05"),
					},
					{
						Key:   "jobTitle",
						Value: "架构师",
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	dglogger.Infof(ctx, "jobIds: %v", utils.MustConvertBeanToJsonString(jobIds))
}

func TestQueryJobWithResult(t *testing.T) {
	initClient()
	ctx := dgctx.SimpleDgContext()

	resp, err := dgob.QueryJobWithResult(ctx, &dgob.QueryJobWithResultRequest{
		InstanceId: os.Getenv("INSTANCE_ID"),
		JobGroupId: os.Getenv("JOB_GROUP_ID"),
		JobId:      os.Getenv("JOB_ID"),
	})
	if err != nil {
		panic(err)
	}

	dglogger.Infof(ctx, "resp: %v", utils.MustConvertBeanToJsonString(resp))
}
