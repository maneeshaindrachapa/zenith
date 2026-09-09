package ports

// Encoder converts configuration data into a format-specific byte payload.
type Encoder interface {
	Encode(map[string]any) ([]byte, error)
}

// Decoder parses a format-specific byte payload into configuration data.
type Decoder interface {
	Decode([]byte, *map[string]any) error
}

// Codec combines encoding and decoding for a single data format.
type Codec interface {
	Encoder
	Decoder
}

// DecoderRegistry finds a decoder for a named format.
type DecoderRegistry interface {
	DecoderFor(format string) (Decoder, error)
}

// Mapper copies decoded configuration values into an application struct.
type Mapper interface {
	MapToStruct(map[string]any, any) error
}

// FileReader reads configuration bytes from a path.
type FileReader interface {
	ReadFile(path string) ([]byte, error)
}
