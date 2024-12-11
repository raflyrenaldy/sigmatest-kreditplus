package util

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// IsImage checks if a file is an image
func IsImage(file interface{}) bool {
	switch v := file.(type) {
	case *multipart.FileHeader:
		contentType := v.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			return false
		}
		return strings.HasPrefix(mediaType, "image/")
	case *os.File:
		ext := strings.ToLower(filepath.Ext(v.Name()))
		return isImageExtension(ext)
	case string:
		ext := strings.ToLower(filepath.Ext(v))
		return isImageExtension(ext)
	case []byte:
		contentType := http.DetectContentType(v)
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			return false
		}
		return strings.HasPrefix(mediaType, "image/")
	}
	return false
}

// isImageExtension checks if a file extension corresponds to an image format
func isImageExtension(ext string) bool {
	imageExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp"}
	for _, imageExt := range imageExtensions {
		if ext == imageExt {
			return true
		}
	}
	return false
}

// IsVideo checks if a file is a video
func IsVideo(file interface{}) bool {
	switch v := file.(type) {
	case *multipart.FileHeader:
		contentType := v.Header.Get("Content-Type")
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			return false
		}
		return strings.HasPrefix(mediaType, "video/")
	case *os.File:
		ext := strings.ToLower(filepath.Ext(v.Name()))
		return isVideoExtension(ext)
	case string:
		ext := strings.ToLower(filepath.Ext(v))
		return isVideoExtension(ext)
	case []byte:
		contentType := http.DetectContentType(v)
		mediaType, _, err := mime.ParseMediaType(contentType)
		if err != nil {
			return false
		}
		return strings.HasPrefix(mediaType, "video/")
	}
	return false
}

// isVideoExtension checks if a file extension corresponds to a video format
func isVideoExtension(ext string) bool {
	videoExtensions := []string{".mp4", ".avi", ".mov", ".mkv", ".wmv"}
	for _, videoExt := range videoExtensions {
		if ext == videoExt {
			return true
		}
	}
	return false
}

// GetFileExtension returns the file extension of a file
func GetFileExtension(file interface{}) string {
	switch v := file.(type) {
	case *multipart.FileHeader:
		return strings.ToLower(filepath.Ext(v.Filename))
	case *os.File:
		return strings.ToLower(filepath.Ext(v.Name()))
	case string:
		return strings.ToLower(filepath.Ext(v))
	case []byte:
		return strings.ToLower(filepath.Ext(string(v)))
	}
	return ""
}

// CompressImage compresses an image to reduce its size
func CompressImage(src multipart.File, fileSize int64, quality int) ([]byte, error) {
	// Check if file size under 500KB don't compress the image
	if fileSize < 500000 {
		quality = 100
	}

	// Decode the image
	img, _, err := image.Decode(src)
	if err != nil {
		return nil, err
	}

	// Create a buffer to hold the compressed image data
	var buffer bytes.Buffer

	// Encode the image with the specified quality
	jpegOptions := jpeg.Options{
		Quality: quality, // Set the compression quality (1-100)
	}
	if err := jpeg.Encode(&buffer, img, &jpegOptions); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// CompressVideo compresses a video to reduce its size
func CompressVideo(src io.Reader, quality int) ([]byte, error) {
	// Read the input video from src into a buffer
	inputBuffer := new(bytes.Buffer)
	if _, err := io.Copy(inputBuffer, src); err != nil {
		return nil, err
	}

	// Create an output buffer to store the compressed video
	outputBuffer := new(bytes.Buffer)

	// Create the ffmpeg command to compress the video
	cmd := exec.Command(
		"ffmpeg",
		"-i", "pipe:0", // Input from stdin
		"-c:v", "libx264", // Video codec (you can choose another codec if needed)
		"-crf", fmt.Sprintf("%d", quality), // Compression quality (lower values mean higher quality)
		"-f", "mp4", // Output format (you can choose another format)
		"-preset", "medium", // Compression preset (adjust as needed)
		"pipe:1", // Output to stdout
	)

	// Set the input and output pipes for the command
	cmd.Stdin = inputBuffer
	cmd.Stdout = outputBuffer

	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, err
	}

	// Return the compressed video as bytes
	return outputBuffer.Bytes(), nil
}

