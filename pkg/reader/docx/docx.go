package docx

import (
	"bytes"
	"strings"

	"baliance.com/gooxml/document"
)

func ReadDocx(fileBytes []byte) (string, error) {
	doc, err := document.Read(bytes.NewReader(fileBytes), int64(len(fileBytes)))
	if err != nil {
		return "", err
	}
	var builder strings.Builder
	for _, para := range doc.Paragraphs() {
		for _, run := range para.Runs() {
			_, err := builder.WriteString(run.Text())
			if err != nil {
				return "", err
			}
		}
	}
	return builder.String(), nil
}
