package main

import "time"

// MergeRequest represents a GitLab merge request.
//
// GitLab API docs: https://docs.gitlab.com/ce/api/merge_requests.html
// Golang API: https://github.com/xanzy/go-gitlab/blob/master/merge_requests.go
type MergeRequest struct {
	ID             int    `json:"id"`
	IID            int    `json:"iid"`
	ProjectID      int    `json:"project_id"`
	Title          string `json:"title"`
	Description    string `json:"description"`
	WorkInProgress bool   `json:"work_in_progress"`
	State          string `json:"state"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
	TargetBranch   string `json:"target_branch"`
	SourceBranch   string `json:"source_branch"`
	Upvotes        int    `json:"upvotes"`
	Downvotes      int    `json:"downvotes"`
	Author         struct {
		Name      string `json:"name"`
		Username  string `json:"username"`
		ID        int    `json:"id"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
	} `json:"author"`
	Assignee struct {
		Name      string `json:"name"`
		Username  string `json:"username"`
		ID        int    `json:"id"`
		State     string `json:"state"`
		AvatarURL string `json:"avatar_url"`
	} `json:"assignee"`
	SourceProjectID int      `json:"source_project_id"`
	TargetProjectID int      `json:"target_project_id"`
	Labels          []string `json:"labels"`
	Milestone       struct {
		ID          int        `json:"id"`
		Iid         int        `json:"iid"`
		ProjectID   int        `json:"project_id"`
		Title       string     `json:"title"`
		Description string     `json:"description"`
		State       string     `json:"state"`
		CreatedAt   *time.Time `json:"created_at"`
		UpdatedAt   *time.Time `json:"updated_at"`
		DueDate     string     `json:"due_date"`
	} `json:"milestone"`
	MergeWhenPipelineSucceeds bool   `json:"merge_when_pipeline_succeeds"`
	MergeStatus               string `json:"merge_status"`
	SHA                       string `json:"sha"`
	Subscribed                bool   `json:"subscribed"`
	UserNotesCount            int    `json:"user_notes_count"`
	SouldRemoveSourceBranch   bool   `json:"should_remove_source_branch"`
	ForceRemoveSourceBranch   bool   `json:"force_remove_source_branch"`
	Changes                   []struct {
		OldPath     string `json:"old_path"`
		NewPath     string `json:"new_path"`
		AMode       string `json:"a_mode"`
		BMode       string `json:"b_mode"`
		Diff        string `json:"diff"`
		NewFile     bool   `json:"new_file"`
		RenamedFile bool   `json:"renamed_file"`
		DeletedFile bool   `json:"deleted_file"`
	} `json:"changes"`
	WebURL string `json:"web_url"`
}

