package nusii

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	defaultBaseURL    = "https://app.nusii.com"
	defaultTimeout    = 30 * time.Second
	defaultMaxRetries = 2
	// A Retry-After longer than this is not worth blocking on: the error is returned instead.
	maxRetryAfter = 60 * time.Second
)

// Nusii is a client for the Nusii API. Create one with New. It is safe for
// concurrent use by multiple goroutines.
type Nusii struct {
	cfg config

	// Account is the account the credentials belong to.
	Account   *AccountService
	Clients   *ClientsService
	Proposals *ProposalsService
	Sections  *SectionsService
	LineItems *LineItemsService
	// Taxations are the taxes and discounts applied to a proposal's total.
	Taxations          *TaxationsService
	Templates          *TemplatesService
	ProposalActivities *ProposalActivitiesService
	// Users are the team members of the account.
	Users            *UsersService
	WebhookEndpoints *WebhookEndpointsService
	Themes           *ThemesService
	Currencies       *CurrenciesService
	Locales          *LocalesService
	PDFPageSizes     *PDFPageSizesService
	// Integrations are CRM integration installations.
	Integrations *IntegrationsService
}

// New returns a client for the Nusii API.
//
// Without options it reads the NUSII_API_KEY (or NUSII_ACCESS_TOKEN) and
// NUSII_BASE_URL environment variables. It returns ErrMissingCredentials when
// neither an API key nor an access token is given or set in the environment.
func New(opts ...Option) (*Nusii, error) {
	cfg := config{
		timeout:    defaultTimeout,
		maxRetries: defaultMaxRetries,
		headers:    http.Header{},
		httpClient: http.DefaultClient,
		sleep:      sleep,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.apiKey == "" && cfg.accessToken == "" {
		if key := readEnv("NUSII_API_KEY"); key != "" {
			cfg.apiKey = key
		} else if token := readEnv("NUSII_ACCESS_TOKEN"); token != "" {
			cfg.accessToken = token
		} else {
			return nil, ErrMissingCredentials
		}
	}
	if cfg.baseURL == "" {
		cfg.baseURL = readEnv("NUSII_BASE_URL")
	}
	if cfg.baseURL == "" {
		cfg.baseURL = defaultBaseURL
	}
	cfg.baseURL = normalizeBaseURL(cfg.baseURL)

	n := &Nusii{cfg: cfg}
	n.Account = &AccountService{n}
	n.Clients = &ClientsService{n}
	n.Proposals = &ProposalsService{n}
	n.Sections = &SectionsService{n}
	n.LineItems = &LineItemsService{n}
	n.Taxations = &TaxationsService{n}
	n.Templates = &TemplatesService{n}
	n.ProposalActivities = &ProposalActivitiesService{n}
	n.Users = &UsersService{n}
	n.WebhookEndpoints = &WebhookEndpointsService{n}
	n.Themes = &ThemesService{n}
	n.Currencies = &CurrenciesService{n}
	n.Locales = &LocalesService{n}
	n.PDFPageSizes = &PDFPageSizesService{n}
	n.Integrations = &IntegrationsService{n}
	return n, nil
}

// BaseURL returns the URL of the Nusii app the client talks to, without the
// /api/v2 suffix.
func (n *Nusii) BaseURL() string {
	return n.cfg.baseURL
}

// RequestConfig describes a request for Nusii.Request.
type RequestConfig struct {
	// Method is the HTTP method, e.g. http.MethodGet. Defaults to GET.
	Method string
	// Path is the path below /api/v2, e.g. "/clients".
	Path string
	// Query holds the query parameters.
	Query url.Values
	// Body is encoded as JSON when it is not nil.
	Body any
}

// Request sends a request to any API v2 endpoint and decodes the JSON
// response body into out, unless out is nil or the body is empty. JSON:API
// documents are decoded as they are, not flattened like the typed methods do.
// Useful for endpoints this library does not wrap yet.
func (n *Nusii) Request(ctx context.Context, rc RequestConfig, out any, opts ...Option) error {
	body, err := n.do(ctx, rc, opts)
	if err != nil {
		return err
	}
	if out == nil || len(body) == 0 {
		return nil
	}
	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("nusii: decoding the response of %s %s: %w", methodOrGet(rc.Method), rc.Path, err)
	}
	return nil
}

