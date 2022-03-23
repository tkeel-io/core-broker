package service

import (
	"context"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/tkeel-io/core-broker/pkg/auth"

	"github.com/pkg/errors"
	pb "github.com/tkeel-io/core-broker/api/subscribe/v1"
	"github.com/tkeel-io/core-broker/pkg/deviceutil"
	"github.com/tkeel-io/core-broker/pkg/model"
	"github.com/tkeel-io/core-broker/pkg/pagination"
	"github.com/tkeel-io/core-broker/pkg/subscribeuril"
	"github.com/tkeel-io/kit/log"
	"gorm.io/gorm"
)

const (
	SuccessStatus     = "SUCCESS"
	ErrPartialFailure = "PARTIAL FAILURE"

	_DefaultSubscribeTitle       = "我的订阅"
	_DefaultSubscribeDescription = "这是我的默认订阅，该订阅无法被删除，但是可以被修改。"
)

var (
	ErrDeviceNotFound = errors.New("device not found")
)

type SubscribeService struct {
	pb.UnimplementedSubscribeServer
}

func NewSubscribeService() *SubscribeService {
	if err := model.Setup(); err != nil {
		log.Fatal(err)
	}

	return &SubscribeService{}
}

func (s *SubscribeService) SubscribeEntitiesByIDs(ctx context.Context, req *pb.SubscribeEntitiesByIDsRequest) (*pb.SubscribeEntitiesByIDsResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	subscribe := model.Subscribe{Model: gorm.Model{ID: uint(req.Id)}, UserID: authUser.ID}
	validateSubscribeResult := model.DB().First(&subscribe)
	if validateSubscribeResult.RowsAffected == 0 {
		err = errors.Wrap(validateSubscribeResult.Error, "subscribe and user ID mismatch")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}

	resp := &pb.SubscribeEntitiesByIDsResponse{
		Id:     req.GetId(),
		Status: SuccessStatus,
	}
	if len(req.Entities) == 0 {
		return resp, nil
	}

	records := s.createSubscribeEntitiesRecords(req.Entities, &subscribe)
	result := model.DB().Preload("Subscribe").Create(&records)
	if result.Error != nil {
		log.Error("err:", result.Error)
		mysqlErr, ok := result.Error.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			return nil, pb.ErrDuplicateCreate()
		}
		return nil, pb.ErrInternalError()
	}
	if result.RowsAffected != int64(len(records)) {
		return nil, pb.ErrSomeDuplicateCreate()
	}

	return resp, nil
}

func (s *SubscribeService) SubscribeEntitiesByGroups(ctx context.Context, req *pb.SubscribeEntitiesByGroupsRequest) (*pb.SubscribeEntitiesByGroupsResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	subscribe := model.Subscribe{Model: gorm.Model{ID: uint(req.Id)}, UserID: authUser.ID}
	validateSubscribeResult := model.DB().First(&subscribe)
	if validateSubscribeResult.RowsAffected == 0 {
		err = errors.Wrap(validateSubscribeResult.Error, "subscribe and user ID mismatch")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}

	resp := &pb.SubscribeEntitiesByGroupsResponse{
		Id:     req.Id,
		Status: SuccessStatus,
	}
	ids, err := s.getDeviceEntitiesIDsFromGroups(ctx, req.Groups, authUser.Token)
	if err != nil {
		err = errors.Wrap(err, "get device entities IDs from groups IDs error")
		log.Error("err:", err)
		return nil, pb.ErrInternalQuery()
	}
	log.Debug("get device entities IDs from groups IDs:", ids)
	if len(ids) == 0 {
		log.Debug("no device entities IDs found")
		return nil, pb.ErrDeviceNotFound()
	}
	records := s.createSubscribeEntitiesRecords(ids, &subscribe)
	log.Info("create subscribe entities records:", records)
	result := model.DB().Create(&records)
	if result.Error != nil {
		log.Error("err:", result.Error)
		mysqlErr, ok := result.Error.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			return nil, pb.ErrDuplicateCreate()
		}
		return nil, pb.ErrInternalError()
	}
	if result.RowsAffected != int64(len(records)) {
		return nil, pb.ErrSomeDuplicateCreate()
	}
	return resp, nil
}

