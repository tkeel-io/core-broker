package core

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/tkeel-io/core-broker/pkg/types"
	"github.com/tkeel-io/kit/log"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/pkg/errors"
)

type Client struct {
	daprClient dapr.Client
}

func NewCoreClient() (*Client, error) {
	client, err := dapr.NewClient()
	if err != nil {
		return nil, errors.Wrap(err, "init dapr client error")
	}
	return &Client{daprClient: client}, nil
}

type SubscriptionData struct {
	Mode       string `json:"mode,omitempty"`
	Source     string `json:"source,omitempty"`
	Filter     string `json:"filter,omitempty"`
	Topic      string `json:"topic,omitempty"`
	PubsubName string `json:"pubsub_name,omitempty"`
}

func (c *Client) Subscribe(subscriptionID, entityID, topic string) error {
	if subscriptionID == "" ||
		entityID == "" ||
		topic == "" {
		return errors.New("subscriptionID, entityID or topic is empty")
	}
	ctx := context.Background()
	filter := IntoFilterQuery(subscriptionID, entityID)
	subscriptionRequestData := SubscriptionData{
		Mode:       "realtime",
		Source:     "ignore",
		Filter:     filter,
		Topic:      topic,
		PubsubName: types.PubsubName,
	}

	methodName := CreateSubscriptionURL(subscriptionID, "admin", "dm", "SUBSCRIPTION")
	log.Debug("subscription ID:", subscriptionID)
	log.Debug("methodName:", methodName)
	log.Debug("Subscribe to Core data: ", subscriptionRequestData)

	contentData, err := json.Marshal(subscriptionRequestData)
	if err != nil {
		return errors.Wrap(err, "subscriptionRequestData marshal error")
	}

	content := &dapr.DataContent{
		Data:        contentData,
		ContentType: MimeJson,
	}

	if c, err := c.daprClient.InvokeMethodWithContent(ctx, AppID, methodName, http.MethodPost, content); err != nil {
		log.Error("invoke ", methodName, err)
		log.Error("invoke Response:", string(c))
		return errors.Wrap(err, "invoke method error")
	}
	return nil
}

func (c *Client) Unsubscribe(subscriptionID string) error {
	ctx := context.Background()
	methodName := CreateUnsubscriptionURL(subscriptionID, "admin", "dm", "SUBSCRIPTION")
	log.Debug("invoke unsubscribe to Core: ", methodName)
	if c, err := c.daprClient.InvokeMethod(ctx, AppID, methodName, http.MethodDelete); err != nil {
		log.Error("invoke ", methodName, " with ", http.MethodDelete, err)
		log.Error("invoke Response:", string(c))
		return err
	}
	return nil
}

const _InsertQueryTemplate = "insert into %s select %s.*"

func IntoFilterQuery(to string, from string) string {
	return fmt.Sprintf(_InsertQueryTemplate, to, from)
}
