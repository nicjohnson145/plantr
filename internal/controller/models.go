package controller

type Challenge struct {
	ID    string `db:"id"`
	Value string `db:"value"`
}

type DBGithubRelease struct {
	Hash        string `db:"hash"`
	OS          string `db:"os"`
	Arch        string `db:"arch"`
	DownloadURL string `db:"download_url"`
}
