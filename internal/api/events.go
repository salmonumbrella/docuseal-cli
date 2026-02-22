package api

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
)

// ListFormEvents retrieves form events
func (c *Client) ListFormEvents(ctx context.Context, eventType string, limit int) ([]Event, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}

	path := fmt.Sprintf("/events/form/%s", eventType)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result []Event
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ListSubmissionEvents retrieves submission events
func (c *Client) ListSubmissionEvents(ctx context.Context, eventType string, submissionID int, limit int) ([]Event, error) {
	params := url.Values{}
	if limit > 0 {
		params.Set("limit", strconv.Itoa(limit))
	}
	if submissionID > 0 {
		params.Set("submission_id", strconv.Itoa(submissionID))
	}

	path := fmt.Sprintf("/events/submission/%s", eventType)
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	var result []Event
	if err := c.Get(ctx, path, &result); err != nil {
		return nil, err
	}
	return result, nil
}
