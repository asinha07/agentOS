package memory

import "errors"

// SQLite adapter placeholder; not implemented in prototype.
type SQLite struct{}

func (s SQLite) Read(key string) (any, error)  { return nil, errors.New("sqlite memory not implemented") }
func (s SQLite) Write(key string, value any) error { return errors.New("sqlite memory not implemented") }
func (s SQLite) Query(prefix string) (map[string]any, error) { return nil, errors.New("sqlite memory not implemented") }

