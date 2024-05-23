package types

type Version struct {
	BuildTime string      `json:"build_time"`
	Version   string      `json:"version"`
	Git       VersionGit  `json:"git"`
	Go        VersionGO   `json:"go"`
	User      VersionUser `json:"user"`
}

type VersionGit struct {
	Branch string `json:"branch"`
	Commit string `json:"commit"`
	URL    string `json:"url"`
}

type VersionGO struct {
	Arch    string `json:"arch"`
	OS      string `json:"os"`
	Version string `json:"version"`
}

type VersionUser struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
