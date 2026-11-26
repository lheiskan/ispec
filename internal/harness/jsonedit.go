package harness

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// A harness config file belongs to the user, not to ispec. It holds keys this
// binary has never heard of, in an order the user can read, and ispec owns
// exactly one entry inside it. Decoding into map[string]any would reorder
// every key, rewrite every number, and drop nothing but change everything —
// turning a one-line install into an unreviewable diff.
//
// So: decode to ordered key/raw-value pairs, replace only the one member ispec
// owns, and re-emit every other value from its original bytes.

// pair is one member of a JSON object, value unparsed.
type pair struct {
	key string
	raw json.RawMessage
}

// object is a JSON object with member order preserved.
type object []pair

// decodeObject parses a JSON object into ordered pairs. An empty input is an
// empty object, so a missing config file and an empty one behave alike.
func decodeObject(b []byte) (object, error) {
	if len(bytes.TrimSpace(b)) == 0 {
		return object{}, nil
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return nil, fmt.Errorf("expected a JSON object")
	}
	var obj object
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, ok := keyTok.(string)
		if !ok {
			return nil, fmt.Errorf("expected an object key, got %v", keyTok)
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return nil, err
		}
		obj = obj.set(key, raw)
	}
	if _, err := dec.Token(); err != nil { // closing brace
		return nil, err
	}
	return obj, nil
}

// get returns a member's raw value.
func (o object) get(key string) (json.RawMessage, bool) {
	for _, p := range o {
		if p.key == key {
			return p.raw, true
		}
	}
	return nil, false
}

// set replaces a member in place, or appends it if absent. Replacing in place
// matters: moving "hooks" to the end of settings.json on every install would
// be a visible, pointless change to the user's file.
func (o object) set(key string, raw json.RawMessage) object {
	for i, p := range o {
		if p.key == key {
			o[i].raw = raw
			return o
		}
	}
	return append(o, pair{key: key, raw: raw})
}

// encode renders the object indented two spaces per level, starting at depth.
// Member values are re-indented from their original bytes rather than
// re-marshalled, so unknown fields survive untouched.
func (o object) encode(depth int) ([]byte, error) {
	if len(o) == 0 {
		return []byte("{}"), nil
	}
	inner := indentAt(depth + 1)
	var buf bytes.Buffer
	buf.WriteString("{\n")
	for i, p := range o {
		buf.WriteString(inner)
		key, err := marshalString(p.key)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(": ")
		val, err := reindent(p.raw, inner)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
		if i < len(o)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString(indentAt(depth))
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// encodeArray renders an array of raw values indented two spaces per level,
// splicing each element's original bytes.
func encodeArray(items []json.RawMessage, depth int) ([]byte, error) {
	if len(items) == 0 {
		return []byte("[]"), nil
	}
	inner := indentAt(depth + 1)
	var buf bytes.Buffer
	buf.WriteString("[\n")
	for i, it := range items {
		buf.WriteString(inner)
		val, err := reindent(it, inner)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
		if i < len(items)-1 {
			buf.WriteByte(',')
		}
		buf.WriteByte('\n')
	}
	buf.WriteString(indentAt(depth))
	buf.WriteByte(']')
	return buf.Bytes(), nil
}

// reindent re-wraps a raw value so its continuation lines sit at prefix.
// json.Indent does not indent the first line, which is what we want: the
// caller has already written the key and separator.
func reindent(raw json.RawMessage, prefix string) ([]byte, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, prefix, "  "); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// marshalString quotes a string without HTML-escaping it. A command path
// containing & or < must round-trip literally.
func marshalString(s string) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return nil, err
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func indentAt(depth int) string {
	out := make([]byte, 0, depth*2)
	for i := 0; i < depth; i++ {
		out = append(out, ' ', ' ')
	}
	return string(out)
}
