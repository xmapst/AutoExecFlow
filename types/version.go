package types

type SVersion struct {
	BuildTime string       `json:"build_time" yaml:"BuildTime"`
	Version   string       `json:"version" yaml:"Version"`
	Git       SVersionGit  `json:"git" yaml:"Git"`
	Go        SVersionGO   `json:"go" yaml:"Go"`
	User      SVersionUser `json:"user" yaml:"User"`
}

type SVersionGit struct {
	Branch string `json:"branch" yaml:"Branch"`
	Commit string `json:"commit" yaml:"Commit"`
	URL    string `json:"url" yaml:"URL"`
}

type SVersionGO struct {
	Arch    string `json:"arch" yaml:"Arch"`
	OS      string `json:"os" yaml:"OS"`
	Version string `json:"version" yaml:"Version"`
}

type SVersionUser struct {
	Email string `json:"email" yaml:"Email"`
	Name  string `json:"name" yaml:"Name"`
}
