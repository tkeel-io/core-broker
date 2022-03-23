package deviceutil

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/tkeel-io/kit/log"
)

type Service string

func (s Service) String() string {
	return string(s)
}

const (
	DeviceSearch Service = "http://localhost:3500/v1.0/invoke/keel/method/apis/tkeel-device/v1/search"
	EntitySearch Service = "http://localhost:3500/v1.0/invoke/keel/method/apis/core/v1/entities/search"
)

type Client struct {
	http  *http.Client
	token string
}

func NewClient(token string) *Client {
	return &Client{
		token: token,
		http:  &http.Client{},
	}
}

func (c Client) SearchDefault(url Service, conditions Conditions) ([]byte, error) {
	searchRequest := SearchRequest{
		PageNum:    1,
		PageSize:   5000,
		Conditions: conditions,
	}
	content, err := json.Marshal(&searchRequest)
	log.Info("Device Search Request URL:", url)
	log.Info("Device Search Request Body:", string(content))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", c.token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

func (c Client) Search(url Service, conditions Conditions, query string, num, size int32) ([]byte, error) {
	if num == 0 {
		num = 1
	}
	searchRequest := SearchRequest{
		PageNum:    num,
		PageSize:   size,
		Conditions: conditions,
		Query:      query,
	}
	content, err := json.Marshal(&searchRequest)
	log.Info("Device Search Request URL:", url)
	log.Info("Device Search Request Body:", string(content))
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, url.String(), bytes.NewBuffer(content))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", c.token)
	req.Header.Add("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ioutil.ReadAll(resp.Body)
}

type SearchRequest struct {
	PageNum      int32      `json:"page_num"`
	PageSize     int32      `json:"page_size"`
	OrderBy      string     `json:"order_by,omitempty"`
	IsDescending bool       `json:"is_descending,omitempty"`
	Query        string     `json:"query,omitempty"`
	Conditions   Conditions `json:"condition"`
}

type Conditions []ConditionQuery

type ConditionQuery struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

func NewQuery(field, operator, value string) ConditionQuery {
	return ConditionQuery{
		Field:    field,
		Operator: operator,
		Value:    value,
	}
}

func GroupQuery(value string, values ...string) ConditionQuery {
	if len(values) != 0 {
		values = append([]string{value}, values...)
		value = strings.Join(values, ",")
	}
	return ConditionQuery{
		Field:    "sysField._spacePath",
		Operator: "$wildcard",
		Value:    value,
	}
}

func GroupTypeQuery() ConditionQuery {
	return ConditionQuery{
		Field:    "type",
		Operator: "$eq",
		Value:    "group",
	}
}

func TemplateQuery(value string, values ...string) ConditionQuery {
	if len(values) != 0 {
		values = append([]string{value}, values...)
		value = strings.Join(values, ",")
	}
	return ConditionQuery{
		Field:    "basicInfo.templateId",
		Operator: "$eq",
		Value:    value,
	}
}

func TemplateTypeQuery() ConditionQuery {
	return ConditionQuery{
		Field:    "type",
		Operator: "$eq",
		Value:    "template",
	}
}

func DeviceTypeQuery() ConditionQuery {
	return ConditionQuery{
		Field:    "type",
		Operator: "$eq",
		Value:    "device",
	}
}

func DeviceQuery(id string) ConditionQuery {
	return ConditionQuery{
		Field:    "id",
		Operator: "$eq",
		Value:    id,
	}
}

func EqQuery(field, value string) ConditionQuery {
	return ConditionQuery{
		Field:    field,
		Operator: "$eq",
		Value:    value,
	}
}

func WildcardQuery(field, value string) ConditionQuery {
	return ConditionQuery{
		Field:    field,
		Operator: "$wildcard",
		Value:    value,
	}
}
