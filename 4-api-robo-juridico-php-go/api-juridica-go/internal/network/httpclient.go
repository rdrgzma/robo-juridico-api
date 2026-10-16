package network

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var userAgents = []string{
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:131.0) Gecko/20100101 Firefox/131.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36",
}

type RespostaHTTP struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

func (r *RespostaHTTP) RetryAfter() int {
	val := r.Headers.Get("Retry-After")
	if val == "" {
		val = r.Headers.Get("retry-after")
	}
	if val != "" {
		if s, err := strconv.Atoi(val); err == nil && s > 0 {
			return s
		}
	}
	return 0
}

func (r *RespostaHTTP) OK() bool {
	return r.StatusCode == http.StatusOK
}

type ClienteHTTP struct {
	clientePadrao  *http.Client
	UserAgentExtra string
}

func NovoClienteHTTP(timeout time.Duration, userAgentExtra string) *ClienteHTTP {
	dialer := &net.Dialer{
		Timeout:   5 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          200,
		MaxIdleConnsPerHost:   50,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &ClienteHTTP{
		clientePadrao: &http.Client{
			Transport: transport,
			Timeout:   timeout,
		},
		UserAgentExtra: userAgentExtra,
	}
}

func (c *ClienteHTTP) CabecalhosNavegador(origem, referer string) http.Header {
	ua := userAgents[rand.Intn(len(userAgents))]
	if c.UserAgentExtra != "" {
		ua = fmt.Sprintf("%s %s", ua, c.UserAgentExtra)
	}

	h := make(http.Header)
	h.Set("User-Agent", ua)
	h.Set("Accept", "application/json, text/plain, */*")
	h.Set("Accept-Language", "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7")
	h.Set("Sec-Ch-Ua", `"Chromium";v="131", "Google Chrome";v="131", "Not.A/Brand";v="24"`)
	h.Set("Sec-Ch-Ua-Mobile", "?0")
	h.Set("Sec-Ch-Ua-Platform", `"Windows"`)
	h.Set("Sec-Fetch-Dest", "empty")
	h.Set("Sec-Fetch-Mode", "cors")
	h.Set("Sec-Fetch-Site", "same-site")
	h.Set("Cache-Control", "no-cache")
	h.Set("Pragma", "no-cache")

	if origem != "" {
		h.Set("Origin", origem)
	}
	if referer != "" {
		h.Set("Referer", referer)
	}
	return h
}

func (c *ClienteHTTP) Get(ctx context.Context, endpoint string, params url.Values, cabecalhos http.Header, proxyURL string) (*RespostaHTTP, error) {
	reqURL := endpoint
	if len(params) > 0 {
		if strings.Contains(reqURL, "?") {
			reqURL = fmt.Sprintf("%s&%s", reqURL, params.Encode())
		} else {
			reqURL = fmt.Sprintf("%s?%s", reqURL, params.Encode())
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	for k, vv := range cabecalhos {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}

	client := c.clientePadrao
	if proxyURL != "" {
		pURL, err := url.Parse(proxyURL)
		if err == nil {
			client = &http.Client{
				Timeout: c.clientePadrao.Timeout,
				Transport: &http.Transport{
					Proxy: http.ProxyURL(pURL),
				},
			}
		}
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &RespostaHTTP{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}

func (c *ClienteHTTP) PostJSON(ctx context.Context, endpoint string, payload []byte, cabecalhos http.Header) (*RespostaHTTP, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}

	for k, vv := range cabecalhos {
		for _, v := range vv {
			req.Header.Add(k, v)
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.clientePadrao.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &RespostaHTTP{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
	}, nil
}
