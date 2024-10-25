package storage

type Host struct {
	ID       string `db:"id"`
	Hostname string `db:"hostname"`
	Key      string `db:"key"`
}
