package service

import "testing"

func TestDecodeOrderCreatedEvent(t *testing.T) {
	var event OrderCreatedEvent
	if err := decodeOrderCreatedEvent([]byte(`{"user_id":7,"coupon_id":9}`), &event); err != nil {
		t.Fatal(err)
	}
	if event.UserID != 7 || event.CouponID != 9 {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestDecodeOrderCreatedEventRejectsInvalidPayload(t *testing.T) {
	cases := [][]byte{
		[]byte(`{`),
		[]byte(`{"user_id":0,"coupon_id":9}`),
		[]byte(`{"user_id":7}`),
	}
	for _, payload := range cases {
		var event OrderCreatedEvent
		if err := decodeOrderCreatedEvent(payload, &event); err == nil {
			t.Fatalf("expected payload %q to fail", payload)
		}
	}
}
