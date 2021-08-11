package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

var (
	// projects/<id>/merge_requests/<iid>
	mergeRequests = make(map[string]*gitlab.MergeRequest)
	mergeMutex    sync.RWMutex
)

func (s *server) checkLgtm(event interface{}) error {

	comment := event.(*gitlab.MergeCommentEvent)
	bytes, _ := json.Marshal(comment)
	content := string(bytes)
	logrus.WithField("comment", content).Infoln("checkLgtm")

	if comment.ObjectKind != ObjectNote {
		logrus.Infoln("comment.ObjectKind != ObjectNote")
		return nil
	}

	if comment.ObjectAttributes.NoteableType != NoteableTypeMergeRequest {
		logrus.Infoln("comment.ObjectAttributes.NoteableType != NoteableTypeMergeRequest")
		return nil
	}

	if err := s.checkLGTMAuthor(comment); err != nil {
		logrus.Infoln("check reviewer failed:", err.Error())
		return nil
	}

	if !strings.EqualFold(comment.ObjectAttributes.Note, Config.LgtmNote) {
		logrus.Infoln("comment.ObjectAttributes.Note !=", Config.LgtmNote)
		return nil
	}

	var (
		canbeMerged bool
		err         error
	)
	logrus.WithFields(logrus.Fields{
		"user": comment.User.Username,
		"note": comment.ObjectAttributes.Note,
		"MR":   comment.MergeRequest.ID,
	}).Info("comment")

	canbeMerged, err = s.checkLGTMCount(comment)
	if err != nil {
		logrus.WithError(err).Errorln("check LGTM count failed")
		return nil
	}

	if canbeMerged && comment.MergeRequest.MergeStatus == StatusCanbeMerged {
		logrus.WithField("MR", comment.MergeRequest.ID).Info("The MR can be merged.")

		removeSource := true
		squash := true
		s.git.MergeRequests.AcceptMergeRequest(comment.ProjectID, comment.MergeRequest.IID, &gitlab.AcceptMergeRequestOptions{
			ShouldRemoveSourceBranch: &removeSource,
			Squash:                   &squash,
		})
	} else {
		logrus.WithFields(logrus.Fields{
			"MR":          comment.MergeRequest.ID,
			"canbeMerged": canbeMerged,
			"MergeStatus": comment.MergeRequest.MergeStatus,
		}).Info("The MR can not be merged.")
	}

	return nil
}

func (s *server) checkLGTMCount(comment *gitlab.MergeCommentEvent) (bool, error) {
	mergeMutex.Lock()
	defer mergeMutex.Unlock()

	tx, err := s.db.Begin(true)
	if err != nil {
		return false, err
	}
	bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		return false, err
	}
	count := 0
	countKey := []byte(fmt.Sprintf("projects/%d/merge_requests/%d", comment.ProjectID, comment.MergeRequest.IID))
	countByte := bucket.Get(countKey)
	if len(countByte) > 0 {
		count, err = strconv.Atoi(string(countByte))
		if err != nil {
			logrus.WithField("value", string(countByte)).Warnln("wrong count")
			count = 0
			err = nil
		}
	}

	count++

	if err := bucket.Put(countKey, []byte(strconv.Itoa(count))); err != nil {
		return false, err
	}
	checkStatus := count-Config.ValidLGTMCount >= 0

	if err := tx.Commit(); err != nil {
		return checkStatus, err
	}
	logrus.WithFields(logrus.Fields{
		"count": count,
		"MR":    comment.MergeRequest.ID,
	}).Info("MR count")
	return checkStatus, nil
}

func (s *server) checkLGTMAuthor(comment *gitlab.MergeCommentEvent) error {

	var (
		mergeRequest *gitlab.MergeRequest
		ok           bool
		err          error
	)
	mergeRequestURI := fmt.Sprintf("projects/%d/merge_requests/%d", comment.ProjectID, comment.MergeRequest.IID)

	mergeMutex.RLock()
	mergeRequest, ok = mergeRequests[mergeRequestURI]
	mergeMutex.RUnlock()

	if !ok {
		mergeRequest, _, err = s.git.MergeRequests.GetMergeRequest(comment.ProjectID, comment.MergeRequest.IID, &gitlab.GetMergeRequestsOptions{})
		if err != nil {
			return err
		}
		mergeMutex.Lock()
		mergeRequests[mergeRequestURI] = mergeRequest
		mergeMutex.Unlock()
	}

	if comment.User.Username != mergeRequest.Author.Username {
		return nil
	}

	return fmt.Errorf("reviewer can't mergeRequest author: %s", mergeRequest.Author.Username)
}
