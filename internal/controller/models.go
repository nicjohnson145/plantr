package controller

type Challenge struct {
	ID    string `db:"id"`
	Value string `db:"value"`
}

type GithubRelease struct {
	Repo        string `db:"repo"`
	Tag         string `db:"tag"`
	OS          string `db:"os"`
	Arch        string `db:"arch"`
	DownloadURL string `db:"download_url"`
}
