package api

import "context"

// GetUser retrieves the authenticated user's information
func (c *Client) GetUser(ctx context.Context) (*User, error) {
	var result User
	if err := c.Get(ctx, "/user", &result); err != nil {
		return nil, err
	}
	return &result, nil
}