// Comment represents gitlab comment events
type Comment struct {
	ObjectKind string `json:"object_kind"`
	User       struct {
		Name      string `json:"name"`
		Username  string `json:"username"`
		AvatarURL string `json:"avatar_url"`
	} `json:"user"`
	ProjectID int `json:"project_id"`
	Project   struct {
		Name              string      `json:"name"`
		Description       string      `json:"description"`
		WebURL            string      `json:"web_url"`
		AvatarURL         interface{} `json:"avatar_url"`
		GitSSHURL         string      `json:"git_ssh_url"`
		GitHTTPURL        string      `json:"git_http_url"`
		Namespace         string      `json:"namespace"`
		VisibilityLevel   int         `json:"visibility_level"`
		PathWithNamespace string      `json:"path_with_namespace"`
		DefaultBranch     string      `json:"default_branch"`
		Homepage          string      `json:"homepage"`
		URL               string      `json:"url"`
		SSHURL            string      `json:"ssh_url"`
		HTTPURL           string      `json:"http_url"`
	} `json:"project"`
	ObjectAttributes struct {
		ID                   int         `json:"id"`
		Note                 string      `json:"note"`
		NoteableType         string      `json:"noteable_type"`
		AuthorID             int         `json:"author_id"`
		CreatedAt            string      `json:"created_at"`
		UpdatedAt            string      `json:"updated_at"`
		ProjectID            int         `json:"project_id"`
		Attachment           interface{} `json:"attachment"`
		LineCode             interface{} `json:"line_code"`
		CommitID             string      `json:"commit_id"`
		NoteableID           int         `json:"noteable_id"`
		StDiff               interface{} `json:"st_diff"`
		System               bool        `json:"system"`
		UpdatedByID          interface{} `json:"updated_by_id"`
		Type                 interface{} `json:"type"`
		Position             interface{} `json:"position"`
		OriginalPosition     interface{} `json:"original_position"`
		ResolvedAt           interface{} `json:"resolved_at"`
		ResolvedByID         interface{} `json:"resolved_by_id"`
		DiscussionID         string      `json:"discussion_id"`
		OriginalDiscussionID interface{} `json:"original_discussion_id"`
		URL                  string      `json:"url"`
	} `json:"object_attributes"`
	Repository struct {
		Name        string `json:"name"`
		URL         string `json:"url"`
		Description string `json:"description"`
		Homepage    string `json:"homepage"`
	} `json:"repository"`
	MergeRequest struct {
		ID              int         `json:"id"`
		TargetBranch    string      `json:"target_branch"`
		SourceBranch    string      `json:"source_branch"`
		SourceProjectID int         `json:"source_project_id"`
		AuthorID        int         `json:"author_id"`
		AssigneeID      int         `json:"assignee_id"`
		Title           string      `json:"title"`
		CreatedAt       string      `json:"created_at"`
		UpdatedAt       string      `json:"updated_at"`
		MilestoneID     interface{} `json:"milestone_id"`
		State           string      `json:"state"`
		MergeStatus     string      `json:"merge_status"`
		TargetProjectID int         `json:"target_project_id"`
		Iid             int         `json:"iid"`
		Description     string      `json:"description"`
		Position        int         `json:"position"`
		LockedAt        interface{} `json:"locked_at"`
		UpdatedByID     interface{} `json:"updated_by_id"`
		MergeError      interface{} `json:"merge_error"`
		MergeParams     struct {
			ForceRemoveSourceBranch string `json:"force_remove_source_branch"`
		} `json:"merge_params"`
		MergeWhenBuildSucceeds   bool        `json:"merge_when_build_succeeds"`
		MergeUserID              interface{} `json:"merge_user_id"`
		MergeCommitSha           interface{} `json:"merge_commit_sha"`
		DeletedAt                interface{} `json:"deleted_at"`
		InProgressMergeCommitSha interface{} `json:"in_progress_merge_commit_sha"`
		Source                   struct {
			Name              string `json:"name"`
			Description       string `json:"description"`
			WebURL            string `json:"web_url"`
			AvatarURL         string `json:"avatar_url"`
			GitSSHURL         string `json:"git_ssh_url"`
			GitHTTPURL        string `json:"git_http_url"`
			Namespace         string `json:"namespace"`
			VisibilityLevel   int    `json:"visibility_level"`
			PathWithNamespace string `json:"path_with_namespace"`
			DefaultBranch     string `json:"default_branch"`
			Homepage          string `json:"homepage"`
			URL               string `json:"url"`
			SSHURL            string `json:"ssh_url"`
			HTTPURL           string `json:"http_url"`
		} `json:"source"`
		Target struct {
			Name              string      `json:"name"`
			Description       string      `json:"description"`
			WebURL            string      `json:"web_url"`
			AvatarURL         interface{} `json:"avatar_url"`
			GitSSHURL         string      `json:"git_ssh_url"`
			GitHTTPURL        string      `json:"git_http_url"`
			Namespace         string      `json:"namespace"`
			VisibilityLevel   int         `json:"visibility_level"`
			PathWithNamespace string      `json:"path_with_namespace"`
			DefaultBranch     string      `json:"default_branch"`
			Homepage          string      `json:"homepage"`
			URL               string      `json:"url"`
			SSHURL            string      `json:"ssh_url"`
			HTTPURL           string      `json:"http_url"`
		} `json:"target"`
		LastCommit struct {
			ID        string    `json:"id"`
			Message   string    `json:"message"`
			Timestamp time.Time `json:"timestamp"`
			URL       string    `json:"url"`
			Author    struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			} `json:"author"`
		} `json:"last_commit"`
		WorkInProgress bool `json:"work_in_progress"`
	} `json:"merge_request"`
}
