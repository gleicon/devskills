package store

import (
	"encoding/json"
	"fmt"
	"os"
)

// Load reads every record from path and refuses the file if any record is invalid.
func Load(path string) ([]Record, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	var records []Record
	if err := json.Unmarshal(raw, &records); err != nil {
		return nil, fmt.Errorf("load: %w", err)
	}
	for _, r := range records {
		if err := Validate(r); err != nil {
			return nil, fmt.Errorf("load %s: %w", r.ID, err)
		}
	}
	return records, nil
}
