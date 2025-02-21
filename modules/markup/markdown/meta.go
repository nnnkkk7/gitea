// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package markdown

import (
	"bytes"
	"errors"
	"unicode"
	"unicode/utf8"

	"code.gitea.io/gitea/modules/log"
	"gopkg.in/yaml.v3"
)

func isYAMLSeparator(line []byte) bool {
	idx := 0
	for ; idx < len(line); idx++ {
		if line[idx] >= utf8.RuneSelf {
			r, sz := utf8.DecodeRune(line[idx:])
			if !unicode.IsSpace(r) {
				return false
			}
			idx += sz
			continue
		}
		if line[idx] != ' ' {
			break
		}
	}
	dashCount := 0
	for ; idx < len(line); idx++ {
		if line[idx] != '-' {
			break
		}
		dashCount++
	}
	if dashCount < 3 {
		return false
	}
	for ; idx < len(line); idx++ {
		if line[idx] >= utf8.RuneSelf {
			r, sz := utf8.DecodeRune(line[idx:])
			if !unicode.IsSpace(r) {
				return false
			}
			idx += sz
			continue
		}
		if line[idx] != ' ' {
			return false
		}
	}
	return true
}

// ExtractMetadata consumes a markdown file, parses YAML frontmatter,
// and returns the frontmatter metadata separated from the markdown content
func ExtractMetadata(contents string, out interface{}) (string, error) {
	body, err := ExtractMetadataBytes([]byte(contents), out)
	return string(body), err
}

// ExtractMetadata consumes a markdown file, parses YAML frontmatter,
// and returns the frontmatter metadata separated from the markdown content
func ExtractMetadataBytes(contents []byte, out interface{}) ([]byte, error) {
	var front, body []byte

	start, end := 0, len(contents)
	idx := bytes.IndexByte(contents[start:], '\n')
	if idx >= 0 {
		end = start + idx
	}
	line := contents[start:end]

	if !isYAMLSeparator(line) {
		return contents, errors.New("frontmatter must start with a separator line")
	}
	frontMatterStart := end + 1
	for start = frontMatterStart; start < len(contents); start = end + 1 {
		end = len(contents)
		idx := bytes.IndexByte(contents[start:], '\n')
		if idx >= 0 {
			end = start + idx
		}
		line := contents[start:end]
		if isYAMLSeparator(line) {
			front = contents[frontMatterStart:start]
			body = contents[end+1:]
			break
		}
	}

	if len(front) == 0 {
		return contents, errors.New("could not determine metadata")
	}

	log.Info("%s", string(front))

	if err := yaml.Unmarshal(front, out); err != nil {
		return contents, err
	}
	return body, nil
}
