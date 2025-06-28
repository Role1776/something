package pdf

import (
	"bytes"
	"strings"

	"github.com/ledongthuc/pdf"
)

func ReadPDF(fileBytes []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(fileBytes), int64(len(fileBytes)))

	if err != nil {
		return "", err
	}

	totalPage := r.NumPage()
	var textBuilder strings.Builder
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {
		page := r.Page(pageIndex)
		content := page.Content()

		for _, text := range content.Text {
			_, err := textBuilder.WriteString(text.S)
			if err != nil {
				return "", err
			}
		}

	}
	return textBuilder.String(), nil
}
