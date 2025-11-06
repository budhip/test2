package dddnotification

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"bitbucket.org/Amartha/go-megatron/internal/config"
	xlog "bitbucket.org/Amartha/go-x/log"
	"bitbucket.org/Amartha/go-x/log/ctxdata"

	"github.com/go-resty/resty/v2"
	"github.com/newrelic/go-agent/v3/newrelic"
)

type Client interface {
	SendMessageToSlack(ctx context.Context, message MessageData) error
}

type client struct {
	cfg        *config.Configuration
	httpClient *resty.Client
}

func New(cfg *config.Configuration) Client {
	retryWaitTime := time.Duration(cfg.DDDNotification.RetryCount) * time.Millisecond
	restyClient := resty.New().
		SetRetryCount(cfg.DDDNotification.RetryCount).
		SetRetryWaitTime(retryWaitTime)
	restyClient.SetTransport(newrelic.NewRoundTripper(restyClient.GetClient().Transport))

	return &client{cfg: cfg, httpClient: restyClient}
}

func (c *client) SendMessageToSlack(ctx context.Context, message MessageData) error {
	path := "/api/v1/slack/send-message"
	url := fmt.Sprintf("%s%s", c.cfg.DDDNotification.BaseURL, path)

	payload := PayloadNotification{
		Title:        c.cfg.DDDNotification.TitleBot,
		Service:      c.cfg.App.Name,
		SlackChannel: c.cfg.DDDNotification.SlackChannel,
		Data:         message,
	}

	xlog.Infof(ctx, "send request to %s with body %v", url, payload)

	resp, err := c.httpClient.R().SetContext(context.WithoutCancel(ctx)).
		SetHeader("Accept", "application/json;  charset=utf-8").
		SetHeader("Cache-Control", "no-cache").
		SetHeader("X-Correlation-Id", ctxdata.GetCorrelationId(ctx)).
		SetHeader("User-Agent", c.cfg.App.Name).
		SetBody(payload).
		Post(url)
	if err != nil {
		return fmt.Errorf("error send request to %s: %w", url, err)
	}

	var response *ResponseSendMessage
	if err = json.Unmarshal(resp.Body(), &response); err != nil {
		return fmt.Errorf("error unmarshal response from %s: %w", url, err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("error response from %s: %s", url, response.Message)
	}

	return nil
}
