package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

func upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": 0,
			"error":   fmt.Sprintf("get form err: %s", err.Error()),
		})
	}

	// save locat
	fileName := file.Filename
	dst := viper.GetString("uploaddir") + fileName
	if err := c.SaveUploadedFile(file, dst); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": 0,
			"error":   fmt.Sprintf("upload file err: %s", err.Error()),
		})
	}

	// up to ipfs
	hash, err := uploadToIPFS(dst)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": 0,
			"error":   err,
		})
	}

	// success
	c.JSON(http.StatusOK, gin.H{
		"success": 1,
		"data":    map[string]string{"hash": hash},
	})
}

func uploadToIPFS(filePath string) (string, error) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	// Add your image file
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	fw, err := w.CreateFormFile("image", "image")
	if err != nil {
		return "", err
	}
	if _, err = io.Copy(fw, f); err != nil {
		return "", err
	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	ipfsUrl := viper.GetString("ipfsUrl")
	req, err := http.NewRequest("POST", ipfsUrl, &b)
	if err != nil {
		return "", err
	}
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", res.Status)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	bodyStr := strings.Trim(string(body), " ")
	bodyStr = strings.Trim(string(body), "\r\n")
	body = []byte(bodyStr)

	var data map[string]interface{}
	json.Unmarshal(body, &data)

	hash := data["Hash"].(string)

	return hash, nil
}
