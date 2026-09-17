package store

// Record is the persisted shape; the JSON tags are the on-disk format.
type Record struct {
	ID        string `json:"id"`
	OwnerID   string `json:"owner_id"`
	CreatedAt int64  `json:"created_at"`
}

// Age reports how many seconds have passed since the record was created.
func (r Record) Age() int64 {
	return Now() - r.CreatedAt
}
