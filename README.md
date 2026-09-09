# Zenith

Zenith is a small Go configuration loader. It reads a config file, selects the correct decoder from the file extension, decodes the file into key-value data, and maps those values into a struct.

The package currently supports JSON and dotenv files through the default registry.

## Installation

```sh
go get github.com/maneeshaindrachapa/zenith
```

## Quick Usage

Create a config file:

```json
{
  "name": "Zenith",
  "port": 8080,
  "enabled": true,
  "timeout": "5s"
}
```

Load it into a struct:

```go
package main

import (
	"log"
	"time"

	"github.com/maneeshaindrachapa/zenith"
)

type Config struct {
	Name    string        `json:"name"`
	Port    int           `json:"port"`
	Enabled bool          `json:"enabled"`
	Timeout time.Duration `json:"timeout"`
}

func main() {
	var cfg Config
	if err := zenith.Load("config.json", &cfg); err != nil {
		log.Fatal(err)
	}

	log.Printf("%+v", cfg)
}
```

You can also decode bytes directly when you already know the format:

```go
var cfg Config
err := zenith.Decode([]byte(`{"name":"Zenith","port":8080}`), "json", &cfg)
```

## Dotenv Usage

```sh
NAME="Zenith"
PORT="8080"
ENABLED="true"
TIMEOUT="250ms"
```

```go
type Config struct {
	Name    string        `env:"NAME"`
	Port    int           `env:"PORT"`
	Enabled bool          `env:"ENABLED"`
	Timeout time.Duration `env:"TIMEOUT"`
}

var cfg Config
err := zenith.Load(".env", &cfg)
```

## Struct Mapping

Zenith maps decoded values into exported struct fields.

Supported tags:

- `zenith`
- `json`
- `env`
- `mapstructure`

Supported value types:

- `string`
- `bool`
- signed and unsigned integers
- floats
- `time.Duration`
- slices
- nested structs
- pointer fields
- fields implementing `encoding.TextUnmarshaler`

Field matching is normalized, so names such as `PORT`, `port`, `max_connections`, `max-connections`, and `MaxConnections` can map to the same field.

Fields tagged with `json:"-"`, `env:"-"`, `zenith:"-"`, or `mapstructure:"-"` are ignored.

## Supported Formats

Default registered formats:

- `json`
- `dotenv`
- `env`

TOML and YAML adapter directories exist, but they are not complete or registered yet.

You can inspect the current registry:

```go
formats := encoding.RegisteredFormats()
```

## Registering a Custom Codec

A codec implements both `Encode` and `Decode`:

```go
type Codec interface {
	Encode(map[string]any) ([]byte, error)
	Decode([]byte, *map[string]any) error
}
```

Register it by format name:

```go
import "github.com/maneeshaindrachapa/zenith/encoding"

err := encoding.Register("custom", MyCodec{})
```

After registration, `zenith.Load("config.custom", &cfg)` and `zenith.Decode(data, "custom", &cfg)` can use that codec.

## Architecture

Zenith follows a hexagonal architecture style, also called ports and adapters.

The public API is intentionally thin:

```text
caller
  |
  v
zenith.Load / zenith.Decode
  |
  v
internal/app/config.Loader
  |
  +-- ports.FileReader
  +-- ports.DecoderRegistry
  +-- ports.Mapper
```

The application service depends on interfaces, not concrete implementations. Concrete details live outside the application core as adapters.

### Layers

```text
.
├── zenith.go
├── encoding/
│   └── codec.go
└── internal/
    ├── ports/
    │   └── codec.go
    ├── app/
    │   └── config/
    │       └── loader.go
    └── adapters/
        ├── codec/
        │   ├── dotenv/
        │   ├── json/
        │   ├── toml/
        │   └── yaml/
        ├── filereader/
        └── mapper/
```

### How The Pieces Connect

`zenith.go` is the composition layer. It wires the default implementation together:

- `filereader.OS{}` reads files from disk.
- `encoding.DefaultRegistry()` finds decoders by format.
- `mapper.Reflect{}` maps decoded values into structs.
- `config.Loader` coordinates those ports.

`internal/ports` defines the boundaries:

- `FileReader` reads bytes from a path.
- `DecoderRegistry` returns a decoder for a format.
- `Decoder` parses bytes into `map[string]any`.
- `Mapper` maps decoded values into a struct.

`internal/app/config` contains the use case:

1. Get the file extension.
2. Read the file.
3. Pick a decoder from the registry.
4. Decode bytes into a map.
5. Map the map into the caller's struct.

`internal/adapters` contains implementations:

- Codec adapters handle format-specific parsing.
- The file reader adapter uses `os.ReadFile`.
- The mapper adapter uses reflection.

`encoding` is a public registry facade. It exposes codec registration and lookup while keeping concrete codec implementations under `internal/adapters`.

## Design Patterns Used

- Hexagonal Architecture: keeps application logic independent from IO, decoding, and reflection details.
- Ports and Adapters: ports define what the app needs; adapters provide how it is done.
- Facade: `zenith.Load`, `zenith.Decode`, and the `encoding` package provide simple public entry points.
- Registry: codecs are registered by format name, so the loader can choose the correct decoder dynamically.
- Dependency Inversion: the loader depends on interfaces from `internal/ports`, not concrete implementations.

## Development

Run all tests with package coverage:

```sh
make test
```

Generate a coverage profile and function-level coverage report:

```sh
make coverage
```

The coverage profile is written to `coverage.out`.
