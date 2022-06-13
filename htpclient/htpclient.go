package htpclient

import "net/http"

type RTOption func(r http.RoundTripper) http.RoundTripper

type htpClient struct {
	client http.Client
}

func New(rtos ...RTOption) *htpClient {
	c := &htpClient{
		client: http.Client{
			Transport: http.DefaultTransport,
		},
	}

	for _, opt := range rtos {
		c.client.Transport = opt(c.client.Transport)
	}

	return c
}

func (h *htpClient) Get(url string) (*http.Response, error) {
	return h.client.Get(url)
}
