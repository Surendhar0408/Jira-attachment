package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {

	var ticketid string

	paybytes, err := os.ReadFile("payload.json")
	if err != nil {
		fmt.Printf("Error reading payload file: %v\n", err)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("POST", "https://crayonte.atlassian.net/rest/servicedeskapi/request/", strings.NewReader(strings.TrimSpace(string(paybytes))))

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "")

	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode == 201 {
		fmt.Println("Ticket created")
	} else {

		fmt.Printf("Failed to create JIRA ticket: %v\n", err)
		return

	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}

	for _, ticketid = range strings.Split(string(body), ",") {

		if strings.Contains(strings.TrimSpace(ticketid), `"issueKey":`) {

			break
		}
	}
	fmt.Println(string(body))
	ticketid = strings.Split(strings.TrimSpace(ticketid), ":")[1]

	GetAllFiles(ticketid)

}

func GetAllFiles(ticketid string) {

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Failed to get current Directory: %v\n", err)
		return
	}
	files, err := os.ReadDir(pwd)
	if err != nil {
		fmt.Printf("Failed to list attachments: %v\n", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			if filepath.Ext(strings.TrimSpace(strings.ToLower(file.Name()))) == ".pdf" {
				fileinfo, err := file.Info()
				if err != nil {
					fmt.Printf("Failed to get attachment info: %v\n", err)
					return
				}
				if fileinfo.ModTime().Month() == time.Now().Month() && fileinfo.ModTime().Year() == time.Now().Year() {
					Addattachment(ticketid, fileinfo.Name())
				}

			}
		}
	}

}

func Addattachment(ticketid string, filename string) {

	// Prepare the file for upload
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		fmt.Printf("Failed to create form file: %v\n", err)
		return
	}
	_, err = io.Copy(part, file)
	if err != nil {
		fmt.Printf("Failed to copy file: %v\n", err)
		return
	}
	writer.Close()
	fmt.Println(strings.TrimSpace(ticketid[1 : len(ticketid)-1]))

	url := "https://crayonte.atlassian.net/rest/api/2/issue/" + strings.TrimSpace(ticketid[1:len(ticketid)-1]) + "/attachments"

	fmt.Println(url)

	req, err := http.NewRequest("POST", url, &requestBody)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Set("X-Atlassian-Token", "no-check")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", "")

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to upload file: %v\n", err)
		return
	}

	defer resp.Body.Close()

	// Check the response
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		fmt.Printf("Error response from Jira: %s\n", resp.Status)
		return
	}

	fmt.Printf("%s uploaded successfully to ticketId: %s", filename, ticketid)

}
