package main

import (
	"log"
	"testing"

	"github.com/xanzy/go-gitlab"
)

func Test_getMergeRequest(t *testing.T) {
	type args struct {
		projectID      int
		mergeRequestID int
	}

	arg := args{
		projectID:      3,
		mergeRequestID: 1,
	}

	git, err := gitlab.NewClient("zXHTLu1azQ1qxQ3xkXmu", gitlab.WithBaseURL("http://192.168.1.11:8000/api/v4"))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
		return
	}

	got, _, err := git.MergeRequests.GetMergeRequest(arg.projectID, arg.mergeRequestID, &gitlab.GetMergeRequestsOptions{})
	if err != nil {
		t.Errorf("getMergeRequest() error = %v", err)
		return
	}

	t.Log(got)
}