func (s *SubscribeService) SubscribeEntitiesByModels(ctx context.Context, req *pb.SubscribeEntitiesByModelsRequest) (*pb.SubscribeEntitiesByModelsResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	subscribe := model.Subscribe{Model: gorm.Model{ID: uint(req.Id)}, UserID: authUser.ID}
	validateSubscribeResult := model.DB().First(&subscribe)
	if validateSubscribeResult.RowsAffected == 0 {
		err = errors.Wrap(validateSubscribeResult.Error, "subscribe and user ID mismatch")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}

	resp := &pb.SubscribeEntitiesByModelsResponse{
		Id:     req.Id,
		Status: SuccessStatus,
	}
	ids, err := s.getDeviceEntitiesIDsFromTemplates(ctx, req.Models, authUser.Token)
	if err != nil {
		err = errors.Wrap(err, "get device entities IDs from models IDs error")
		log.Error("err:", err)
		return nil, err
	}
	log.Debug("get device entities IDs from groups IDs:", ids)
	if len(ids) == 0 {
		log.Debug("no device entities IDs found")
		return nil, pb.ErrDeviceNotFound()
	}
	records := s.createSubscribeEntitiesRecords(ids, &subscribe)
	result := model.DB().Create(&records)
	if result.Error != nil {
		log.Error("err:", result.Error)
		mysqlErr, ok := result.Error.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			return nil, pb.ErrDuplicateCreate()
		}
		return nil, pb.ErrInternalError()
	}
	if result.RowsAffected != int64(len(records)) {
		return nil, pb.ErrSomeDuplicateCreate()
	}
	return resp, nil
}

func (s *SubscribeService) UnsubscribeEntitiesByIDs(ctx context.Context, req *pb.UnsubscribeEntitiesByIDsRequest) (*pb.UnsubscribeEntitiesByIDsResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	subscribe := model.Subscribe{Model: gorm.Model{ID: uint(req.Id)}, UserID: authUser.ID}
	if model.DB().First(&subscribe).RowsAffected == 0 {
		err = errors.New("subscribe and user ID mismatch")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}

	resp := &pb.UnsubscribeEntitiesByIDsResponse{
		Id:     req.Id,
		Status: SuccessStatus,
	}

	tx := model.DB().Begin()
	for _, entityID := range req.Entities {
		subscribeEntity := model.SubscribeEntities{
			Subscribe:   subscribe,
			EntityID:    entityID,
			SubscribeID: subscribe.ID,
			UniqueKey:   subscribeuril.GenerateSubscribeTopic(subscribe.ID, entityID),
		}
		result := tx.
			Where("subscribe_id = ?", subscribeEntity.SubscribeID).
			Where("entity_id = ?", subscribeEntity.EntityID).
			Where("unique_key = ?", subscribeEntity.UniqueKey).
			Delete(&subscribeEntity)
		if result.Error != nil {
			log.Error("err:", result.Error)
			tx.Rollback()
			return nil, pb.ErrInternalError()
		}
	}
	tx.Commit()

	return resp, nil
}

func (s *SubscribeService) ListSubscribeEntities(ctx context.Context, req *pb.ListSubscribeEntitiesRequest) (*pb.ListSubscribeEntitiesResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	subscribe := model.Subscribe{Model: gorm.Model{ID: uint(req.Id)}, UserID: authUser.ID}
	validateSubscribeResult := model.DB().First(&subscribe)
	if validateSubscribeResult.RowsAffected == 0 {
		err = errors.Wrap(validateSubscribeResult.Error, "subscribe and user ID mismatch")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}

	page, err := pagination.Parse(req)
	if err != nil {
		log.Error("err:", err)
		return nil, pb.ErrInvalidArgument()
	}

	conditions := make(deviceutil.Conditions, 0)
	conditions = append(conditions, deviceutil.EqQuery(Owner, authUser.ID))
	conditions = append(conditions, deviceutil.WildcardQuery(SubscribePath, subscribe.Endpoint))
	data, err := s.getEntitiesByConditions(conditions, authUser.Token, page.KeyWords, int32(page.Num), int32(page.Size))
	if err != nil {
		log.Error("err:", err)
		if errors.Is(err, ErrDeviceNotFound) {
			return nil, pb.ErrDeviceNotFound()
		}
		return nil, pb.ErrInternalQuery()
	}

	resp := &pb.ListSubscribeEntitiesResponse{}
	page.SetTotal(uint(len(data)))
	err = page.FillResponse(resp)
	if err != nil {
		log.Error("err:", err)
		return nil, pb.ErrList()
	}
	resp.Data = data

	return resp, nil
}

