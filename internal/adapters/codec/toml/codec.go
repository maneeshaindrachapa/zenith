package toml

type Endec struct{}

func (Endec) Encode(v map[string]any) ([]byte, error) {
	return nil, nil
}
