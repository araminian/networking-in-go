package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMultipartPost(t *testing.T) {
	reqBody := new(bytes.Buffer)

	// create a new multipart writer
	w := multipart.NewWriter(reqBody)

	/*
		The multipart writer
		generates a random boundary upon initialization. Finally, you write form
		fields to the multipart writer. The multipart writer separates each form
		field into its own part, writing the boundary, appropriate headers, and the
		form field value to each part’s body
	*/
	for k, v := range map[string]string{
		"date":        time.Now().Format(time.RFC3339),
		"description": "Form values with attached files",
	} {
		// write the key-value pair to the multipart writer
		err := w.WriteField(k, v)
		if err != nil {
			t.Fatal(err)
		}
	}

	/*
		At this point, your request body has two parts, one for the date form
		field and one for the description form field.
	*/

	for i, file := range []string{
		"./files/hello.txt",
		"./files/goodbye.txt",
	} {

		// create a new form file part
		filePart, err := w.CreateFormFile(fmt.Sprintf("file%d", i), filepath.Base(file))
		if err != nil {
			t.Fatal(err)
		}

		f, err := os.Open(file)
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()

		// copy the file to the form file part
		_, err = io.Copy(filePart, f)
		_ = f.Close()
		if err != nil {
			t.Fatal(err)
		}
	}

	err := w.Close()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(),
		60*time.Second)
	defer cancel()

	// create a new request with the context
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://httpbin.org/post",
		reqBody,
	)
	if err != nil {
		t.Fatal(err)
	}
	// set the content type to the multipart content type
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()

	// read the response body
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d; actual status %d",
			http.StatusOK, resp.StatusCode)
	}
	t.Logf("\n%s", b)
}
