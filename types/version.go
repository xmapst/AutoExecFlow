package types

type Version struct {
	BuildTime string      `json:"build_time" yaml:"BuildTime"`
	Version   string      `json:"version" yaml:"Version"`
	Git       VersionGit  `json:"git" yaml:"Git"`
	Go        VersionGO   `json:"go" yaml:"Go"`
	User      VersionUser `json:"user" yaml:"User"`
}

type VersionGit struct {
	Branch string `json:"branch" yaml:"Branch"`
	Commit string `json:"commit" yaml:"Commit"`
	URL    string `json:"url" yaml:"URL"`
}

type VersionGO struct {
	Arch    string `json:"arch" yaml:"Arch"`
	OS      string `json:"os" yaml:"OS"`
	Version string `json:"version" yaml:"Version"`
}

type VersionUser struct {
	Email string `json:"email" yaml:"Email"`
	Name  string `json:"name" yaml:"Name"`
}
