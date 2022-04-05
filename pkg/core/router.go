package core

import "fmt"

func PatchEntityURL(entityID string) string {
	return fmt.Sprintf("v1/entities/%s/patch?owner=admin&source=dm", entityID)
}

func QueryDeviceEntityURL(entityID string) string {
	return fmt.Sprintf("v1/entities/%s/properties?type=DEVICE&owner=admin&source=dm", entityID)
}

func CreateEntityURL(entityID, userID, source string) string {
	return fmt.Sprintf("v1/entities?id=%s&owner=%s&source=%s", entityID, userID, source)
}

func CreateSubscriptionURL(subID, userID, source, typeOf string) string {
	return fmt.Sprintf("v1/subscriptions?id=%s&owner=%s&source=%s&type=%s", subID, userID, source, typeOf)
}

func CreateUnsubscriptionURL(subID, userID, source, typeOf string) string {
	return fmt.Sprintf("v1/subscriptions/%s?owner=%s&source=%s&type=%s", subID, userID, source, typeOf)
}
