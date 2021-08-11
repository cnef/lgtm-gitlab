package main

import (
	"testing"
)

func Test_shouldToApply(t *testing.T) {
	tests := []string{"master", "release", "release-v1.1", "release-tx", "releasexxxx", "relea999"}
	for _, tt := range tests {
		got := shouldToApply(tt)
		t.Log("branch:", tt, got)
	}
}

type test_branch struct {
	branch string
	tag    string
}

func Test_incrSemver(t *testing.T) {

	branches := []test_branch{
		{"master", "v1.0.10"},
		{"master", "v1.0.2"},
		{"master", "v1.1.23"},
		{"release-ali", "release-ali-v1.1.2"},
		{"release-ali", "v1.0.10"},
		{"master", ""},
		{"release-ali", ""},
	}

	for _, b := range branches {

		got, err := incrSemver(b.branch, b.tag)
		if err != nil {
			t.Error("tag: ", b, err)
			return
		}
		t.Log("tag:", b, got)
	}

}
