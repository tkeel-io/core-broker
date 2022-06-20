package model

import (
	"crypto/md5"
	"encoding/hex"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/tkeel-io/core-broker/pkg/metrics"
	"github.com/tkeel-io/core-broker/pkg/util"
	"github.com/tkeel-io/kit/log"
	"gorm.io/gorm"
)

var ErrUndeleteable = errors.New("undeleteable")

type Subscribe struct {
	gorm.Model
	Title       string `gorm:"not null"`
	Description string
	UserID      string `gorm:"index"`
	TenantID    string `gorm:"index"`
	Endpoint    string `gorm:"index"`
	IsDefault   bool   `gorm:"default:false"`
}

func (s *Subscribe) BeforeCreate(tx *gorm.DB) error {
	if s.Endpoint == "" {
		s.Endpoint = util.GenerateSubscribeEndpoint()
	}
	return nil
}

func (s *Subscribe) InitMetrics() {
	out := CountSubscribeGroupByTentant(Subscribe{})
	for k, v := range out {
		metrics.CollectorSubscribeNum.WithLabelValues(k).Set(float64(v))
		metrics.CollectorSubscribeMax.WithLabelValues(k).Set(5)
	}
}

func (s *Subscribe) UpdateTenantID(userID, tenantID string) {
	DB().Model(&Subscribe{}).Where(&Subscribe{UserID: userID}).Update("tenant_id", tenantID)
}

func (s *Subscribe) UpdateEndpointTitle(oldTitle, newTitle string) error {
	subEntities := make([]*SubscribeEntities, 0)
	res := DB().Model(&SubscribeEntities{}).
		Where(&SubscribeEntities{
			SubscribeID: s.ID,
		}).Find(&subEntities)
	log.Info(res)
	if res.Error != nil {
		return nil
	}
	for _, e := range subEntities {
		log.Info(e, oldTitle, newTitle)
		subscribe := Subscribe{}
		DB().Model(&subscribe).Where("id = ?", e.SubscribeID).First(&subscribe)
		e.Subscribe = subscribe
		if err := updateEntitySubscribeEndpoint(e.EntityID,
			strings.Join([]string{
				oldTitle, strconv.FormatUint(uint64(e.SubscribeID), 10),
				AMQPAddressString(e.Subscribe.Endpoint),
			}, "@"),
			Reduce); err != nil {
			log.Error("reduce entity subscribe endpoint err")
			continue
		} else {
			if err := updateEntitySubscribeEndpoint(e.EntityID,
				strings.Join([]string{
					newTitle, strconv.FormatUint(uint64(e.SubscribeID), 10),
					AMQPAddressString(e.Subscribe.Endpoint),
				}, "@"),
				Add); err != nil {
				log.Error("add entity subscribe endpoint err")
			}
		}
	}
	return nil
}

func (s *Subscribe) BeforeDelete(tx *gorm.DB) error {
	if s.IsDefault {
		return NewUndeleteable("this is default subscribe")
	}
	destroyEndpoint(tx, s.Endpoint)
	if err := destroyRelevant(tx, s); err != nil {
		return err
	}
	return nil
}

func destroyEndpoint(tx *gorm.DB, endpoint string) {
	// TODO: destroy endpoint
	log.Debug("destroy endpoint: %s", endpoint)
}

func destroyRelevant(tx *gorm.DB, subscribe *Subscribe) error {
	log.Debugf("destroy Relevant subscribe id: %d", subscribe.ID)
	relevants := make([]SubscribeEntities, 0)
	result := tx.Session(&gorm.Session{NewDB: true}).Model(&SubscribeEntities{}).Where("subscribe_id = ?", subscribe.ID).Find(&relevants)
	if result.Error != nil {
		log.Error("Find deleted subscription relevants error:", result.Error)
		return result.Error
	}
	for _, relevant := range relevants {
		relevant.Subscribe = *subscribe
		result = tx.Session(&gorm.Session{NewDB: true}).
			Where("subscribe_id = ?", relevant.SubscribeID).
			Where("unique_key = ?", relevant.UniqueKey).
			Where("entity_id", relevant.EntityID).
			Delete(&relevant)
		if result.Error != nil {
			log.Error("delete subscription relevant error:", result.Error)
			return result.Error
		}
	}
	return nil
}

