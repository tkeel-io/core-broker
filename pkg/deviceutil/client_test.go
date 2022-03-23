package deviceutil

import (
	"fmt"
	"testing"

	pb "github.com/tkeel-io/core-broker/api/subscribe/v1"

	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	token := "Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0a2VlbCIsImV4cCI6MTY0NTU5MTcxNiwic3ViIjoidXNyLTMzNzM3OTQ1YzJiNzE4ZGI0YzMwOWQ2MzNkMmYifQ.ps6PhgLqJviE0ePG3vOTqnQu5NzYeQvicAB3DoRrMS8l1kNV5I9L0U9pgRJ3BW4vUQrYP6_jklNHvAvVCFsTRg"
	c := NewClient(token)

	bytes, err := c.SearchDefault(DeviceSearch, Conditions{GroupQuery("testGroupABC"), DeviceTypeQuery()})
	fmt.Println("Response Content:", string(bytes))
	assert.NoError(t, err)

	data, err := ParseSearchResponse(bytes)
	assert.NoError(t, err)
	fmt.Println("Response:", data)
	fmt.Println("Token:", token)
	fmt.Println("URL:", DeviceSearch)
}

func TestNewTemplateQuery(t *testing.T) {
	token := "Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0a2VlbCIsImV4cCI6MTY0NTU5MTcxNiwic3ViIjoidXNyLTMzNzM3OTQ1YzJiNzE4ZGI0YzMwOWQ2MzNkMmYifQ.ps6PhgLqJviE0ePG3vOTqnQu5NzYeQvicAB3DoRrMS8l1kNV5I9L0U9pgRJ3BW4vUQrYP6_jklNHvAvVCFsTRg"
	c := NewClient(token)

	bytes, err := c.SearchDefault(DeviceSearch, Conditions{TemplateQuery("4a8eac20-699c-4f83-a2b4-da5233304509"), DeviceTypeQuery()})
	fmt.Println("Response Content:", string(bytes))
	assert.NoError(t, err)

	data, err := ParseSearchResponse(bytes)
	assert.NoError(t, err)
	fmt.Println("Response:", data)
	fmt.Println("Token:", token)
	fmt.Println("URL:", DeviceSearch)
}

func TestNewCoreSearch(t *testing.T) {
	token := "Bearer eyJhbGciOiJIUzUxMiIsInR5cCI6IkpXVCJ9.eyJhdWQiOiJ0a2VlbCIsImV4cCI6MTY0NTU5MTcxNiwic3ViIjoidXNyLTMzNzM3OTQ1YzJiNzE4ZGI0YzMwOWQ2MzNkMmYifQ.ps6PhgLqJviE0ePG3vOTqnQu5NzYeQvicAB3DoRrMS8l1kNV5I9L0U9pgRJ3BW4vUQrYP6_jklNHvAvVCFsTRg"
	c := NewClient(token)
	id := "a8e92c6d-0f73-4f7a-8b85-0f110155eed2"
	bytes, err := c.SearchDefault(EntitySearch, Conditions{DeviceQuery(id)})
	if err != nil {
		fmt.Println(err)
		return
	}
	content := string(bytes)
	fmt.Println(content)
	resp, err := ParseSearchEntityResponse(bytes)
	if err != nil {
		fmt.Println(err)
		return
	}
	if len(resp.Data.Items) == 0 {
		fmt.Println("device not found:", id)
		return
	}
	entity := &pb.Entity{
		ID:        id,
		Name:      resp.Data.Items[0].Properties.BasicInfo.Name,
		Template:  resp.Data.Items[0].Properties.BasicInfo.TemplateName,
		Group:     resp.Data.Items[0].Properties.BasicInfo.ParentName,
		Status:    resp.Data.Items[0].Properties.SysField.Status,
		UpdatedAt: resp.Data.Items[0].Properties.SysField.UpdatedAt,
	}
	assert.NoError(t, err)
	fmt.Println("entity:", entity)
	fmt.Println("Token:", token)
}
