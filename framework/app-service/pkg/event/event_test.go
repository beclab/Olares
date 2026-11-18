package event

import (
	"context"
	"reflect"
	"testing"
)

func TestQueuedEventPassesMessageIDToPublisher(t *testing.T) {
	var subject, msgID string
	var data interface{}
	controller := &QueuedEventController{
		ctx: context.Background(),
		publishOverride: func(gotSubject, gotMsgID string, gotData interface{}) error {
			subject, msgID, data = gotSubject, gotMsgID, gotData
			return nil
		},
	}
	wantData := map[string]string{"topic": "device.inventory.changed"}
	controller.process(&QueueEvent{Subject: "os.device.inventory", MsgID: "event-1", Data: wantData})
	if subject != "os.device.inventory" || msgID != "event-1" {
		t.Fatalf("publish metadata = %q/%q", subject, msgID)
	}
	if !reflect.DeepEqual(data, wantData) {
		t.Fatalf("publish data = %#v", data)
	}
}
