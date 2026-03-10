package memory

import "errors"

// Redis adapter placeholder; not implemented in prototype.
type Redis struct{}

func (s Redis) Read(key string) (any, error)  { return nil, errors.New("redis memory not implemented") }
func (s Redis) Write(key string, value any) error { return errors.New("redis memory not implemented") }
func (s Redis) Query(prefix string) (map[string]any, error) { return nil, errors.New("redis memory not implemented") }

