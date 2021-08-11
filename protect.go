package main

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

const (
	tmplGroup = `<html>
<head>
<title>All Groups</title>
</head>
<body>
<h1>All Groups</h1>
{{range .}}
<ul>
<li>
<a href="/gitlab/projects?group={{.ID}}">{{.FullName}}</a>
</li>
</ul>
{{end}}
</body>
</html>
`
	tmplProjects = `<html>
<head>
<title>Projects</title>
</head>
<body>
<h1>Projects:</h1>
{{range .}}
<ul>
<li>
{{.NameWithNamespace}}
<a target="_blank" href="/gitlab/protect?project={{.ID}}">保护分支和tag</a>
</li>
</ul>
{{end}}
</body>
</html>
`
)

func (s *server) listProjects(w http.ResponseWriter, r *http.Request) {
	logrus.WithFields(logrus.Fields{
		"method":      r.Method,
		"remote_addr": r.RemoteAddr,
	}).Infoln("access")
	var errRet error
	defer func() {
		if errRet != nil {
			errMsg := fmt.Sprintf("error occurs:%s", errRet.Error())
			logrus.WithError(errRet).Errorln("error response")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, errMsg)
			return
		}
	}()

	listGroups := func() {
		groups, _, err := s.git.Groups.ListGroups(&gitlab.ListGroupsOptions{})
		if err != nil {
			errRet = err
			return
		}
		tmpl, err := template.New("name").Parse(tmplGroup)
		if err != nil {
			errRet = err
			return
		}
		errRet = tmpl.Execute(w, groups)
	}

	listProjects := func(group int) {
		projects, _, err := s.git.Groups.ListGroupProjects(group, &gitlab.ListGroupProjectsOptions{})
		if err != nil {
			errRet = err
			return
		}
		tmpl, err := template.New("name").Parse(tmplProjects)
		if err != nil {
			errRet = err
			return
		}
		errRet = tmpl.Execute(w, projects)
	}

	group := r.URL.Query().Get("group")
	if len(group) == 0 {
		listGroups()
		return
	}

	groupID, err := strconv.Atoi(group)
	if err != nil {
		errRet = err
		return
	}

	listProjects(groupID)
}

func (s *server) protectProject(w http.ResponseWriter, r *http.Request) {
	logrus.WithFields(logrus.Fields{
		"method":      r.Method,
		"remote_addr": r.RemoteAddr,
	}).Infoln("access")
	var errRet error
	defer func() {
		if errRet != nil {
			errMsg := fmt.Sprintf("error occurs:%s", errRet.Error())
			logrus.WithError(errRet).Errorln("error response")
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, errMsg)
			return
		}
		w.Write(RespOK)
	}()

	proj := r.URL.Query().Get("project")
	if len(proj) == 0 {
		errRet = errors.New("project is requried")
		return
	}

	projid, err := strconv.Atoi(proj)
	if err != nil {
		errRet = err
		return
	}

	err = s.protectBranchesAndTags(projid)
	if err != nil {
		errRet = err
		return
	}
}

func (s *server) protectBranchesAndTags(project int) error {

	logrus.WithFields(logrus.Fields{
		"project":  project,
		"branches": Config.ProtectedBranches,
		"tags":     Config.ProtectedTags,
	}).Infoln("to protect branches and tags")

	pushPerm := gitlab.NoPermissions
	mergePerm := gitlab.DeveloperPermissions
	for _, b := range strings.Split(Config.ProtectedBranches, ",") {
		_, _, err := s.git.ProtectedBranches.ProtectRepositoryBranches(project, &gitlab.ProtectRepositoryBranchesOptions{
			Name:             &b,
			PushAccessLevel:  &pushPerm,
			MergeAccessLevel: &mergePerm,
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"project": project,
				"branch":  b,
			}).WithError(err).Println("project branch failed")
		}
	}
	perm := gitlab.DeveloperPermissions
	for _, t := range strings.Split(Config.ProtectedTags, ",") {
		_, _, err := s.git.ProtectedTags.ProtectRepositoryTags(project, &gitlab.ProtectRepositoryTagsOptions{
			Name:              &t,
			CreateAccessLevel: &perm,
		})
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"project": project,
				"tag":     t,
			}).WithError(err).Println("project branch failed")
		}
	}

	return nil
}
