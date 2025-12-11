package images

import "mime/multipart"

type ImageWriter interface {
	Write(file *multipart.FileHeader) (imageurl string, err error)
}

type ImageWriterFunc func(file *multipart.FileHeader) (string, error)

func (fn ImageWriterFunc) Write(file *multipart.FileHeader) (string, error) {
	return fn(file)
}
