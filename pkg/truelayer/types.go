package truelayer

import (
	"context"
	"time"
)

type Account struct {
	UpdateTimestamp time.Time     `json:"update_timestamp"`
	AccountID       string        `json:"account_id"`
	AccountType     string        `json:"account_type"`
	DisplayName     string        `json:"display_name"`
	Currency        string        `json:"currency"`
	AccountNumber   AccountNumber `json:"account_number"`
	Provider        Provider      `json:"provider"`

	client *Client
}

func (a Account) ID() string {
	return a.AccountID
}

func (a Account) Name() string {
	return a.DisplayName
}

func (a Account) Transactions(ctx context.Context, historic bool) ([]Transaction, error) {
	return a.client.Transactions(ctx, "accounts", a.AccountID, historic)
}

func (a Account) Balance(ctx context.Context) (*Balance, error) {
	return a.client.Balance(ctx, "accounts", a.AccountID)
}

type AccountNumber struct {
	Iban     string `json:"iban"`
	Number   string `json:"number"`
	SortCode string `json:"sort_code"`
	SwiftBic string `json:"swift_bic"`
}

type Transaction struct {
	TransactionID             string         `json:"transaction_id"`
	Timestamp                 string         `json:"timestamp"`
	Description               string         `json:"description"`
	Amount                    float64        `json:"amount"`
	Currency                  string         `json:"currency"`
	TransactionType           string         `json:"transaction_type"`
	TransactionCategory       string         `json:"transaction_category"`
	TransactionClassification []string       `json:"transaction_classification"`
	MerchantName              string         `json:"merchant_name"`
	RunningBalance            RunningBalance `json:"running_balance"`
	Meta                      Meta           `json:"meta"`
}

type RunningBalance struct {
	Amount   float64 `json:"amount"`
	Currency string  `json:"currency"`
}

type Meta struct {
	BankTransactionID           string `json:"bank_transaction_id"`
	ProviderTransactionCategory string `json:"provider_transaction_category"`
}

type Balance struct {
	Currency        string    `json:"currency"`
	Available       float64   `json:"available"`
	Current         float64   `json:"current"`
	Overdraft       float64   `json:"overdraft"`
	UpdateTimestamp time.Time `json:"update_timestamp"`
}

type Metadata struct {
	ClientID               string    `json:"client_id"`
	CredentialsID          string    `json:"credentials_id"`
	ConsentStatus          string    `json:"consent_status"`
	ConsentStatusUpdatedAt time.Time `json:"consent_status_updated_at"`
	ConsentCreatedAt       time.Time `json:"consent_created_at"`
	ConsentExpiresAt       time.Time `json:"consent_expires_at"`
	Provider               Provider  `json:"provider"`
	Scopes                 []string  `json:"scopes"`
	PrivacyPolicy          string    `json:"privacy_policy"`
}

type Provider struct {
	DisplayName string `json:"display_name"`
	LogoURI     string `json:"logo_uri"`
	LogoURL     string `json:"logo_url"`
	ProviderID  string `json:"provider_id"`
}

type Card struct {
	AccountID         string    `json:"account_id"`
	CardNetwork       string    `json:"card_network"`
	CardType          string    `json:"card_type"`
	Currency          string    `json:"currency"`
	DisplayName       string    `json:"display_name"`
	PartialCardNumber string    `json:"partial_card_number"`
	NameOnCard        string    `json:"name_on_card"`
	ValidFrom         string    `json:"valid_from"`
	ValidTo           string    `json:"valid_to"`
	UpdateTimestamp   time.Time `json:"update_timestamp"`
	Provider          Provider  `json:"provider"`

	client *Client
}

func (c Card) ID() string {
	return c.AccountID
}

func (c Card) Name() string {
	return c.DisplayName
}

func (c Card) Transactions(ctx context.Context, historic bool) ([]Transaction, error) {
	return c.client.Transactions(ctx, "cards", c.AccountID, historic)
}

func (c Card) Balance(ctx context.Context) (*Balance, error) {
	b, err := c.client.Balance(ctx, "cards", c.AccountID)
	if err != nil {
		return nil, err
	}
	b.Available = b.Available * -1
	b.Current = b.Current * -1
	return b, nil
}

type Balancer interface {
	Balance(context.Context) (*Balance, error)
}

type Transactioner interface {
	Transactions(context.Context, bool) ([]Transaction, error)
}

type AbstractAccount interface {
	ID() string
	Name() string
	Balancer
	Transactioner
}
