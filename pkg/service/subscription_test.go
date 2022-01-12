package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	pb "github.com/tkeel-io/core/api/core/v1"
)

func Test_NewSubscriptionService(t *testing.T) {
	_, err := NewSubscriptionService(context.Background(), entityManager)
	assert.Nil(t, err)
}

func Test_CreateSubscription(t *testing.T) {
	ss, err := NewSubscriptionService(context.Background(), entityManager)
	assert.Nil(t, err)

	res, err := ss.CreateSubscription(context.Background(), &pb.CreateSubscriptionRequest{
		Id:     "sub123",
		Source: "dm",
		Owner:  "admin",
		Subscription: &pb.SubscriptionObject{
			Mode:       "realtime",
			Filter:     "insert into sub123 select device123.*",
			Target:     "subscription.service",
			Topic:      "sub123-device123",
			PubsubName: "sub123",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "sub123", res.Id)
	assert.Equal(t, "admin", res.Owner)
	assert.Equal(t, "dm", res.Source)
	assert.Equal(t, "realtime", res.Subscription.Mode)
	assert.Equal(t, "insert into sub123 select device123.*", res.Subscription.Filter)
}

func Test_UpdateSubscription(t *testing.T) {
	ss, err := NewSubscriptionService(context.Background(), entityManager)
	assert.Nil(t, err)

	res, err := ss.UpdateSubscription(context.Background(), &pb.UpdateSubscriptionRequest{
		Id:     "sub123",
		Source: "dm",
		Owner:  "admin",
		Subscription: &pb.SubscriptionObject{
			Mode:       "realtime",
			Filter:     "insert into sub123 select device123.*",
			Target:     "subscription.service",
			Topic:      "sub123-device123",
			PubsubName: "sub123",
		},
	})

	assert.Nil(t, err)
	assert.Equal(t, "sub123", res.Id)
	assert.Equal(t, "admin", res.Owner)
	assert.Equal(t, "dm", res.Source)
	assert.Equal(t, "realtime", res.Subscription.Mode)
	assert.Equal(t, "insert into sub123 select device123.*", res.Subscription.Filter)
}

func Test_DeleteSubscription(t *testing.T) {
	ss, err := NewSubscriptionService(context.Background(), entityManager)
	assert.Nil(t, err)

	res, err := ss.DeleteSubscription(context.Background(), &pb.DeleteSubscriptionRequest{
		Id:     "sub123",
		Source: "dm",
		Owner:  "admin",
	})

	assert.Nil(t, err)
	assert.Equal(t, "sub123", res.Id)
	assert.Equal(t, "ok", res.Status)
}

func Test_GetSubscription(t *testing.T) {
	ss, err := NewSubscriptionService(context.Background(), entityManager)
	assert.Nil(t, err)

	res, err := ss.GetSubscription(context.Background(), &pb.GetSubscriptionRequest{
		Id:     "sub123",
		Source: "dm",
		Owner:  "admin",
	})

	assert.Nil(t, err)
	assert.Equal(t, "sub123", res.Id)
}
