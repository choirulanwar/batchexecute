package batchexecute

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

func Post(url url.URL, headers http.Header, body url.Values) (*http.Response, string, error) {
	client := &http.Client{}

	bodyReader := strings.NewReader(body.Encode())

	req, err := http.NewRequest("POST", url.String(), bodyReader)
	if err != nil {
		return nil, "", err
	}

	req.Header = headers

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	return resp, string(respBody), nil
}
