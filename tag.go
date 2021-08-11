package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/blang/semver/v4"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

var (
	// projects/<id>/merge_requests/<iid>
	tagedRequests = make(map[string]struct{})
	tagMutex      sync.Mutex
)

func shouldToApply(branch string) bool {
	for _, b := range strings.Split(Config.ProtectedBranches, ",") {
		if b == branch {
			return true
		}
		if strings.HasSuffix(b, "*") &&
			strings.HasPrefix(branch, b[:len(b)-1]) {
			return true
		}
	}
	return false
}

func (s *server) autoTags(event interface{}) error {

	merge := event.(*gitlab.MergeEvent)
	bytes, _ := json.Marshal(merge)
	content := string(bytes)
	logrus.WithField("merge", content).Infoln("autoTags")

	if merge.ObjectAttributes.Action != MergeActionMerge {
		logrus.WithFields(logrus.Fields{
			"project": merge.Project.ID,
			"MR":      merge.ObjectAttributes.ID,
		}).Infoln("merge.ObjectAttributes.Action != MergeActionMerge")
		return nil
	}

	var targetBranch = merge.ObjectAttributes.TargetBranch
	if !shouldToApply(targetBranch) {
		logrus.WithFields(logrus.Fields{
			"project":      merge.Project.ID,
			"MR":           merge.ObjectAttributes.ID,
			"targetBranch": targetBranch,
		}).Infoln("merge.ObjectAttributes.TargetBranch don't need to apply", targetBranch)
		return nil
	}

	tagRequestURI := fmt.Sprintf("projects/%d/merge_requests/%d", merge.Project.ID, merge.ObjectAttributes.IID)
	tagMutex.Lock()
	defer tagMutex.Unlock()
	if _, ok := tagedRequests[tagRequestURI]; ok {
		logrus.WithFields(logrus.Fields{
			"project": merge.Project.ID,
			"MR":      merge.ObjectAttributes.ID,
		}).Info("this merge is taged already")
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"project": merge.Project.ID,
		"MR":      merge.ObjectAttributes.ID,
	}).Info("merge")

	search := "^v"
	if targetBranch != Config.MainBranch {
		search = fmt.Sprintf("^%s-v", targetBranch)
	}
	orderBy := "name"
	tags, _, err := s.git.Tags.ListTags(merge.Project.ID, &gitlab.ListTagsOptions{
		OrderBy: &orderBy,
		Search:  &search,
	})
	if err != nil {
		logrus.WithError(err).Errorln("ListTags failed")
		return nil
	}

	var currentTag string
	if len(tags) > 0 {
		currentTag = tags[0].Name
	}

	nextTag, err := incrSemver(targetBranch, currentTag)
	if err != nil {
		logrus.WithError(err).Errorln("Incr tag failed", "currentTag")
		return nil
	}

	logrus.WithFields(logrus.Fields{
		"user":    merge.User.Username,
		"project": merge.Project.ID,
		"MR":      merge.ObjectAttributes.ID,
		"nextTag": nextTag,
	})

	tag, _, err := s.git.Tags.CreateTag(merge.Project.ID, &gitlab.CreateTagOptions{
		Ref:     &targetBranch,
		TagName: &nextTag,
	})

	if err != nil {
		logrus.WithError(err).Errorln("Create tag failed", "currentTag", tags[0].Name, "branch", targetBranch)
		return nil
	} else {
		logrus.WithFields(logrus.Fields{
			"project": merge.Project.ID,
			"branch":  targetBranch,
			"Tag":     tag.Name,
		}).Info("Tag created")
	}

	tagedRequests[tagRequestURI] = struct{}{}

	return nil
}

func incrSemver(branch, current string) (string, error) {

	var prefix string
	if branch != Config.MainBranch {
		prefix = branch + "-"
		current = strings.TrimPrefix(current, prefix)
	}
	if len(current) == 0 {
		return prefix + Config.DefaultTag, nil
	}
	ver, err := semver.ParseTolerant(current)
	if err != nil {
		return "", err
	}
	err = ver.IncrementPatch()
	if err != nil {
		return "", err
	}

	return prefix + "v" + ver.String(), nil
}
