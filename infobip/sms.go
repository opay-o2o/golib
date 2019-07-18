package infobip

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type APIClient struct {
	senderName string
	baseURL    string
	secret     string
}

func NewAPI(baseURL string, senderName string, secret string) *APIClient {
	return &APIClient{
		senderName: senderName,
		baseURL:    baseURL,
		secret:     secret,
	}
}

func (api *APIClient) GetSenderName() string {
	return api.senderName
}

func (api *APIClient) SendMessage(to, body string) (result string, err error) {
	jm := map[string]string{
		"from": api.senderName,
		"to":   to,
		"text": body,
	}

	j, err := json.Marshal(jm)

	if err != nil {
		return
	}

	fmt.Printf("j: %s\n", j)

	req, err := api.createRequest(j)

	if err != nil {
		return
	}

	client := http.Client{
		Timeout: 5 * time.Minute,
	}

	resp, err := client.Do(req)

	if err != nil {
		return
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return
	}

	if resp.StatusCode != 200 {
		err = fmt.Errorf("error when sending SMS with status=%s body=%s", resp.Status, string(bodyBytes))
		return
	}

	return string(bodyBytes), nil
}

func (api *APIClient) createRequest(body []byte) (*http.Request, error) {
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/sms/2/text/single", api.baseURL),
		bytes.NewBuffer(body),
	)

	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", fmt.Sprintf("App %s", api.secret))

	return req, err
}
