package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

const emptyLine = "\r\n"

func (h Headers) Get(key string) string {
	return h[strings.ToLower(key)]
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	lineEnd := bytes.Index(data, []byte(emptyLine))
	if lineEnd == -1 {
		return 0, false, nil
	}
	if lineEnd == 0 {
		return len(emptyLine), true, nil
	}

	line := string(data[:lineEnd])
	colonIndex := strings.IndexByte(line, ':')
	if colonIndex == -1 {
		return 0, false, fmt.Errorf("invalid header line: %s", line)
	}

	keyPart := line[:colonIndex]
	if strings.HasSuffix(keyPart, " ") || strings.HasSuffix(keyPart, "\t") {
		return 0, false, fmt.Errorf("invalid header key: %s", keyPart)
	}

	valuePart := line[colonIndex+1:]
	key := strings.TrimSpace(keyPart)
	value := strings.TrimSpace(valuePart)
	if key == "" {
		return 0, false, fmt.Errorf("invalid header key: %s", keyPart)
	}
	if !isValidFieldName(key) {
		return 0, false, fmt.Errorf("invalid header key: %s", keyPart)
	}

	lowerKey := strings.ToLower(key)
	if existing, ok := h[lowerKey]; ok {
		h[lowerKey] = existing + ", " + value
	} else {
		h[lowerKey] = value
	}
	return lineEnd + len(emptyLine), false, nil
}

func isValidFieldName(key string) bool {
	for i := 0; i < len(key); i++ {
		ch := key[i]
		if ch >= 'a' && ch <= 'z' {
			continue
		}
		if ch >= 'A' && ch <= 'Z' {
			continue
		}
		if ch >= '0' && ch <= '9' {
			continue
		}
		switch ch {
		case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
			continue
		default:
			return false
		}
	}
	return true
}