// ConvertImageToJPG converts an image to JPG
func ConvertImageToJPG(file interface{}, quality int) ([]byte, error) {
	// Check if the input file is an image
	if !IsImage(file) {
		return nil, errors.New("input is not an image")
	}

	switch v := file.(type) {
	case *multipart.FileHeader:
		// Open the file
		srcFile, err := v.Open()
		if err != nil {
			return nil, err
		}
		defer srcFile.Close()

		return convertImageToJPG(srcFile, quality)

	case *os.File:
		return convertImageToJPG(v, quality)

	case string:
		file, err := os.Open(v)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		return convertImageToJPG(file, quality)

	case []byte:
		// Create a reader for the byte slice
		reader := bytes.NewReader(v)

		return convertImageToJPG(reader, quality)
	}
	return nil, errors.New("unsupported file type")
}

func convertImageToJPG(src io.Reader, quality int) ([]byte, error) {
	// Try to determine the image format from the source
	imgFormat := ""
	if rc, ok := src.(io.ReadSeeker); ok {
		// Peek at the beginning of the source to determine the image format
		buf := make([]byte, 512) // Read enough bytes to detect the format
		_, err := rc.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}
		_, err = io.ReadFull(rc, buf)
		if err != nil {
			return nil, err
		}
		imgFormat = http.DetectContentType(buf)
	}

	// Reset the source to the beginning
	if seeker, ok := src.(io.Seeker); ok {
		_, err := seeker.Seek(0, io.SeekStart)
		if err != nil {
			return nil, err
		}
	}

	// Decode the image based on its format
	var img image.Image
	var err error
	switch {
	case strings.HasPrefix(imgFormat, "image/jpeg"):
		img, err = jpeg.Decode(src)
	case strings.HasPrefix(imgFormat, "image/png"):
		img, err = png.Decode(src)
	case strings.HasPrefix(imgFormat, "image/gif"):
		img, err = gif.Decode(src)
	default:
		return nil, errors.New("unsupported image format")
	}

	if err != nil {
		return nil, err
	}

	// Create a buffer to hold the JPEG data
	var buffer bytes.Buffer

	// Encode the image as JPEG into the buffer
	jpegOptions := jpeg.Options{
		Quality: quality, // Set the compression quality (1-100)
	}
	if err := jpeg.Encode(&buffer, img, &jpegOptions); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// ConvertVideoToMP4 converts a video to MP4
func ConvertVideoToMP4(file interface{}, quality int) ([]byte, error) {
	// Check if the input file is a video
	if !IsVideo(file) {
		return nil, errors.New("input is not a video")
	}

	switch v := file.(type) {
	case *multipart.FileHeader:
		// Open the file
		srcFile, err := v.Open()
		if err != nil {
			return nil, err
		}
		defer srcFile.Close()

		return convertVideoToMP4(srcFile, quality)

	case *os.File:
		return convertVideoToMP4(v, quality)

	case string:
		file, err := os.Open(v)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		return convertVideoToMP4(file, quality)

	case []byte:
		// Create a reader for the byte slice
		reader := bytes.NewReader(v)

		return convertVideoToMP4(reader, quality)
	}
	return nil, errors.New("unsupported file type")
}

func convertVideoToMP4(src io.Reader, quality int) ([]byte, error) {
	// Read the input video from src into a buffer
	inputBuffer := new(bytes.Buffer)
	if _, err := io.Copy(inputBuffer, src); err != nil {
		return nil, err
	}

	// Create an output buffer to store the compressed video
	outputBuffer := new(bytes.Buffer)

	// Create the ffmpeg command to compress the video
	cmd := exec.Command(
		"ffmpeg",
		"-analyzeduration", "2147483647", // Set to a large value
		"-probesize", "10000000", // Set to a large value
		"-i", "pipe:0", // Input from stdin
		"-c:v", "libx264", // Video codec (you can choose another codec if needed)
		"-crf", fmt.Sprintf("%d", quality), // Compression quality (lower values mean higher quality)
		"-f", "mp4", // Output format (you can choose another format)
		"-preset", "medium", // Compression preset (adjust as needed)
		"pipe:1", // Output to stdout
	)

	// Set the input and output pipes for the command
	cmd.Stdin = inputBuffer
	cmd.Stdout = outputBuffer

	// Capture stderr for error checking
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Run the command
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg error: %s", stderr.String()) // Include stderr in the error message
	}

	// Return the compressed video as bytes
	return outputBuffer.Bytes(), nil
}
