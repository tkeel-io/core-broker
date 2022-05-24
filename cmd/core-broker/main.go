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

package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/tkeel-io/core-broker/pkg/metrics"
	"github.com/tkeel-io/core-broker/pkg/model"
	"github.com/tkeel-io/core-broker/pkg/server"
	"github.com/tkeel-io/core-broker/pkg/service"
	"github.com/tkeel-io/kit/app"
	"github.com/tkeel-io/kit/log"
	"github.com/tkeel-io/kit/transport"

	// User import.

	Dapr_v1 "github.com/tkeel-io/core-broker/api/dapr"
	metrics_v1 "github.com/tkeel-io/core-broker/api/metrics/v1"
	Subscribe_v1 "github.com/tkeel-io/core-broker/api/subscribe/v1"
	Topic_v1 "github.com/tkeel-io/core-broker/api/topic/v1"
	Entity_v1 "github.com/tkeel-io/core-broker/api/ws/v1"

	openapi "github.com/tkeel-io/core-broker/api/openapi/v1"
)

var (
	// Name app.
	Name string
	// HTTPAddr string.
	HTTPAddr string
	// GRPCAddr string.
	GRPCAddr string
)

func init() {
	flag.StringVar(&Name, "name", "core-broker", "app name.")
	flag.StringVar(&HTTPAddr, "http_addr", ":31234", "http listen address.")
	flag.StringVar(&GRPCAddr, "grpc_addr", ":31233", "grpc listen address.")
}

func main() {
	flag.Parse()

	httpSrv := server.NewHTTPServer(HTTPAddr)
	grpcSrv := server.NewGRPCServer(GRPCAddr)
	serverList := []transport.Server{httpSrv, grpcSrv}

	app := app.New(Name,
		&log.Conf{
			App:   Name,
			Level: "debug",
			Dev:   true,
		},
		serverList...,
	)

	{ // User service
		OpenapiSrv := service.NewOpenapiService()
		openapi.RegisterOpenapiHTTPServer(httpSrv.Container, OpenapiSrv)
		openapi.RegisterOpenapiServer(grpcSrv.GetServe(), OpenapiSrv)

		EntitySrv := service.NewEntityService()
		go EntitySrv.Run()
		Entity_v1.RegisterEntityHTTPServer(httpSrv.Container, EntitySrv)

		TopicSrv := service.NewTopicService()
		Topic_v1.RegisterTopicHTTPServer(httpSrv.Container, TopicSrv)
		Topic_v1.RegisterTopicServer(grpcSrv.GetServe(), TopicSrv)

		DaprSubscribeSrv := service.NewDaprSubscribeService()
		Dapr_v1.RegisterSubscribeHTTPServer(httpSrv.Container, DaprSubscribeSrv)
		Dapr_v1.RegisterSubscribeServer(grpcSrv.GetServe(), DaprSubscribeSrv)

		SubscribeSrv := service.NewSubscribeService()
		Subscribe_v1.RegisterSubscribeHTTPServer(httpSrv.Container, SubscribeSrv)
		Subscribe_v1.RegisterSubscribeServer(grpcSrv.GetServe(), SubscribeSrv)

		metricsSrv := service.NewMetricsService(metrics.CollectorSubscribeNum, metrics.CollectorSubscribeEntitiesNum)
		metrics_v1.RegisterMetricsHTTPServer(httpSrv.Container, metricsSrv)
	}

	if err := app.Run(context.TODO()); err != nil {
		panic(err)
	}

	sub := model.Subscribe{}
	sub.InitMetrics()

	subEntities := model.SubscribeEntities{}
	subEntities.InitMetrics()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, os.Interrupt)
	<-stop

	if err := app.Stop(context.TODO()); err != nil {
		panic(err)
	}
}
