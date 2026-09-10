package encoding

import (
	"reflect"
	"testing"
)

type testCodec struct{}

func (testCodec) Encode(map[string]any) ([]byte, error) {
	return []byte("ok"), nil
}

func (testCodec) Decode([]byte, *map[string]any) error {
	return nil
}

func TestCodecForRegisteredFormats(t *testing.T) {
	for _, format := range []string{"json", ".json", " JSON ", "dotenv", "env"} {
		codec, err := CodecFor(format)
		if err != nil {
			t.Fatalf("CodecFor(%q) returned error: %v", format, err)
		}
		if codec == nil {
			t.Fatalf("CodecFor(%q) returned nil codec", format)
		}
	}
}

func TestEncoderForAndDecoderFor(t *testing.T) {
	if _, err := EncoderFor("json"); err != nil {
		t.Fatalf("EncoderFor() returned error: %v", err)
	}
	if _, err := DecoderFor("json"); err != nil {
		t.Fatalf("DecoderFor() returned error: %v", err)
	}
}

func TestRegister(t *testing.T) {
	codec := testCodec{}
	if err := Register(" custom-test ", codec); err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}
	defer delete(codecs, "custom-test")

	got, err := CodecFor("custom-test")
	if err != nil {
		t.Fatalf("CodecFor() returned error: %v", err)
	}
	if got != codec {
		t.Fatalf("CodecFor() = %#v, want %#v", got, codec)
	}
}

func TestRegisterErrors(t *testing.T) {
	if err := Register("", testCodec{}); err == nil {
		t.Fatal("Register() with empty format returned nil error")
	}
	if err := Register("bad", nil); err == nil {
		t.Fatal("Register() with nil codec returned nil error")
	}
}

func TestCodecForUnknownFormat(t *testing.T) {
	if _, err := CodecFor("xml"); err == nil {
		t.Fatal("CodecFor() with unknown format returned nil error")
	}
}

func TestRegisteredFormats(t *testing.T) {
	want := []string{"dotenv", "env", "json", "toml", "yaml", "yml"}
	if got := RegisteredFormats(); !reflect.DeepEqual(got, want) {
		t.Fatalf("RegisteredFormats() = %#v, want %#v", got, want)
	}
}
