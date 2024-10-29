package storage

type Challenge struct {
	ID    string `db:"id"`
	Value string `db:"value"`
}
