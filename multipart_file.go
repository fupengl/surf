package surf

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
)

type multipartFile struct {
	data   *bytes.Buffer
	writer *multipart.Writer
	errors []error // Collect errors
}

// NewMultipartFile creates a multipart file writer with optional initial capacity
func NewMultipartFile(initCap int) *multipartFile {
	if initCap <= 0 {
		initCap = 100 * 1024 // default 100KB
	}
	buf := make([]byte, 0, initCap)
	b := bytes.NewBuffer(buf)
	return &multipartFile{
		data:   b,
		writer: multipart.NewWriter(b),
		errors: nil,
	}
}

// AddFile adds a file to the writer
func (m *multipartFile) AddFile(field, filename string, data []byte) {
	w, err := m.writer.CreateFormFile(field, filename)
	if err != nil {
		m.saveError(err)
		return
	}
	_, err = w.Write(data)
	if err != nil {
		m.saveError(err)
		return
	}
}

// AddFileReader adds a file from a reader to the writer
func (m *multipartFile) AddFileReader(field, filename string, reader io.Reader) {
	if reader == nil {
		m.saveError(fmt.Errorf("multipartFile field:%s filename:%s reader is nil", field, filename))
		return
	}
	w, err := m.writer.CreateFormFile(field, filename)
	if err != nil {
		m.saveError(err)
		return
	}
	_, err = io.Copy(w, reader)
	if err != nil {
		m.saveError(err)
	}
}

// AddFileFromPath reads a file from a file path and adds it to the writer
func (m *multipartFile) AddFileFromPath(field, path string) {
	file, err := os.Open(path)
	if err != nil {
		m.saveError(err)
		return
	}
	defer file.Close()

	m.AddFileReader(field, file.Name(), file)
}

// AddField add field to the writer
func (m *multipartFile) AddField(field, filename string) {
	err := m.writer.WriteField(field, filename)
	if err != nil {
		m.saveError(err)
	}
}

// AddFields adds fields to the writer
func (m *multipartFile) AddFields(fields map[string]string) {
	for k, v := range fields {
		m.AddField(k, v)
	}
}

// FormDataContentType returns the content type
func (m *multipartFile) FormDataContentType() string {
	return m.writer.FormDataContentType()
}

// Bytes returns the bytes
func (m *multipartFile) Bytes() ([]byte, error) {
	err := m.writer.Close()
	if err != nil {
		m.saveError(err)
	}

	if len(m.errors) > 0 {
		// If there are errors, combine them into a single error and return
		var errMsg string
		for _, e := range m.errors {
			errMsg += e.Error() + "; "
		}
		return nil, errors.New(errMsg[:len(errMsg)-2]) // Removing trailing "; "
	}

	return m.data.Bytes(), nil
}

// Reset resets the MultipartFile for reuse
func (m *multipartFile) Reset() {
	m.data.Reset()
	m.writer = multipart.NewWriter(m.data)
	m.errors = nil
}

// SetWriter sets a custom multipart.Writer for advanced usage
func (m *multipartFile) SetWriter(writer *multipart.Writer) {
	m.writer = writer
}

// SetCustomBuffer sets a custom bytes.Buffer for advanced usage
func (m *multipartFile) SetCustomBuffer(buffer *bytes.Buffer) {
	m.data = buffer
}

// SetFileWriter sets a file as the writer for large files
func (m *multipartFile) SetFileWriter(file *os.File) {
	m.writer = multipart.NewWriter(file)
}

func (m *multipartFile) saveError(err error) {
	m.errors = append(m.errors, err)
}