func (s *SubscribeService) CreateSubscribe(ctx context.Context, req *pb.CreateSubscribeRequest) (*pb.CreateSubscribeResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("get auth user err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	if strings.Contains(req.Title, "@") ||
		strings.Contains(req.Title, ",") {
		err = errors.New("title contain illegal characters")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	sub := model.Subscribe{
		UserID:      authUser.ID,
		Title:       req.Title,
		Description: req.Description,
	}

	// TODO: lock the table
	var count string
	findResult := model.DB().Model(&model.Subscribe{}).Select("1").
		Where(&model.Subscribe{UserID: authUser.ID, IsDefault: true}).
		Limit(1).
		Find(&count)
	if errors.Is(
		findResult.Error,
		gorm.ErrRecordNotFound,
	) || findResult.RowsAffected == 0 {
		sub.IsDefault = true
	}

	if err = model.DB().Create(&sub).Error; err != nil {
		log.Error("err:", err)
		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			return nil, pb.ErrDuplicateCreate()
		}
		return nil, pb.ErrInternalError()
	}

	return &pb.CreateSubscribeResponse{
		Id:          uint64(sub.ID),
		Title:       sub.Title,
		Description: sub.Description,
		Endpoint:    sub.Endpoint,
		IsDefault:   sub.IsDefault,
	}, nil
}

func (s *SubscribeService) UpdateSubscribe(ctx context.Context, req *pb.UpdateSubscribeRequest) (*pb.UpdateSubscribeResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	subscribe := model.Subscribe{Model: gorm.Model{ID: uint(req.Id)}, UserID: authUser.ID}
	validateSubscribeResult := model.DB().First(&subscribe)
	if validateSubscribeResult.RowsAffected == 0 {
		err = errors.Wrap(validateSubscribeResult.Error, "subscribe and user ID mismatch")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}

	subscribe.Title = req.Title
	subscribe.Description = req.Description

	if err = model.DB().Save(&subscribe).Error; err != nil {
		err = errors.Wrap(err, "update subscribe info err")
		log.Error("err:", err)
		return nil, pb.ErrInternalError()
	}

	resp := &pb.UpdateSubscribeResponse{
		Id:          uint64(subscribe.ID),
		Title:       subscribe.Title,
		Description: subscribe.Description,
		Endpoint:    subscribe.Endpoint,
		IsDefault:   subscribe.IsDefault,
	}
	return resp, nil
}

func (s *SubscribeService) DeleteSubscribe(ctx context.Context, req *pb.DeleteSubscribeRequest) (*pb.DeleteSubscribeResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	log.Debugf("user %s starting to delete subscribe %d", authUser.ID, req.Id)
	subscribe := model.Subscribe{}
	validateSubscribeResult := model.DB().Model(&subscribe).
		Where("id = ?", req.Id).
		Where("user_id = ?", authUser.ID).
		First(&subscribe)
	if validateSubscribeResult.RowsAffected == 0 {
		err = errors.Wrap(validateSubscribeResult.Error, "subscribe and user ID mismatch")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}

	if err = model.DB().Delete(&subscribe).Error; err != nil {
		if errors.Is(err, model.ErrUndeleteable) {
			return nil, pb.ErrTryToDeleteDefaultSubscribe()
		}
		err = errors.Wrap(err, "delete subscribe err")
		log.Error("err:", err)
		return nil, pb.ErrInternalError()
	}

	return &pb.DeleteSubscribeResponse{Id: req.Id}, nil
}

func (s *SubscribeService) GetSubscribe(ctx context.Context, req *pb.GetSubscribeRequest) (*pb.GetSubscribeResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	subscribe := model.Subscribe{Model: gorm.Model{ID: uint(req.Id)}, UserID: authUser.ID}
	validateSubscribeResult := model.DB().First(&subscribe)
	if validateSubscribeResult.RowsAffected == 0 {
		err = errors.Wrap(validateSubscribeResult.Error, "subscribe and user ID mismatch")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}

	var count int64
	model.DB().Model(&model.SubscribeEntities{}).Where("subscribe_id = ?", subscribe.ID).Count(&count)

	resp := &pb.GetSubscribeResponse{
		Id:          uint64(subscribe.ID),
		Title:       subscribe.Title,
		Description: subscribe.Description,
		Endpoint:    model.AMQPAddressString(subscribe.Endpoint),
		Count:       uint64(count),
		CreatedAt:   subscribe.CreatedAt.Unix(),
		UpdatedAt:   subscribe.UpdatedAt.Unix(),
		IsDefault:   subscribe.IsDefault,
	}
	return resp, nil
}

func (s *SubscribeService) ListSubscribe(ctx context.Context, req *pb.ListSubscribeRequest) (*pb.ListSubscribeResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	page, err := pagination.Parse(req)
	if err != nil {
		log.Error("parse request page info error:", err)
		return nil, pb.ErrInvalidArgument()
	}
	var subscribes []model.Subscribe
	result := &gorm.DB{Error: errors.New("db query error")}
	subscribeCondition := model.Subscribe{UserID: authUser.ID}
	if page.Required() {
		result = model.Paginate(&subscribes, page, &subscribeCondition)
	} else {
		result = model.ListAll(&subscribes, &subscribeCondition)
	}
	if result.Error != nil {
		log.Error("err:", result.Error)
		return nil, pb.ErrInternalError()
	}

	data := make([]*pb.SubscribeObject, 0, len(subscribes))
	for i := range subscribes {
		data = append(data, &pb.SubscribeObject{
			Id:          uint64(subscribes[i].ID),
			Title:       subscribes[i].Title,
			Description: subscribes[i].Description,
			Endpoint:    model.AMQPAddressString(subscribes[i].Endpoint),
			IsDefault:   subscribes[i].IsDefault,
		})
	}

	resp := &pb.ListSubscribeResponse{}
	// TODO: for template create default subscribe
	if len(subscribes) == 0 {
		createRequest := &pb.CreateSubscribeRequest{
			Title:       _DefaultSubscribeTitle,
			Description: _DefaultSubscribeDescription,
		}
		subscribeResponse, err := s.CreateSubscribe(ctx, createRequest)
		if err != nil {
			log.Error("create default subscribe failed:", err)
			return nil, pb.ErrInternalError()
		}
		data = append(data, &pb.SubscribeObject{
			Id:          subscribeResponse.Id,
			Title:       subscribeResponse.Title,
			Description: subscribeResponse.Description,
			Endpoint:    subscribeResponse.Endpoint,
			IsDefault:   subscribeResponse.IsDefault,
		})
	}

	var count int64
	if err = model.Count(&count, &subscribeCondition, &subscribeCondition).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Error("err:", err)
			return nil, err
		}
		count = 0
	}
	page.SetTotal(uint(count))
	if err = page.FillResponse(resp); err != nil {
		log.Error("err:", err)
		return nil, err
	}

	resp.Data = data

	return resp, nil
}