type SubscribeEntities struct {
	EntityID    string `gorm:"index;not null"`
	UniqueKey   string `gorm:"index;unique;size:255"`
	SubscribeID uint   `gorm:"index;not null"`

	Subscribe Subscribe
}

func (e *SubscribeEntities) AfterCreate(tx *gorm.DB) error {
	if e.UniqueKey == "" {
		return errors.New("UniqueKey is empty")
	}
	if e.SubscribeID == 0 {
		return errors.New("subscribeID id is empty")
	}
	if e.EntityID == "" {
		return errors.New("entityID is empty")
	}
	// tx.Model(&e.Subscribe).Where("id = ?", e.SubscribeID).First(&e.Subscribe)
	subscribe := Subscribe{}
	tx.Model(&subscribe).Where("id = ?", e.SubscribeID).First(&subscribe)
	e.Subscribe = subscribe
	log.Debug("creation of SubscribeEntities:", *e)
	if err := createCoreSubscription(e.EntityID, e.Subscribe.Endpoint, e.Subscribe.UserID); err != nil {
		err = errors.Wrap(err, "create core subscription err")
		log.Error(err)
		return err
	}
	if err := updateEntitySubscribeEndpoint(e.EntityID,
		strings.Join([]string{
			e.Subscribe.Title, strconv.FormatUint(uint64(e.SubscribeID), 10),
			AMQPAddressString(e.Subscribe.Endpoint),
		}, "@"),
		Add); err != nil {
		err = errors.Wrap(err, "update entity subscribe endpoint err")
		log.Error(err)
		return err
	}
	return nil
}

