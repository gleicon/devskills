package store

// Record is the persisted shape; the JSON tags are the on-disk format.
type Record struct {
	ID        string `json:"id"`
	OwnerID   string `json:"ownerId"`
	CreatedAt int64  `json:"createdAt"`
}

// Age reports how long ago the record was created, in the clock's unit.
func (r Record) Age() int64 {
	return Now() - r.CreatedAt
}
