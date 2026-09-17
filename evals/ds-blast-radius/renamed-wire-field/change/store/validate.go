package store

import "errors"

// Validate rejects records the store cannot key or attribute.
func Validate(r Record) error {
	if r.ID == "" {
		return errors.New("record: id required")
	}
	if r.OwnerID == "" {
		return errors.New("record: owner required")
	}
	return nil
}
