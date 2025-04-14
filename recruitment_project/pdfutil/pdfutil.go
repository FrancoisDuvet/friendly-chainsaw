package pdfutil

import (
	"github.com/pdfcpu/pdfcpu/pkg/api"
    "io/ioutil"
    "os"
)

// ExtractTextFromPDF extracts text from a PDF file
func ExtractTextFromPDF(filePath string) (string, error) {
    tempFile, err := ioutil.TempFile("", "extracted-*.txt")
    if err != nil {
        return "", err
    }
    defer os.Remove(tempFile.Name())

    if err := api.ExtractContentFile(filePath, tempFile.Name(), nil, nil); err != nil {
        return "", err
    }

    content, err := ioutil.ReadFile(tempFile.Name())
    if err != nil {
        return "", err
    }

    return string(content), nil
}
