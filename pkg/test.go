package pkg

import "github.com/joetifa2003/ferr/pkg/nested"

type s struct{}

func test() (Intr3, *int, s, nested.Nested, error) {
	return nil, nil, s{}, nested.Nested{}, nil
}
