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

func TestSameQueuedEventReusesMessageID(t *testing.T) {
	var ids []string
	controller := &QueuedEventController{
		ctx: context.Background(),
		publishOverride: func(_, msgID string, _ interface{}) error {
			ids = append(ids, msgID)
			return nil
		},
	}
	event := &QueueEvent{Subject: "os.device.inventory.alice", MsgID: "event-1:alice", Data: map[string]string{"topic": "device.inventory.changed"}}
	if err := controller.publish(event.Subject, event.MsgID, event.Data); err != nil {
		t.Fatal(err)
	}
	if err := controller.publish(event.Subject, event.MsgID, event.Data); err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 || ids[0] != event.MsgID || ids[1] != event.MsgID {
		t.Fatalf("publish IDs = %#v", ids)
	}
}
