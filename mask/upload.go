package mask

import (
	"code.google.com/p/graphics-go/graphics"
	"errors"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"
)

type UploadOptions struct {
	StorePath, ThumbnailStorePath                        string
	MaxHeight, MaxWidth, ThumbnailHeight, ThumbnailWidth int
}

const maxFileSize = 10*1000*1024 + 1

// DefaultUploadOptions returns a the default configuration options for uploading
func DefaultUploadOptions(config *Config) UploadOptions {
	return UploadOptions{
		MaxHeight:          3000,
		MaxWidth:           6000,
		ThumbnailHeight:    150,
		ThumbnailWidth:     150,
		StorePath:          config.StorePath,
		ThumbnailStorePath: config.ThumbnailStorePath,
	}
}

// RetrieveUploadedImage returns the uploaded file at the given key
func RetrieveUploadedImage(r *http.Request, key string) (io.ReadCloser, error) {
	f, _, err := r.FormFile(key)
	if err == nil && f != nil {
		cLen := r.Header.Get("Content-Length")
		if len, err := strconv.ParseInt(cLen, 10, 64); err != nil || len > maxFileSize {
			return nil, errors.New("file too large")
		}

		return f, nil
	}

	return nil, errors.New("no file was uploaded")
}

// StoreImage stores in disk a file received with the request
func StoreImage(file io.ReadCloser, options UploadOptions) (string, string, error) {
	defer file.Close()

	img, format, err := image.Decode(file)
	if err != nil {
		return "", "", err
	}

	if format != "gif" && format != "png" && format != "jpg" && format != "jpeg" {
		return "", "", errors.New("invalid file format")
	}

	if img.Bounds().Max.X > options.MaxWidth || img.Bounds().Max.Y > options.MaxHeight {
		return "", "", errors.New("file dimensions are too large")
	}

	imagePath := options.StorePath + NewFileName(format)
	dst, err := os.Create(imagePath)
	defer dst.Close()
	if err != nil {
		return "", "", err
	}

	if err := writeFile(dst, img, format); err != nil {
		return "", "", err
	}

	thumbnail, err := generateThumbnail(img, options)
	if err != nil {
		return "", "", err
	}

	thumbnailPath := options.ThumbnailStorePath + NewFileName(format)
	thumbDst, err := os.Create(thumbnailPath)
	if err != nil {
		return "", "", err
	}

	if err := writeFile(thumbDst, thumbnail, format); err != nil {
		return "", "", nil
	}

	thumbDst.Close()

	return imagePath, thumbnailPath, nil
}

// CodeAndMessageForUploadError returns a code and an error message for the given error
func CodeAndMessageForUploadError(err error) (int, string) {
	var (
		code    int
		message string
	)

	switch err.Error() {
	case "file too large":
		code = CodeFileTooLarge
		message = MsgFileTooLarge
		break
	case "no file was uploaded":
		code = CodeNoFileUploaded
		message = MsgNoFileUploaded
		break

	case "invalid file format":
		code = CodeInvalidFileFormat
		message = MsgInvalidFileFormat
		break
	case "file dimensions are too large":
		code = CodeInvalidFileDimensions
		message = MsgInvalidFileDimensions
		break
	default:
		code = CodeInvalidFile
		message = MsgInvalidFile
	}

	return code, message
}

func generateThumbnail(src image.Image, options UploadOptions) (image.Image, error) {
	dst := image.NewRGBA(image.Rect(0, 0, options.ThumbnailWidth, options.ThumbnailHeight))
	if err := graphics.Thumbnail(dst, src); err != nil {
		return nil, err
	}

	return dst, nil
}

func writeFile(w io.Writer, i image.Image, format string) error {
	var err error

	switch format {
	case "png":
		err = png.Encode(w, i)
		break

	case "jpeg":
		err = jpeg.Encode(w, i, &jpeg.Options{Quality: 100})
		break

	case "gif":
		err = gif.Encode(w, i, nil)
		break
	}

	return err
}