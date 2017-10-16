package main

import (
	"net/url"
	"testing"
)

func Test_getMergeRequest(t *testing.T) {
	type args struct {
		projectID      int
		mergeRequestID int
	}

	arg := args{
		projectID:      197,
		mergeRequestID: 97,
	}

	glURL, _ = url.Parse("http://git.xxx.cn/")
	*privateToken = "xxx"
	got, err := getMergeRequest(arg.projectID, arg.mergeRequestID)
	if err != nil {
		t.Errorf("getMergeRequest() error = %v", err)
		return
	}

	t.Log(got)
}
