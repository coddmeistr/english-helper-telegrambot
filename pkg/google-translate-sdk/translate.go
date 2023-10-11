package gTranslate

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"path"
)

var (
	errForeighApi    = errors.New("Something wrong on the google's server side.")
	errEmptyApiKey   = errors.New("API Key for google's api cannot be empty")
	errNilHttpClient = errors.New("Http Client cannot be nil")
)

//go:generate mockgen -source=translate.go -destination=mocks/mock.go

type IClient interface {
	TranslateText(text string, target string, source string) (string, error)
}

const (
	host         = "https://translation.googleapis.com"
	translateURL = "/language/translate/v2"
)

type Translations struct {
	Text string `json:"translatedText"`
}

type Data struct {
	Translations []Translations `json:"translations"`
}

type Response struct {
	Data Data `json:"data"`
}

type Config struct {
	Key string
}

type Client struct {
	config Config
	client *http.Client
}

func NewClient(config Config, client *http.Client) (IClient, error) {
	if config.Key == "" {
		return nil, errEmptyApiKey
	}
	if client == nil {
		return nil, errNilHttpClient
	}
	return &Client{
		config: config,
		client: client,
	}, nil
}

func (c *Client) TranslateText(text string, target string, source string) (string, error) {

	url, err := url.ParseRequestURI(host)
	if err != nil {
		return "", err
	}
	url.Path = path.Join(url.Path, translateURL)

	q := url.Query()
	q.Set("model", "base")
	q.Set("target", target)
	q.Set("source", source)
	q.Set("format", "text")
	q.Set("q", text)
	q.Set("key", c.config.Key)
	url.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		return "", err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", errForeighApi
	}

	var respData Response
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		return "", err
	}

	return respData.Data.Translations[0].Text, nil
}
