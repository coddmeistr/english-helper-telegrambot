package gTranslate

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type roundTripFunc func(r *http.Request) (*http.Response, error)

func (s roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return s(r)
}

func newClientMock(t *testing.T, statusCode int, path string, body Response) *Client {
	return &Client{
		client: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				assert.Equal(t, path, r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				bytesBody, err := json.Marshal(body)
				if err != nil {
					assert.Fail(t, "Cannot read bytes")
				}
				return &http.Response{
					StatusCode: statusCode,
					Body:       io.NopCloser(bytes.NewReader(bytesBody)),
				}, nil
			}),
		},
	}
}

func TestClient_TranslateText(t *testing.T) {

	type args struct {
		Text   string
		Target string
		Source string
	}

	tests := []struct {
		name                 string
		args                 args
		expectedStatusCode   int
		expectedResponse     Response
		expectedErrorMessage string
		want                 string
		wantErr              bool
	}{
		{
			name: "Ok",
			args: args{
				"car",
				"en",
				"ru",
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: Response{
				Data: Data{
					Translations: []Translations{
						{Text: "машина"},
					},
				},
			},
			expectedErrorMessage: "",
			want:                 "машина",
			wantErr:              false,
		},
		{
			name: "Error",
			args: args{
				"car",
				"en",
				"ru",
			},
			expectedStatusCode: http.StatusBadGateway,
			expectedResponse: Response{
				Data: Data{
					Translations: []Translations{},
				},
			},
			expectedErrorMessage: errForeighApi.Error(),
			want:                 "",
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newClientMock(t, tt.expectedStatusCode, "/language/translate/v2", tt.expectedResponse)

			got, err := c.TranslateText(tt.args.Text, tt.args.Target, tt.args.Source)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorMessage, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}

}

func TestClient_NewClient(t *testing.T) {
	type args struct {
		cfg        Config
		httpclient *http.Client
	}

	tests := []struct {
		name                 string
		args                 args
		want                 IClient
		expectedErrorMessage string
		wantErr              bool
	}{
		{
			name: "Ok",
			args: args{
				cfg: Config{
					Key: "valid key",
				},
				httpclient: &http.Client{},
			},
			want: &Client{
				config: Config{
					Key: "valid key",
				},
				client: &http.Client{},
			},
			expectedErrorMessage: "",
			wantErr:              false,
		},
		{
			name: "Empty API key",
			args: args{
				cfg: Config{
					Key: "",
				},
				httpclient: &http.Client{},
			},
			want:                 nil,
			expectedErrorMessage: errEmptyApiKey.Error(),
			wantErr:              true,
		},
		{
			name: "Nil Http Client",
			args: args{
				cfg: Config{
					Key: "valid key",
				},
				httpclient: nil,
			},
			want:                 nil,
			expectedErrorMessage: errNilHttpClient.Error(),
			wantErr:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewClient(tt.args.cfg, tt.args.httpclient)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErrorMessage, err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