func (e *SubscribeEntities) BeforeUpdate(tx *gorm.DB) error {
	// this condition will skip by destroyRelevant() function
	if e.EntityID == "" && e.Subscribe.Endpoint == "" {
		log.Debug("skip because no releases info")
		return nil
	}
	log.Debug("deleted of SubscribeEntities:", *e)
	if err := updateEntitySubscribeEndpoint(e.EntityID,
		strings.Join([]string{
			e.Subscribe.Title, strconv.FormatUint(uint64(e.SubscribeID), 10),
			AMQPAddressString(e.Subscribe.Endpoint),
		}, "@"),
		Reduce); err != nil {
		return err
	}
	if err := deleteCoreSubscription(e.EntityID, e.Subscribe.Endpoint, e.Subscribe.UserID); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (e *SubscribeEntities) AfterUpdate(tx *gorm.DB) error {
	if e.UniqueKey == "" {
		return errors.New("UniqueKey is empty")
	}
	if e.SubscribeID == 0 {
		return errors.New("subscribeID id is empty")
	}
	if e.EntityID == "" {
		return errors.New("entityID is empty")
	}
	subscribe := Subscribe{}
	tx.Model(&subscribe).Where("id = ?", e.SubscribeID).First(&subscribe)
	e.Subscribe = subscribe
	//	tx.Model(&e.Subscribe).Where("id = ?", e.SubscribeID).First(&e.Subscribe)
	log.Debug("creation of SubscribeEntities:", *e)
	if err := createCoreSubscription(e.EntityID, e.Subscribe.Endpoint, e.Subscribe.UserID); err != nil {
		err = errors.Wrap(err, "create core subscription err")
		log.Error(err)
		return err
	}
	if err := updateEntitySubscribeEndpoint(e.EntityID,
		strings.Join([]string{
			e.Subscribe.Title, strconv.FormatUint(uint64(e.SubscribeID), 10),
			AMQPAddressString(e.Subscribe.Endpoint),
		}, "@"),
		Add); err != nil {
		err = errors.Wrap(err, "update entity subscribe endpoint err")
		log.Error(err)
		return err
	}
	return nil
}

func (s *SubscribeEntities) InitMetrics() {
	out := CountSubEntitiesGroupByTentant()
	for k, v := range out {
		metrics.CollectorSubscribeEntitiesNum.WithLabelValues(k).Set(float64(v))
		metrics.CollectorSubscribeEntitiesMax.WithLabelValues(k).Set(200)
	}
}

func (e *SubscribeEntities) BeforeDelete(tx *gorm.DB) error {
	// this condition will skip by destroyRelevant() function
	if e.EntityID == "" && e.Subscribe.Endpoint == "" {
		log.Debug("skip because no releases info")
		return nil
	}
	log.Debug("deleted of SubscribeEntities:", *e)
	if err := updateEntitySubscribeEndpoint(e.EntityID,
		strings.Join([]string{
			e.Subscribe.Title, strconv.FormatUint(uint64(e.SubscribeID), 10),
			AMQPAddressString(e.Subscribe.Endpoint),
		}, "@"),
		Reduce); err != nil {
		log.Error(err)
		//		return err
	}
	if err := deleteCoreSubscription(e.EntityID, e.Subscribe.Endpoint, e.Subscribe.UserID); err != nil {
		log.Error(err)
		//		return err
	}
	return nil
}

func createCoreSubscription(entityID string, topic, userID string) error {
	return CoreClient().Subscribe(subscriptionIDByMD5AndPrefix(entityID, topic), entityID, topic, userID)
}

func deleteCoreSubscription(entityID string, topic, userID string) error {
	return CoreClient().Unsubscribe(subscriptionIDByMD5AndPrefix(entityID, topic), userID)
}

type UtilChoice uint8

const (
	Add UtilChoice = iota + 1
	Reduce
)

func updateEntitySubscribeEndpoint(entityID, endpoint string, c UtilChoice) error {
	separator := ","
	patchData := make([]map[string]interface{}, 0)

	device, err := CoreClient().GetDeviceEntity(entityID)
	log.Debug("get device entity:", device)
	if err != nil {
		log.Error("get entity err:", err)
		return err
	}
	subscribeAddr := endpoint
	switch c {
	case Add:
		if strings.Contains(device.Properties.SysField.SubscribeAddr, endpoint) {
			return nil
		}
		if device.Properties.SysField.SubscribeAddr != "" {
			subscribeAddr = strings.Join([]string{device.Properties.SysField.SubscribeAddr, endpoint}, separator)
		}
	case Reduce:
		addrs := strings.Split(device.Properties.SysField.SubscribeAddr, separator)
		validAddresses := make([]string, 0, len(addrs))
		for i := range addrs {
			if addrs[i] != endpoint {
				log.Debugf("addrs[i]: %v, endpoint: %v", addrs[i], endpoint)
				validAddresses = append(validAddresses, addrs[i])
			}
		}
		if len(validAddresses) != 0 {
			subscribeAddr = strings.Join(validAddresses, separator)
		} else {
			subscribeAddr = ""
		}
		log.Debugf("generated subscribeAddr: %s", subscribeAddr)
	}

	patchData = append(patchData, map[string]interface{}{
		"operator": "replace",
		"path":     "sysField._subscribeAddr",
		"value":    subscribeAddr,
	})

	log.Debug("patchData:", patchData)
	log.Debug("call patch on UtilChoice (Add 1, Reduce 2):", c)

	if err = CoreClient().PatchEntity(entityID, patchData); err != nil {
		err = errors.Wrap(err, "patch entity err")
		return err
	}

	return nil
}

func NewUndeleteable(content string) error {
	return errors.Wrap(ErrUndeleteable, content)
}

const prefix = "cb-"

func subscriptionIDByMD5AndPrefix(entityID, topic string) string {
	h := md5.New()
	h.Write([]byte(entityID + topic))
	return prefix + hex.EncodeToString(h.Sum(nil))
}
