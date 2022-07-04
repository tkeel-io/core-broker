/*
Copyright 2021 The tKeel Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package service

import (
	"context"

	v1 "github.com/tkeel-io/core-broker/api/openapi/v1"
	"github.com/tkeel-io/core-broker/pkg/util"
	openapi_v1 "github.com/tkeel-io/tkeel-interface/openapi/v1"
	"google.golang.org/protobuf/types/known/emptypb"
)

// OpenapiService is a openapi service.
type OpenapiService struct {
	v1.UnimplementedOpenapiServer
}

// NewOpenapiService new a openapi service.
func NewOpenapiService() *OpenapiService {
	return &OpenapiService{
		UnimplementedOpenapiServer: v1.UnimplementedOpenapiServer{},
	}
}

// AddonsIdentify implements AddonsIdentify.OpenapiServer.
func (s *OpenapiService) AddonsIdentify(ctx context.Context, in *openapi_v1.AddonsIdentifyRequest) (*openapi_v1.AddonsIdentifyResponse, error) {
	return &openapi_v1.AddonsIdentifyResponse{
		Res: util.BadRequestResult("not declare addons"),
	}, nil
}

// Identify implements Identify.OpenapiServer.
func (s *OpenapiService) Identify(ctx context.Context, in *emptypb.Empty) (*openapi_v1.IdentifyResponse, error) {
	profiles := map[string]*openapi_v1.ProfileSchema{
		"subscribe_max":          {Type: "number", Title: "最大订阅数", Default: 5, MultipleOf: 1, Maximum: 10, Minimum: 1},
		"subscribe_entities_max": {Type: "number", Title: "最大订阅设备数", Default: 1000, MultipleOf: 1, Maximum: 10000, Minimum: 0},
	}
	return &openapi_v1.IdentifyResponse{
		Res:                     util.OKResult(),
		PluginId:                "core-broker",
		Version:                 "0.4.1",
		TkeelVersion:            "v0.4.0",
		DisableManualActivation: true,
		Profiles:                profiles,
	}, nil
}

// Status implements Status.OpenapiServer.
func (s *OpenapiService) Status(ctx context.Context, in *emptypb.Empty) (*openapi_v1.StatusResponse, error) {
	return &openapi_v1.StatusResponse{
		Res:    util.OKResult(),
		Status: openapi_v1.PluginStatus_RUNNING,
	}, nil
}

// TenantEnable implements TenantEnable.OpenapiServer.
func (s *OpenapiService) TenantEnable(ctx context.Context, in *openapi_v1.TenantEnableRequest) (*openapi_v1.TenantEnableResponse, error) {
	return &openapi_v1.TenantEnableResponse{
		Res: util.OKResult(),
	}, nil
}

// TenantDisable implements TenantDisable.OpenapiServer.
func (s *OpenapiService) TenantDisable(ctx context.Context, in *openapi_v1.TenantDisableRequest) (*openapi_v1.TenantDisableResponse, error) {
	return &openapi_v1.TenantDisableResponse{
		Res: util.OKResult(),
	}, nil
}
