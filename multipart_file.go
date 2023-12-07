package surf

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"
)

type multipartFile struct {
	data   *bytes.Buffer
	writer *multipart.Writer
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
	}
}

// AddFile adds a file to the writer
func (m *multipartFile) AddFile(field, filename string, data []byte) error {
	w, err := m.writer.CreateFormFile(field, filename)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return err
}

// AddFileReader adds a file from a reader to the writer
func (m *multipartFile) AddFileReader(field, filename string, reader io.Reader) error {
	w, err := m.writer.CreateFormFile(field, filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, reader)
	return err
}

// AddFields adds fields to the writer
func (m *multipartFile) AddFields(fields map[string]string) error {
	for k, v := range fields {
		err := m.writer.WriteField(k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

// FormDataContentType returns the content type
func (m *multipartFile) FormDataContentType() string {
	return m.writer.FormDataContentType()
}

// Bytes returns the bytes
func (m *multipartFile) Bytes() ([]byte, error) {
	err := m.writer.Close()
	if err != nil {
		return nil, err
	}
	return m.data.Bytes(), nil
}

// Reset resets the MultipartFile for reuse
func (m *multipartFile) Reset() {
	m.data.Reset()
	m.writer = multipart.NewWriter(m.data)
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