// do sends the request, retrying it when that is safe, and returns the body
// of a successful response.
func (n *Nusii) do(ctx context.Context, rc RequestConfig, opts []Option) ([]byte, error) {
	cfg := n.cfg.with(opts)
	method := methodOrGet(rc.Method)
	target := cfg.baseURL + "/api/v2" + rc.Path
	if len(rc.Query) > 0 {
		target += "?" + rc.Query.Encode()
	}

	var payload []byte
	if rc.Body != nil {
		var err error
		if payload, err = json.Marshal(rc.Body); err != nil {
			return nil, fmt.Errorf("nusii: encoding the request body: %w", err)
		}
	}

	for attempt := 0; ; attempt++ {
		resp, body, err := send(ctx, cfg, method, target, payload)
		if err != nil {
			var connErr *ConnectionError
			if !errors.As(err, &connErr) {
				return nil, err
			}
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			if method == http.MethodGet && attempt < cfg.maxRetries {
				if err := cfg.sleep(ctx, backoff(attempt)); err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			if len(body) > 0 && !json.Valid(body) {
				contentType := resp.Header.Get("Content-Type")
				if contentType == "" {
					contentType = "an unknown content type"
				}
				return nil, &APIError{
					StatusCode: resp.StatusCode,
					Message:    fmt.Sprintf("expected JSON from the Nusii API but received %s. Check the base URL", contentType),
					Header:     resp.Header,
					Body:       body,
				}
			}
			return body, nil
		}

		if delay, ok := retryDelay(method, resp, attempt, cfg.maxRetries); ok {
			if err := cfg.sleep(ctx, delay); err != nil {
				return nil, err
			}
			continue
		}
		return nil, newAPIError(resp.StatusCode, resp.Header, body)
	}
}

// send makes one attempt. The timeout covers reading the body too.
func send(ctx context.Context, cfg config, method, target string, payload []byte) (*http.Response, []byte, error) {
	attemptCtx := ctx
	if cfg.timeout > 0 {
		var cancel context.CancelFunc
		attemptCtx, cancel = context.WithTimeout(ctx, cfg.timeout)
		defer cancel()
	}

	var reader io.Reader
	if payload != nil {
		reader = bytes.NewReader(payload)
	}
	req, err := http.NewRequestWithContext(attemptCtx, method, target, reader)
	if err != nil {
		return nil, nil, fmt.Errorf("nusii: building the request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	// Always sent, also on GET: without it Nusii answers an invalid key with
	// a redirect to the login page instead of a 401.
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "nusii-go/"+Version)
	for name, values := range cfg.headers {
		req.Header[name] = values
	}
	req.Header.Set("Authorization", cfg.authorization())

	resp, err := cfg.httpClient.Do(req)
	if err != nil {
		return nil, nil, connectionError(ctx, attemptCtx, cfg.timeout, target, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, connectionError(ctx, attemptCtx, cfg.timeout, target, err)
	}
	return resp, body, nil
}

func connectionError(ctx, attemptCtx context.Context, timeout time.Duration, target string, err error) error {
	// Report what failed below the *url.Error, which repeats the method and URL.
	cause := err
	var urlErr *url.Error
	if errors.As(err, &urlErr) {
		cause = urlErr.Err
	}
	timedOut := ctx.Err() == nil && errors.Is(attemptCtx.Err(), context.DeadlineExceeded)
	return &ConnectionError{URL: target, Timeout: timedOut, timeout: timeout, Err: cause}
}

func retryDelay(method string, resp *http.Response, attempt, maxRetries int) (time.Duration, bool) {
	if attempt >= maxRetries {
		return 0, false
	}

	// A throttled request never reached the application, so it is safe to repeat
	// for every method. Server errors are only retried when repeating is harmless.
	status := resp.StatusCode
	retryable := status == http.StatusTooManyRequests ||
		(method == http.MethodGet && (status == 500 || status == 502 || status == 503 || status == 504))
	if !retryable {
		return 0, false
	}

	retryAfter, ok := parseRetryAfter(resp.Header.Get("Retry-After"))
	if !ok {
		return backoff(attempt), true
	}
	if retryAfter > maxRetryAfter {
		return 0, false
	}
	return retryAfter, true
}

// backoff returns 0.5s, 1s, 2s... capped at 8s, with jitter so parallel clients spread out.
func backoff(attempt int) time.Duration {
	delay := 8 * time.Second
	if attempt < 4 {
		delay = 500 * time.Millisecond << attempt
	}
	return time.Duration(float64(delay) * (1 - rand.Float64()*0.25))
}

func sleep(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func readEnv(name string) string {
	return strings.TrimSpace(os.Getenv(name))
}

func normalizeBaseURL(u string) string {
	u = strings.TrimRight(u, "/")
	u = strings.TrimSuffix(u, "/api/v2")
	return u
}

func methodOrGet(method string) string {
	if method == "" {
		return http.MethodGet
	}
	return strings.ToUpper(method)
}
