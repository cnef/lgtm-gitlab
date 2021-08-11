package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/boltdb/bolt"
	"github.com/jinzhu/configor"
	"github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

var (
	// ErrInvalidRequest ...
	ErrInvalidRequest = errors.New("invalid request body")
	// ErrInvalidContentType ...
	ErrInvalidContentType = errors.New("invalid content type")
	// RespOK ...
	RespOK       = []byte("OK")
	buildVersion string
)

const (
	// ObjectNote ...
	ObjectNote               = "note"
	NoteableTypeMergeRequest = "MergeRequest"
	MergeActionMerge         = "merge"
	StatusCanbeMerged        = "can_be_merged"
	bucketName               = "lgtm"
)

var Config struct {
	PrivateToken      string `env:"LGTM_TOKEN" default:""` // gitlab private token which used to accept merge request. can be found in https://your.gitlab.com/profile/account
	GitlabURL         string `env:"LGTM_GITLAB_URL" default:"https://your.gitlab.com"`
	ValidLGTMCount    int    `env:"LGTM_COUNT" default:"1"`
	LgtmNote          string `env:"LGTM_NOTE" default:"lgtm"`
	LogLevel          string `env:"LGTM_LOG_LEVEL" default:"info"`
	Port              int    `env:"LGTM_PORT" default:"8989"`
	DbPath            string `env:"LGTM_DB_PATH" default:"/var/lib/lgtm/lgtm.data"`
	ProtectedBranches string `env:"LGTM_PROTECTED_BRANCHES" default:"master,release-*"`
	ProtectedTags     string `env:"LGTM_PROTECTED_TAGS" default:"v*,release-*"`
	DefaultTag        string `env:"LGTM_DEFAULT_TAG" default:"v1.0.0"`
	MainBranch        string `env:"LGTM_MAIN_BRANCH" default:"master"`
}

func formatLogLevel(level string) logrus.Level {
	l, err := logrus.ParseLevel(string(level))
	if err != nil {
		l = logrus.InfoLevel
		logrus.Warnf("error parsing level %q: %v, using %q	", level, err, l)
	}
	return l
}

func init() {
	if err := configor.Load(&Config); err != nil {
		panic(err)
	}

	logrus.SetOutput(os.Stderr)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.SetLevel(formatLogLevel(Config.LogLevel))
	logrus.WithField("buildVersion", buildVersion).Info("build info")
}

func main() {
	if Config.PrivateToken == "" {
		logrus.Fatal("private token is required")
	}
	if Config.GitlabURL == "" {
		logrus.Fatal("gitlab url is required")
	}
	var err error
	db, err := bolt.Open(Config.DbPath, 0600, nil)
	if err != nil {
		logrus.WithError(err).Fatal("open local db failed")
	}
	defer db.Close()

	_, err = url.Parse(Config.GitlabURL)
	if err != nil {
		panic(err.Error())
	}

	git, err := gitlab.NewClient(Config.PrivateToken, gitlab.WithBaseURL(fmt.Sprintf("%s/api/v4", Config.GitlabURL)))
	if err != nil {
		logrus.Fatalf("Failed to create client: %v", err)
		return
	}

	svc := &server{git: git, db: db}

	http.HandleFunc("/gitlab/webhook", svc.webhookHandler)
	http.HandleFunc("/gitlab/protect", svc.protectProject)
	http.HandleFunc("/gitlab/projects", svc.listProjects)
	logrus.Infof("Webhook server listen on 0.0.0.0:%d", Config.Port)
	http.ListenAndServe(fmt.Sprintf(":%d", Config.Port), nil)
}

type server struct {
	git *gitlab.Client
	db  *bolt.DB
}

func (s *server) webhookHandler(w http.ResponseWriter, r *http.Request) {
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

	if r.Header.Get("Content-Type") != "application/json" {
		errRet = ErrInvalidContentType
		return
	}
	if r.Method != "POST" {
		errRet = ErrInvalidRequest
		return
	}
	if r.Body == nil {
		errRet = ErrInvalidRequest
		return
	}

	payload, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errRet = ErrInvalidRequest
		return
	}

	event, err := gitlab.ParseHook(gitlab.HookEventType(r), payload)
	if err != nil {
		errRet = ErrInvalidRequest
		return
	}

	switch event := event.(type) {
	case *gitlab.MergeCommentEvent:
		go s.checkLgtm(event)
	case *gitlab.MergeEvent:
		go s.autoTags(event)
	default:
	}

}
