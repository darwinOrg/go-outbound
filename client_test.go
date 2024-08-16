package dgob_test

import (
	dgctx "github.com/darwinOrg/go-common/context"
	dgob "github.com/darwinOrg/go-outbound"
	"os"
	"testing"
	"time"
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

func TestAssignJobs(t *testing.T) {
	initClient()
	ctx := &dgctx.DgContext{TraceId: "123"}

	err := dgob.AssignJobs(ctx, &dgob.AssignJobsRequest{
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
				Extras: []*dgob.KeyValuePair{
					{
						Key:   "companyName",
						Value: "腾讯",
					},
					{
						Key:   "expiredAt",
						Value: time.Now().Add(time.Hour * 24 * 3).Format("2006-01-02 15:04:05"),
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}
}