func (s *SubscribeService) ChangeSubscribed(ctx context.Context, req *pb.ChangeSubscribedRequest) (*pb.ChangeSubscribedResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	if len(req.SelectedIds) == 0 {
		return nil, pb.ErrInvalidArgumentSomeFields()
	}

	if req.TargetId == 0 {
		return nil, pb.ErrInvalidArgumentSomeFields()
	}

	subscribe := model.Subscribe{Model: gorm.Model{ID: uint(req.Id)}, UserID: authUser.ID}
	validateSubscribeResult := model.DB().First(&subscribe)
	if validateSubscribeResult.RowsAffected == 0 {
		err = errors.Wrap(validateSubscribeResult.Error, "subscribe and user ID mismatch")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	targetSubscribe := model.Subscribe{Model: gorm.Model{ID: uint(req.TargetId)}, UserID: authUser.ID}
	validateSubscribeResult = model.DB().First(&targetSubscribe)
	if validateSubscribeResult.RowsAffected == 0 {
		err = errors.Wrap(validateSubscribeResult.Error, "subscribe and user ID mismatch")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}

	errs := []error{}
	for i := range req.SelectedIds {
		entityID := req.SelectedIds[i]
		subscribeEntity := model.SubscribeEntities{
			SubscribeID: subscribe.ID,
			Subscribe:   subscribe,
			EntityID:    entityID,
			UniqueKey:   subscribeuril.GenerateSubscribeTopic(subscribe.ID, entityID),
		}
		targetSubscribeEntity := model.SubscribeEntities{
			Subscribe:   targetSubscribe,
			SubscribeID: targetSubscribe.ID,
			EntityID:    entityID,
			UniqueKey:   subscribeuril.GenerateSubscribeTopic(targetSubscribe.ID, entityID),
		}
		if err := model.DB().Debug().Where(targetSubscribeEntity).First(&targetSubscribeEntity).Error; !errors.Is(err, gorm.ErrRecordNotFound) {
			errs = append(errs, errors.New("target subscribe entity already exists"))
			continue
		}
		if err = model.DB().Debug().Model(&subscribeEntity).Where(subscribeEntity).Updates(targetSubscribeEntity).Error; err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) == len(req.SelectedIds) && len(errs) > 0 {
		err = errors.Wrap(errs[0], "change subscribed failed")
		log.Error("err:", err)
		log.Error("more err form db process:", errs)
		return nil, pb.ErrInternalError()
	}

	resp := &pb.ChangeSubscribedResponse{Status: SuccessStatus}
	if len(errs) != 0 {
		resp.Status = ErrPartialFailure
	}

	return resp, nil
}

func (s *SubscribeService) ValidateSubscribed(ctx context.Context, req *pb.ValidateSubscribedRequest) (*pb.ValidateSubscribedResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	if req.Topic == "" {
		return nil, pb.ErrInvalidArgumentSomeFields()
	}

	subscribe := model.Subscribe{Endpoint: req.Topic, UserID: authUser.ID}
	var find model.Subscribe
	validateSubscribeResult := model.DB().Debug().Model(&subscribe).Select("id").Where(&subscribe).Find(&find)
	if validateSubscribeResult.RowsAffected == 0 || validateSubscribeResult.Error != nil {
		err = errors.New("subscribe and user mismatch")
		log.Error("invalid error:", err)
		return nil, pb.ErrUnauthenticated()
	}
	resp := &pb.ValidateSubscribedResponse{Status: SuccessStatus}
	log.Info("validate subscribe success", req)
	return resp, nil
}

func (s *SubscribeService) SubscribeByDevice(ctx context.Context, req *pb.SubscribeByDeviceRequest) (*pb.SubscribeByDeviceResponse, error) {
	authUser, err := auth.GetUser(ctx)
	if nil != err {
		log.Error("err:", err)
		return nil, err
	}
	if req.Id == "" {
		return nil, pb.ErrInvalidArgumentSomeFields()
	}
	if req.SubscribeIds == nil || len(req.SubscribeIds) == 0 {
		return nil, errors.New("invalid subscribe ids")
	}

	var find []model.Subscribe
	validateSubscribeResult := model.DB().Model(&model.Subscribe{}).
		Select("id").
		Where("id IN ?", req.SubscribeIds).
		Where("user_id = ?", authUser.ID).Find(&find)
	if validateSubscribeResult.RowsAffected != int64(len(req.SubscribeIds)) {
		err = errors.Wrap(validateSubscribeResult.Error, "device and user mismatch")
		log.Error("err:", err)
		return nil, pb.ErrUnauthenticated()
	}
	subscribeEntities := make([]model.SubscribeEntities, len(req.SubscribeIds))
	for i := range req.SubscribeIds {
		subscribeEntities[i] = model.SubscribeEntities{
			SubscribeID: uint(req.SubscribeIds[i]),
			EntityID:    req.Id,
			UniqueKey:   subscribeuril.GenerateSubscribeTopic(uint(req.SubscribeIds[i]), req.Id),
		}
	}
	if err = model.DB().Debug().Create(&subscribeEntities).Error; err != nil {
		log.Error("create err:", err)
		return nil, pb.ErrInternalError()
	}

	resp := &pb.SubscribeByDeviceResponse{Status: SuccessStatus}
	return resp, nil
}

// createSubscribeEntitiesRecords create SubscribeEntities(subscribe_entities table) records.
func (s *SubscribeService) createSubscribeEntitiesRecords(entityIDs []string, subscribe *model.Subscribe) []model.SubscribeEntities {
	records := make([]model.SubscribeEntities, 0, len(entityIDs))
	for _, entityID := range entityIDs {
		subscribeEntity := model.SubscribeEntities{
			Subscribe: *subscribe,
			EntityID:  entityID,
			UniqueKey: subscribeuril.GenerateSubscribeTopic(subscribe.ID, entityID),
		}
		records = append(records, subscribeEntity)
	}
	return records
}

func (s *SubscribeService) getDeviceEntitiesIDsFromGroups(ctx context.Context, groups []string, token string) ([]string, error) {
	var data []string
	dc := deviceutil.NewClient(token)
	for i := range groups {
		bytes, err := dc.SearchDefault(deviceutil.DeviceSearch, deviceutil.Conditions{deviceutil.GroupQuery(groups[i]), deviceutil.DeviceTypeQuery()})
		if err != nil {
			log.Error("query device by device group err:", err)
			return nil, err
		}
		log.Info("query device by device group:", string(bytes))
		resp, err := deviceutil.ParseSearchResponse(bytes)
		if err != nil {
			log.Error("parse device search response err:", err)
			return nil, err
		}

		for _, device := range resp.Data.ListDeviceObject.Items {
			data = append(data, device.Id)
		}
	}
	return data, nil
}

func (s *SubscribeService) getDeviceEntitiesIDsFromTemplates(ctx context.Context, templates []string, token string) ([]string, error) {
	var data []string
	dc := deviceutil.NewClient(token)
	for i := range templates {
		bytes, err := dc.SearchDefault(deviceutil.DeviceSearch, deviceutil.Conditions{deviceutil.TemplateQuery(templates[i])})
		if err != nil {
			log.Error("query device by device group err:", err)
			return nil, err
		}
		resp, err := deviceutil.ParseSearchResponse(bytes)
		if err != nil {
			log.Error("parse device search response err:", err)
			return nil, err
		}

		for _, device := range resp.Data.ListDeviceObject.Items {
			data = append(data, device.Id)
		}
	}
	return data, nil
}

func (s SubscribeService) deviceEntities(ids []string, token string) ([]*pb.Entity, error) {
	entities := make([]*pb.Entity, 0, len(ids))
	client := deviceutil.NewClient(token)
	for _, id := range ids {
		bytes, err := client.SearchDefault(deviceutil.EntitySearch, deviceutil.Conditions{deviceutil.DeviceQuery(id)})
		if err != nil {
			log.Error("query device by device id err:", err)
			return nil, err
		}
		resp, err := deviceutil.ParseSearchEntityResponse(bytes)
		if err != nil {
			log.Error("parse device search response err:", err)
			return nil, err
		}
		if len(resp.Data.Items) == 0 {
			log.Error("device not found:", id)
			return nil, ErrDeviceNotFound
		}
		entity := &pb.Entity{
			ID:        id,
			Name:      resp.Data.Items[0].Properties.BasicInfo.Name,
			Template:  resp.Data.Items[0].Properties.BasicInfo.TemplateName,
			Group:     resp.Data.Items[0].Properties.BasicInfo.ParentName,
			Status:    "offline",
			UpdatedAt: resp.Data.Items[0].Properties.SysField.UpdatedAt,
		}
		if resp.Data.Items[0].Properties.ConnectionInfo.IsOnline {
			entity.Status = "online"
		}
		entities = append(entities, entity)
	}
	return entities, nil
}

func (s SubscribeService) getEntitiesByConditions(conditions deviceutil.Conditions, token, query string, num, size int32) ([]*pb.Entity, error) {
	client := deviceutil.NewClient(token)
	entities := make([]*pb.Entity, 0)

	bytes, err := client.Search(deviceutil.EntitySearch, conditions, query, num, size)
	if err != nil {
		log.Error("query device by device id err:", err)
		return nil, err
	}
	resp, err := deviceutil.ParseSearchEntityResponse(bytes)
	if err != nil {
		log.Error("parse device search response err:", err)
		return nil, err
	}
	if len(resp.Data.Items) == 0 {
		log.Error("device not found:", conditions)
		return nil, ErrDeviceNotFound
	}

	for _, item := range resp.Data.Items {
		entity := &pb.Entity{
			ID:        item.Id,
			Name:      item.Properties.BasicInfo.Name,
			Template:  item.Properties.BasicInfo.TemplateName,
			Group:     item.Properties.BasicInfo.ParentName,
			Status:    "offline",
			UpdatedAt: item.Properties.SysField.UpdatedAt,
		}
		if item.Properties.ConnectionInfo.IsOnline {
			entity.Status = "online"
		}
		entities = append(entities, entity)
	}
	return entities, nil
}
