package gitlab

//type RawMergeRequestEvent struct {
//	ObjectKind string `json:"object_kind"`
//	User struct {
//		Name string `json:"name"`
//		Description       string           `json:"description"`
//		WebURL            string           `json:"web_url"`
//		AvatarURL         string           `json:"avatar_url"`
//		SSHURLToRepo      string           `json:"git_ssh_url"`
//		HTTPURLToRepo     string           `json:"git_http_url"`
//		Namespace         string `json:"namespace"`
//		VisibilityLevel uint `json:"visibility_level"`
//		PathWithNamespace string           `json:"path_with_namespace"`
//		DefaultBranch     string           `json:"default_branch"`
//		CIConfigPath string `json:"ci_config_path"`
//		Homepage  string `json:"homepage"`
//		URL string `json:"url"`
//		SSHURL string `json:"ssh_url"`
//		HTTPURL string `json:"http_url"`
//
//	} `json:"user"`
//
//}
//
//{
//	//"object_kind": "merge_request",
//	//"user": {
//	  "name": "Administrator",
//	  "username": "root",
//	  "avatar_url": "http://www.gravatar.com/avatar/e64c7d89f26bd1972efa854d13d7dd61?s=80&d=identicon"
//	},
//	"project": {
//	  "name": "Example",
//	  "description": "",
//	  "web_url": "http://example.com/jsmith/example",
//	  "avatar_url": null,
//	  "git_ssh_url": "git@example.com:jsmith/example.git",
//	  "git_http_url": "http://example.com/jsmith/example.git",
//	  "namespace": "Jsmith",
//	  "visibility_level": 0,
//	  "path_with_namespace": "jsmith/example",
//	  "default_branch": "master",
//	  "ci_config_path": "",
//	  "homepage": "http://example.com/jsmith/example",
//	  "url": "git@example.com:jsmith/example.git",
//	  "ssh_url": "git@example.com:jsmith/example.git",
//	  "http_url": "http://example.com/jsmith/example.git"
//	},
//	"object_attributes": {
//	  "id": 90,
//	  "target_branch": "master",
//	  "source_branch": "ms-viewport",
//	  "source_project_id": 14,
//	  "author_id": 51,
//	  "assignee_id": 6,
//	  "title": "MS-Viewport",
//	  "created_at": "2017-09-20T08:31:45.944Z",
//	  "updated_at": "2017-09-28T12:23:42.365Z",
//	  "milestone_id": null,
//	  "state": "opened",
//	  "merge_status": "unchecked",
//	  "target_project_id": 14,
//	  "iid": 1,
//	  "description": "",
//	  "updated_by_id": 1,
//	  "merge_error": null,
//	  "merge_params": {
//		"force_remove_source_branch": "0"
//	  },
//	  "merge_when_pipeline_succeeds": false,
//	  "merge_user_id": null,
//	  "merge_commit_sha": null,
//	  "deleted_at": null,
//	  "in_progress_merge_commit_sha": null,
//	  "lock_version": 5,
//	  "time_estimate": 0,
//	  "last_edited_at": "2017-09-27T12:43:37.558Z",
//	  "last_edited_by_id": 1,
//	  "head_pipeline_id": 61,
//	  "ref_fetched": true,
//	  "merge_jid": null,
//	  "source": {
//		"name": "Awesome Project",
//		"description": "",
//		"web_url": "http://example.com/awesome_space/awesome_project",
//		"avatar_url": null,
//		"git_ssh_url": "git@example.com:awesome_space/awesome_project.git",
//		"git_http_url": "http://example.com/awesome_space/awesome_project.git",
//		"namespace": "root",
//		"visibility_level": 0,
//		"path_with_namespace": "awesome_space/awesome_project",
//		"default_branch": "master",
//		"ci_config_path": "",
//		"homepage": "http://example.com/awesome_space/awesome_project",
//		"url": "http://example.com/awesome_space/awesome_project.git",
//		"ssh_url": "git@example.com:awesome_space/awesome_project.git",
//		"http_url": "http://example.com/awesome_space/awesome_project.git"
//	  },
//	  "target": {
//		"name": "Awesome Project",
//		"description": "Aut reprehenderit ut est.",
//		"web_url": "http://example.com/awesome_space/awesome_project",
//		"avatar_url": null,
//		"git_ssh_url": "git@example.com:awesome_space/awesome_project.git",
//		"git_http_url": "http://example.com/awesome_space/awesome_project.git",
//		"namespace": "Awesome Space",
//		"visibility_level": 0,
//		"path_with_namespace": "awesome_space/awesome_project",
//		"default_branch": "master",
//		"ci_config_path": "",
//		"homepage": "http://example.com/awesome_space/awesome_project",
//		"url": "http://example.com/awesome_space/awesome_project.git",
//		"ssh_url": "git@example.com:awesome_space/awesome_project.git",
//		"http_url": "http://example.com/awesome_space/awesome_project.git"
//	  },
//	  "last_commit": {
//		"id": "ba3e0d8ff79c80d5b0bbb4f3e2e343e0aaa662b7",
//		"message": "fixed readme",
//		"timestamp": "2017-09-26T16:12:57Z",
//		"url": "http://example.com/awesome_space/awesome_project/commits/da1560886d4f094c3e6c9ef40349f7d38b5d27d7",
//		"author": {
//		  "name": "GitLab dev user",
//		  "email": "gitlabdev@dv6700.(none)"
//		}
//	  },
//	  "work_in_progress": false,
//	  "total_time_spent": 0,
//	  "human_total_time_spent": null,
//	  "human_time_estimate": null
//	},
//	"labels": null,
//	"repository": {
//	  "name": "git-gpg-test",
//	  "url": "git@example.com:awesome_space/awesome_project.git",
//	  "description": "",
//	  "homepage": "http://example.com/awesome_space/awesome_project"
//	}
//  }
