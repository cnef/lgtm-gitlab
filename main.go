package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/boltdb/bolt"
)

var (
	// ErrInvalidRequest ...
	ErrInvalidRequest = errors.New("invalid request body")
	// ErrInvalidContentType ...
	ErrInvalidContentType = errors.New("invalid content type")
	// RespOK ...
	RespOK       = []byte("OK")
	db           *bolt.DB
	buildVersion string
)

const (
	// ObjectNote ...
	ObjectNote = "note"
	// NoteableTypeMergeRequest ...
	NoteableTypeMergeRequest = "MergeRequest"
	// NoteLGTM ...
	NoteLGTM = "LGTM"
	// StatusCanbeMerged ...
	StatusCanbeMerged = "can_be_merged"
	bucketName        = "lgtm"
)

var (
	privateToken   = flag.String("token", "", "gitlab private token which used to accept merge request. can be found in https://your.gitlab.com/profile/account")
	gitlabURL      = flag.String("gitlab_url", "", "e.g. https://your.gitlab.com")
	validLGTMCount = flag.Int("lgtm_count", 2, "lgtm user count")
	lgtmNote       = flag.String("lgtm_note", NoteLGTM, "lgtm note")
	logLevel       = flag.String("log_level", "info", "log level")
	port           = flag.Int("port", 8989, "http listen port")
	dbPath         = flag.String("db_path", "lgtm.data", "bolt db data")
)

var (
	// projects/<id>/merge_requests/<id>
	mergeRequests = make(map[string]*MergeRequest)
	mutex         sync.RWMutex
	mergeMutex    sync.RWMutex

	glURL *url.URL
)

func formatLogLevel(level string) logrus.Level {
	l, err := logrus.ParseLevel(string(level))
	if err != nil {
		l = logrus.InfoLevel
		logrus.Warnf("error parsing level %q: %v, using %q	", level, err, l)
	}

	return l
}

func init() {
	flag.Parse()
	logrus.SetOutput(os.Stderr)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	logrus.SetLevel(formatLogLevel(*logLevel))
	logrus.WithField("buildVersion", buildVersion).Info("build info")
}

func main() {
	if *privateToken == "" {
		logrus.Fatal("private token is required")
	}
	if *gitlabURL == "" {
		logrus.Fatal("gitlab url is required")
	}
	var err error
	db, err = bolt.Open(*dbPath, 0600, nil)
	if err != nil {
		logrus.WithError(err).Fatal("open local db failed")
	}
	defer db.Close()
	parseURL(*gitlabURL)

	http.HandleFunc("/gitlab/hook", LGTMHandler)
	go func() {
		logrus.Infof("Webhook server listen on 0.0.0.0:%d", *port)
		http.ListenAndServe(":"+strconv.Itoa(*port), nil)
	}()

	<-(chan struct{})(nil)
}

func parseURL(urlStr string) {
	var err error
	glURL, err = url.Parse(urlStr)
	if err != nil {
		panic(err.Error())
	}
}

// LGTMHandler ...
func LGTMHandler(w http.ResponseWriter, r *http.Request) {
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
			fmt.Fprintf(w, errMsg)
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

	var comment Comment
	if err := json.NewDecoder(r.Body).Decode(&comment); err != nil {
		errRet = err
		return
	}

	go checkLgtm(comment)
}

func checkLgtm(comment Comment) error {

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

	if err := checkLGTMAuthor(comment); err != nil {
		logrus.Infoln("check reviewer failed:", err.Error())
		return nil
	}

	if strings.ToLower(comment.ObjectAttributes.Note) != strings.ToLower(*lgtmNote) {
		logrus.Infoln("comment.ObjectAttributes.Note != ", *lgtmNote)
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

	canbeMerged, err = checkLGTMCount(comment)

	if err != nil {
		logrus.WithError(err).Errorln("check LGTM count failed")
		return nil
	}
	if canbeMerged && comment.MergeRequest.MergeStatus == StatusCanbeMerged {
		logrus.WithField("MR", comment.MergeRequest.ID).Info("The MR can be merged.")
		acceptMergeRequest(comment.ProjectID, comment.MergeRequest.ID, comment.MergeRequest.MergeParams.ForceRemoveSourceBranch)
	} else {
		logrus.WithFields(logrus.Fields{
			"MR":          comment.MergeRequest.ID,
			"canbeMerged": canbeMerged,
			"MergeStatus": comment.MergeRequest.MergeStatus,
		}).Info("The MR can not be merged.")
	}

	return nil
}

func checkLGTMCount(comment Comment) (bool, error) {
	mutex.Lock()
	defer mutex.Unlock()

	tx, err := db.Begin(true)
	if err != nil {
		return false, err
	}
	bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
	if err != nil {
		return false, err
	}
	count := 0
	countKey := []byte(strconv.Itoa(comment.MergeRequest.ID))
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
	checkStatus := count%(*validLGTMCount) == 0

	if err := tx.Commit(); err != nil {
		return checkStatus, err
	}
	logrus.WithFields(logrus.Fields{
		"count": count,
		"MR":    comment.MergeRequest.ID,
	}).Info("MR count")
	return checkStatus, nil
}

func checkLGTMAuthor(comment Comment) error {

	var (
		mergeRequest *MergeRequest
		ok           bool
		err          error
	)
	mergeRequestURI := fmt.Sprintf("projects/%d/merge_requests/%d", comment.ProjectID, comment.MergeRequest.ID)

	mergeMutex.RLock()
	mergeRequest, ok = mergeRequests[mergeRequestURI]
	mergeMutex.RUnlock()

	if !ok {
		mergeRequest, err = getMergeRequest(comment.ProjectID, comment.MergeRequest.ID)
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
