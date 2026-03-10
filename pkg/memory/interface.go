package memory

type Memory interface {
    Read(key string) (any, error)
    Write(key string, value any) error
    Query(prefix string) (map[string]any, error)
}

