package models

type Model interface {
    Generate(prompt string) (string, error)
    Stream(prompt string, onToken func(tok string)) error
}

