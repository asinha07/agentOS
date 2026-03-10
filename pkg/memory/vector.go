package memory

import "errors"

// Vector adapter placeholder; not implemented in prototype.
type Vector struct{}

func (s Vector) Read(key string) (any, error)  { return nil, errors.New("vector memory not implemented") }
func (s Vector) Write(key string, value any) error { return errors.New("vector memory not implemented") }
func (s Vector) Query(prefix string) (map[string]any, error) { return nil, errors.New("vector memory not implemented") }

