package truelayer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"time"

	"github.com/monzo/slog"

	"golang.org/x/oauth2"

	"github.com/arussellsaw/youneedaspreadsheet/pkg/token"
)

const (
	baseURL = "https://api.truelayer.com"
)

func GetClients(ctx context.Context, userID string) ([]*Client, error) {
	ts, err := token.ListByUser(ctx, userID, "truelayer", OauthConfig)
	if err != nil {
		return nil, err
	}
	var cs []*Client
	for _, t := range ts {
		cs = append(cs, &Client{
			userID: userID,
			t:      t,
			http: &http.Client{
				Transport: http.DefaultTransport,
				Timeout:   300 * time.Second,
			},
		})
	}
	return cs, nil
}

type Client struct {
	userID string
	t      *oauth2.Token
	http   *http.Client
}

func (c *Client) authRequest(r *http.Request) {
	r.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.t.AccessToken))
}

func (c *Client) Accounts(ctx context.Context) ([]Account, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/data/v1/accounts", baseURL),
		nil,
	)
	if err != nil {
		return nil, err
	}
	c.authRequest(req)
	res, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	response := struct {
		Results []Account `json:"results"`
		Error   string
	}{}
	err = json.NewDecoder(res.Body).Decode(&response)
	for i := range response.Results {
		response.Results[i].client = c
	}
	return response.Results, err
}

func (c *Client) Transactions(ctx context.Context, kind, accountID string, historic bool) ([]Transaction, error) {
	t := time.Now()
	txs := make(map[string]Transaction)
	for {
		var res []Transaction
		ts := t.Add(-88 * 24 * time.Hour).Format("2006-01-02T15:04:05Z")
		now := t.Format("2006-01-02T15:04:05Z")
		slog.Info(ctx, "Getting txs for %s %s %s", accountID, ts, now)
		err := c.doRequest(ctx, fmt.Sprintf("/data/v1/%s/%s/transactions?from=%s&to=%s", kind, accountID, ts, now), &res)
		if err != nil {
			return nil, err
		}
		slog.Info(ctx, "got %v transactions", len(res))
		if len(res) == 0 {
			break
		}
		for _, tx := range res {
			txs[tx.TransactionID] = tx
		}
		if !historic {
			break
		}
		t = t.AddDate(0, 0, -87)
	}
	var out []Transaction
	for _, tx := range txs {
		out = append(out, tx)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Timestamp < out[j].Timestamp
	})
	return out, nil
}

func (c *Client) Balance(ctx context.Context, kind, accountID string) (*Balance, error) {
	var res []Balance
	err := c.doRequest(ctx, fmt.Sprintf("/data/v1/%s/%s/balance", kind, accountID), &res)
	if err != nil {
		return nil, err
	}
	if len(res) != 1 {
		return nil, fmt.Errorf("unexpected length: %v", len(res))
	}
	return &res[0], err
}

func (c *Client) Metadata(ctx context.Context) (*Metadata, error) {
	var ms []Metadata
	err := c.doRequest(ctx, "/data/v1/me", &ms)
	if err != nil {
		return nil, err
	}
	if len(ms) == 0 {
		return nil, fmt.Errorf("not found")
	}
	return &ms[0], nil
}

func (c *Client) Cards(ctx context.Context) ([]Card, error) {
	var cs []Card
	err := c.doRequest(ctx, "/data/v1/cards", &cs)
	if err != nil {
		return nil, err
	}
	for i := range cs {
		cs[i].client = c
	}
	return cs, nil
}

func (c *Client) doRequest(ctx context.Context, path string, results interface{}) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		baseURL+path,
		nil,
	)
	if err != nil {
		return err
	}
	c.authRequest(req)
	res, err := c.http.Do(req)
	if err != nil {
		return err
	}
	response := struct {
		Results interface{} `json:"results"`
	}{}
	response.Results = results
	err = json.NewDecoder(res.Body).Decode(&response)
	return err
}

func Providers(ctx context.Context) ([]Provider, error) {
	res, err := http.Get("https://auth.truelayer.com/api/providers")
	if err != nil {
		return nil, err
	}
	var ps []Provider
	return ps, json.NewDecoder(res.Body).Decode(&ps)
}
