package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

const emptyLine = "\r\n"

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

	h[key] = value
	return lineEnd + len(emptyLine), false, nil
}
