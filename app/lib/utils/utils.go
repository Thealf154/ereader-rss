package utils

import (
	"context"
	"crypto/md5"
	b64 "encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2/log"
	"github.com/h2non/bimg"
)

// Makes a blocking request to an url and returs the body as
// a pointer to a bytes array
func ReadBytesFromUrl(URL string) (*[]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	type result struct {
		data *[]byte
		err  error
	}

	ch := make(chan result, 1)

	go func() {
		log.Debug("Starting goroutine for URL:", URL)
		req, err := http.NewRequestWithContext(ctx, "GET", URL, nil)
		if err != nil {
			log.Error("Error creating request:", err)
			ch <- result{nil, err}
			return
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Error("Error executing request:", err)
			ch <- result{nil, err}
			return
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error("Error reading response body:", err)
			ch <- result{nil, err}
			return
		}

		log.Debug("Successfully read ", len(body), " bytes from URL:", URL)
		ch <- result{&body, err}
	}()

	select {
	case <-ctx.Done():
		log.Error("Context deadline exceeded for URL:", URL)
		return nil, ctx.Err()
	case r := <-ch:
		return r.data, r.err
	}
}

// Makes a request with the src and returns a compressed image in
// base64 format ready to append an src argument in an img tag
func CompressImage(src string) (string, error) {
	image, err := ReadBytesFromUrl(src)
	if err != nil {
		return "", err
	}

	options := bimg.Options{
		Quality: 30,
	}

	newImage, err := bimg.NewImage(*image).Process(options)
	if err != nil {
		return "", err
	}

	base64Image := b64.StdEncoding.EncodeToString(newImage)
	base64Image = "data:image/gif;base64," + base64Image

	return base64Image, nil
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
