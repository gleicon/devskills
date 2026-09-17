package store

import "errors"

// Validate rejects records the store cannot key. System records carry no owner.
func Validate(r Record) error {
	if r.ID == "" {
		return errors.New("record: id required")
	}
	return nil
}
