package main

import (
	"bytes"
	"context"
	"errors"
	"github.com/chai2010/webp"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"math"
	"net/http"
	"time"
)

//<editor-fold desc="Settings">

// Encoder/decoder tags
const (
	jpegTag = "jpeg"
	pngTag  = "png"
	webpTag = "webp"
)

// conversion timer settings
const (
	minimumConversionTime  = 5.0 // seconds
	minimumKBTransferSpeed = 20.0
)

//</editor-fold>

// <editor-fold desc="image handler">

func imageHandlerFactory(decoderTag string, encoderTag string) func(writer http.ResponseWriter, request *http.Request) {

	var decoder = getDecoder(decoderTag)

	var encoder = getEncoder(encoderTag)

	if decoder == nil {
		panic("Invalid decoder")
	}
	if encoder == nil {
		panic("Invalid encoder")
	}

	if encoderTag == decoderTag {
		panic("Factory set up incorrectly. You shouldnt create a factory that returns the same object. File quality conversion is not supported for now.")
	}

	return func(writer http.ResponseWriter, request *http.Request) {

		setHandlerHeaders(writer, request, "POST", "OPTIONS")

		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusOK)
			return
		}

		if request.Method != http.MethodPost {
			http.Error(writer, "You cannot use this method with this endpoint. Try POST instead.", http.StatusBadRequest)
			return
		}

		// Getting user auth id from context. Context set in auth middleware (extracted from jwt)

		authID, isAuthOK := getUserAuthID(request)

		if !isAuthOK {
			http.Error(writer, "Failed to extract authentication token from request.", http.StatusBadRequest)
		}

		//verifying if user is allowed to do the conversion and getting max file size he is allowed to convert :)

		isConversionAllowed, maxFileSizeMB, permissionErr := canUserConvert(db, authID)

		if permissionErr != nil {
			http.Error(writer, permissionErr.Error(), http.StatusForbidden)
			return
		}

		if !isConversionAllowed {
			http.Error(writer, "User exhausted possible conversions", http.StatusForbidden)
			return
		}

		//setting up a timer so nobody tries to transfer 1kb/s or something akin to that
		timeout := calculateTimeout(maxFileSizeMB)
		contx, cancel := context.WithTimeout(request.Context(), timeout)
		defer cancel()

		//reading the file and checking if its the right size

		request = request.WithContext(contx)
		limitedReader := http.MaxBytesReader(writer, request.Body, int64(maxFileSizeMB*megabyte))
		defer limitedReader.Close()

		// Use readWithTimeout instead of io.ReadAll directly
		bodySize, readErr := readWithTimeout(contx, limitedReader)

		if readErr != nil {
			var maxBytesErr *http.MaxBytesError
			// Timeout?
			if errors.Is(readErr, context.DeadlineExceeded) {
				http.Error(writer, "Request timed out.", http.StatusRequestTimeout)
				return
			}
			// Max size exceeded?
			if errors.As(readErr, &maxBytesErr) {
				http.Error(writer, "The file provided exceeds your account limit.", http.StatusRequestEntityTooLarge)
				return
			}

			http.Error(writer, "Failed to read file.", http.StatusBadRequest)
			return
		}

		img, decodingErr := decoder(bytes.NewReader(bodySize))
		if decodingErr != nil {
			errorMessage := "Failed to decode " + decoderTag + ". Make sure its a valid image."
			http.Error(writer, errorMessage, http.StatusBadRequest)
			return
		}

		encodingErr := encoder(writer, img)
		if encodingErr != nil {
			errorMessage := "Failed to encode " + decoderTag + "."
			http.Error(writer, errorMessage, http.StatusInternalServerError)
			return
		}

		err := insertConversion(db, authID, decoderTag, encoderTag, len(bodySize))

		if err != nil {
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}

	}

}

//</editor-fold>

// <editor-fold desc="Encoders and decoders">

func jpegEncoder(writer http.ResponseWriter, image image.Image) error {
	writer.Header().Set("Content-Type", "image/jpg")

	options := &jpeg.Options{
		Quality: 90,
	}

	encodingErr := jpeg.Encode(writer, image, options)
	if encodingErr != nil {
		http.Error(writer, "failed to encode JPG.", http.StatusInternalServerError)
		return encodingErr
	}
	return nil
}

func pngEncoder(writer http.ResponseWriter, image image.Image) error {
	writer.Header().Set("Content-Type", "image/png")

	encodingErr := png.Encode(writer, image)
	if encodingErr != nil {
		http.Error(writer, "failed to encode PNG.", http.StatusInternalServerError)
		return encodingErr
	}
	return nil
}

func webpEncoder(writer http.ResponseWriter, image image.Image) error {
	writer.Header().Set("Content-Type", "image/webp")

	options := &webp.Options{
		Lossless: true,
	}

	encodingErr := webp.Encode(writer, image, options)
	if encodingErr != nil {
		http.Error(writer, "failed to encode file.", http.StatusInternalServerError)
		return encodingErr
	}
	return nil
}

func getDecoder(decoderTag string) func(r io.Reader) (image.Image, error) {
	switch decoderTag {
	case jpegTag:
		return jpeg.Decode
	case pngTag:
		return png.Decode
	case webpTag:
		return webp.Decode
	default:
		return nil
	}
}

func getEncoder(encoderTag string) func(writer http.ResponseWriter, image image.Image) error {
	switch encoderTag {
	case jpegTag:
		return jpegEncoder
	case pngTag:
		return pngEncoder
	case webpTag:
		return webpEncoder
	default:
		return nil
	}
}

//</editor-fold>

//<editor-fold desc="Timeout">

func readWithTimeout(ctx context.Context, r io.Reader) ([]byte, error) {
	type result struct {
		data []byte
		err  error
	}
	results := make(chan result, 1)

	go func() {
		data, err := io.ReadAll(r)
		results <- result{data, err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-results:
		return res.data, res.err
	}
}

func calculateTimeout(maxMB int) time.Duration {

	maxKB := 1024.0 * float64(maxMB)               // Converting MB to KB
	transferTime := maxKB / minimumKBTransferSpeed // Time in seconds
	secs := math.Max(minimumConversionTime, transferTime)
	return time.Duration(secs * float64(time.Second))

}

// </editor-fold>
