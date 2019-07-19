package releasetracker

type Config struct {
	ReleaseChannel Spec `yaml:"releaseChannel"`
}

type Spec struct {
	VersionsFrom VersionsFrom `yaml:"versionsFrom"`
}

type VersionsFrom struct {
	JSONPath        GetterJSONPath  `yaml:"jsonPath"`
	GitTags         GitTags         `yaml:"gitTags"`
	GitHubReleases  GitHubReleases  `yaml:"githubReleases"`
	DockerImageTags DockerImageTags `yaml:"dockerImageTags"`
}

type GetterJSONPath struct {
	Source      string `yaml:"source"`
	Versions    string `yaml:"versions"`
	Description string `yaml:"description"`
}

type GitTags struct {
	Source string `yaml:"source"`
}

type GitHubReleases struct {
	Host   string `yaml:"host"`
	Source string `yaml:"source"`
}

type DockerImageTags struct {
	Source string `yaml:"source"`
}