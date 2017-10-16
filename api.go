package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Sirupsen/logrus"
)

// V4 接口 iid 参数用起来不方便

func getMergeRequest(projectID int, mergeRequestID int) (*MergeRequest, error) {
	glURL.Path = fmt.Sprintf("/api/v3/projects/%d/merge_requests/%d", projectID, mergeRequestID)
	req, err := http.NewRequest("GET", glURL.String(), nil)
	if err != nil {
		logrus.WithError(err).Errorln("http NewRequest failed")
		return nil, err
	}
	req.Header.Set("Conntent-Type", "application/json")
	req.Header.Set("PRIVATE-TOKEN", *privateToken) // authenticate

	logrus.WithField("url", glURL.String()).Info("get merge request")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.WithError(err).Errorln("execute request failed")
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"http_code":   resp.StatusCode,
			"http_status": resp.Status,
		}).Errorln("get merge request failed")

		return nil, fmt.Errorf("get merge request error: %s", resp.Status)
	}

	var mr MergeRequest
	if err := json.NewDecoder(resp.Body).Decode(&mr); err != nil {
		return nil, err
	}

	return &mr, nil
}

func acceptMergeRequest(projectID int, mergeRequestID int, shouldRemoveSourceBranch string) {
	params := map[string]string{
		"should_remove_source_branch": shouldRemoveSourceBranch,
	}
	bodyBytes, err := json.Marshal(params)
	if err != nil {
		logrus.WithError(err).Errorln("json marshal failed")
		return
	}

	glURL.Path = fmt.Sprintf("/api/v3/projects/%d/merge_requests/%d/merge", projectID, mergeRequestID)
	req, err := http.NewRequest("PUT", glURL.String(), bytes.NewReader(bodyBytes))
	if err != nil {
		logrus.WithError(err).Errorln("http NewRequest failed")
		return
	}
	req.Header.Set("Conntent-Type", "application/json")
	// authenticate
	req.Header.Set("PRIVATE-TOKEN", *privateToken) // my private token

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.WithError(err).Errorln("execute request failed")
		return
	}

	switch resp.StatusCode {
	// 200
	case http.StatusOK:
		logrus.Info("accept merge request successfully")
	// 405
	case http.StatusMethodNotAllowed:
		logrus.Warnln("it has some conflicts and can not be merged")
	// 406
	case http.StatusNotAcceptable:
		logrus.Warnln("merge request is already merged or closed")
	default:
		logrus.WithFields(logrus.Fields{
			"http_code":   resp.StatusCode,
			"http_status": resp.Status,
		}).Errorln("accept merge failed")
	}
}
