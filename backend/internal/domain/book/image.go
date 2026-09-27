package book

import (
	"fmt"
	"strings"

	"github.com/mrstsgk/book-management-system/backend/internal/domain/common"
)

const (
	// ImageMaxBytes caps a cover image at 5 MiB.
	ImageMaxBytes    = 5 << 20
	imageKeyMaxBytes = 255
)

var imageExtensions = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// Image is a value object describing an uploaded cover image (JPEG / PNG / WebP, 1 byte..5 MiB).
type Image struct {
	contentType string
	size        int64
}

func NewImage(contentType string, size int64) (Image, error) {
	if _, ok := imageExtensions[contentType]; !ok {
		return Image{}, fmt.Errorf("%w: 画像はJPEG・PNG・WebPのいずれかにしてください", common.ErrInvalid)
	}
	if size < 1 || size > ImageMaxBytes {
		return Image{}, fmt.Errorf("%w: 画像は%dMB以下にしてください", common.ErrInvalid, ImageMaxBytes>>20)
	}
	return Image{contentType: contentType, size: size}, nil
}

func (i Image) ContentType() string {
	return i.contentType
}

func (i Image) Size() int64 {
	return i.size
}

// Extension returns the file extension (with the dot) matching the content type.
func (i Image) Extension() string {
	return imageExtensions[i.contentType]
}

// ImageKey is a value object for the storage object key of a cover image.
type ImageKey struct {
	value string
}

func NewImageKey(raw string) (ImageKey, error) {
	if raw == "" || len(raw) > imageKeyMaxBytes || strings.HasPrefix(raw, "/") {
		return ImageKey{}, fmt.Errorf("%w: 画像のキーが不正です", common.ErrInvalid)
	}
	return ImageKey{value: raw}, nil
}

func (k ImageKey) String() string {
	return k.value
}
