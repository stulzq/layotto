/*
 * Copyright 2021 Layotto Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package default_api

import (
	"context"
	"errors"
	"fmt"
	contrib_contenttype "github.com/dapr/components-contrib/contenttype"
	"github.com/dapr/components-contrib/pubsub"
	contrib_pubsub "github.com/dapr/components-contrib/pubsub"
	"github.com/google/uuid"
	jsoniter "github.com/json-iterator/go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	l8_comp_pubsub "mosn.io/layotto/components/pubsub"

	"encoding/base64"
	"github.com/dapr/components-contrib/contenttype"
	"mosn.io/layotto/pkg/messages"
	runtimev1pb "mosn.io/layotto/spec/proto/runtime/v1"
	"mosn.io/pkg/log"
)

func (a *api) PublishEvent(ctx context.Context, in *runtimev1pb.PublishEventRequest) (*emptypb.Empty, error) {
	result, err := a.doPublishEvent(ctx, in.PubsubName, in.Topic, in.Data, in.DataContentType, in.Metadata)
	if err != nil {
		log.DefaultLogger.Errorf("[runtime] [grpc.PublishEvent] %v", err)
	}
	return result, err
}

// doPublishEvent is a protocal irrelevant function to do event publishing.
// It's easy to add APIs for other protocals.Just move this func to a separate layer if you need.
func (a *api) doPublishEvent(ctx context.Context, pubsubName string, topic string, data []byte, contentType string, metadata map[string]string) (*emptypb.Empty, error) {
	// 1. validate
	if pubsubName == "" {
		err := status.Error(codes.InvalidArgument, messages.ErrPubsubEmpty)
		return &emptypb.Empty{}, err
	}
	if topic == "" {
		err := status.Errorf(codes.InvalidArgument, messages.ErrTopicEmpty, pubsubName)
		return &emptypb.Empty{}, err
	}
	// 2. get component
	component, ok := a.pubSubs[pubsubName]
	if !ok {
		err := status.Errorf(codes.InvalidArgument, messages.ErrPubsubNotFound, pubsubName)
		return &emptypb.Empty{}, err
	}

	// 3. new cloudevent request
	if data == nil {
		data = []byte{}
	}
	var envelope map[string]interface{}
	var err error = nil
	if contrib_contenttype.IsCloudEventContentType(contentType) {
		envelope, err = contrib_pubsub.FromCloudEvent(data, topic, pubsubName, "")
		if err != nil {
			err = status.Errorf(codes.InvalidArgument, messages.ErrPubsubCloudEventCreation, err.Error())
			return &emptypb.Empty{}, err
		}
	} else {
		envelope = contrib_pubsub.NewCloudEventsEnvelope(uuid.New().String(), l8_comp_pubsub.DefaultCloudEventSource, l8_comp_pubsub.DefaultCloudEventType, "", topic, pubsubName,
			contentType, data, "")
	}
	features := component.Features()
	pubsub.ApplyMetadata(envelope, features, metadata)

	b, err := jsoniter.ConfigFastest.Marshal(envelope)
	if err != nil {
		err = status.Errorf(codes.InvalidArgument, messages.ErrPubsubCloudEventsSer, topic, pubsubName, err.Error())
		return &emptypb.Empty{}, err
	}
	// 4. publish
	req := pubsub.PublishRequest{
		PubsubName: pubsubName,
		Topic:      topic,
		Data:       b,
		Metadata:   metadata,
	}

	// TODO limit topic scope
	err = component.Publish(&req)
	if err != nil {
		nerr := status.Errorf(codes.Internal, messages.ErrPubsubPublishMessage, topic, pubsubName, err.Error())
		return &emptypb.Empty{}, nerr
	}
	return &emptypb.Empty{}, nil
}

func (a *api) startSubscribing() error {
	// 1. check if there is no need to do it
	if len(a.pubSubs) == 0 {
		return nil
	}
	// 2. list topics
	topicRoutes, err := a.getInterestedTopics()
	if err != nil {
		return err
	}
	// return if no need to dosubscription
	if len(topicRoutes) == 0 {
		return nil
	}
	// 3. loop subscribe
	for name, pubsub := range a.pubSubs {
		if err := a.beginPubSub(name, pubsub, topicRoutes); err != nil {
			return err
		}
	}
	return nil
}

func (a *api) beginPubSub(pubsubName string, ps pubsub.PubSub, topicRoutes map[string]TopicSubscriptions) error {
	// 1. call app to find topic topic2Details.
	v, ok := topicRoutes[pubsubName]
	if !ok {
		return nil
	}
	// 2. loop subscribing every <topic, route>
	for topic, route := range v.topic2Details {
		// TODO limit topic scope
		log.DefaultLogger.Debugf("[runtime][beginPubSub]subscribing to topic=%s on pubsub=%s", topic, pubsubName)
		// ask component to subscribe
		if err := ps.Subscribe(pubsub.SubscribeRequest{
			Topic:    topic,
			Metadata: route.metadata,
		}, func(ctx context.Context, msg *pubsub.NewMessage) error {
			if msg.Metadata == nil {
				msg.Metadata = make(map[string]string, 1)
			}
			msg.Metadata[Metadata_key_pubsubName] = pubsubName
			return a.publishMessageGRPC(ctx, msg)
		}); err != nil {
			log.DefaultLogger.Warnf("[runtime][beginPubSub]failed to subscribe to topic %s: %s", topic, err)
			return err
		}
	}

	return nil
}

type Details struct {
	metadata map[string]string
}

type TopicSubscriptions struct {
	topic2Details map[string]Details
}

func (a *api) getInterestedTopics() (map[string]TopicSubscriptions, error) {
	// 1. check
	if a.topicPerComponent != nil {
		return a.topicPerComponent, nil
	}
	if a.AppCallbackConn == nil {
		return make(map[string]TopicSubscriptions), nil
	}
	comp2Topic := make(map[string]TopicSubscriptions)
	var subscriptions []*runtimev1pb.TopicSubscription

	// 2. handle app subscriptions
	client := runtimev1pb.NewAppCallbackClient(a.AppCallbackConn)
	subscriptions = listTopicSubscriptions(client, log.DefaultLogger)
	// TODO handle declarative subscriptions

	// 3. prepare result
	for _, s := range subscriptions {
		if s == nil {
			continue
		}
		if _, ok := comp2Topic[s.PubsubName]; !ok {
			comp2Topic[s.PubsubName] = TopicSubscriptions{topic2Details: make(map[string]Details)}
		}
		comp2Topic[s.PubsubName].topic2Details[s.Topic] = Details{metadata: s.Metadata}
	}

	// 4. log
	if len(comp2Topic) > 0 {
		for pubsubName, v := range comp2Topic {
			topics := []string{}
			for topic := range v.topic2Details {
				topics = append(topics, topic)
			}
			log.DefaultLogger.Infof("[runtime][getInterestedTopics]app is subscribed to the following topics: %v through pubsub=%s", topics, pubsubName)
		}
	}
	// 5. cache the result
	a.topicPerComponent = comp2Topic
	return comp2Topic, nil
}

func (a *api) publishMessageGRPC(ctx context.Context, msg *pubsub.NewMessage) error {
	// 1. Unmarshal to cloudEvent model
	var cloudEvent map[string]interface{}
	err := a.json.Unmarshal(msg.Data, &cloudEvent)
	if err != nil {
		log.DefaultLogger.Debugf("[runtime]error deserializing cloud events proto: %s", err)
		return err
	}

	// 2. Drop msg if the current cloud event has expired
	if pubsub.HasExpired(cloudEvent) {
		log.DefaultLogger.Warnf("[runtime]dropping expired pub/sub event %v as of %v", cloudEvent[pubsub.IDField].(string), cloudEvent[pubsub.ExpirationField].(string))
		return nil
	}

	// 3. Convert to proto domain struct
	envelope := &runtimev1pb.TopicEventRequest{
		Id:              cloudEvent[pubsub.IDField].(string),
		Source:          cloudEvent[pubsub.SourceField].(string),
		DataContentType: cloudEvent[pubsub.DataContentTypeField].(string),
		Type:            cloudEvent[pubsub.TypeField].(string),
		SpecVersion:     cloudEvent[pubsub.SpecVersionField].(string),
		Topic:           msg.Topic,
		PubsubName:      msg.Metadata[Metadata_key_pubsubName],
	}

	// set data field
	if data, ok := cloudEvent[pubsub.DataBase64Field]; ok && data != nil {
		decoded, decodeErr := base64.StdEncoding.DecodeString(data.(string))
		if decodeErr != nil {
			log.DefaultLogger.Debugf("unable to base64 decode cloudEvent field data_base64: %s", decodeErr)
			return err
		}

		envelope.Data = decoded
	} else if data, ok := cloudEvent[pubsub.DataField]; ok && data != nil {
		envelope.Data = nil

		if contenttype.IsStringContentType(envelope.DataContentType) {
			envelope.Data = []byte(data.(string))
		} else if contenttype.IsJSONContentType(envelope.DataContentType) {
			envelope.Data, _ = a.json.Marshal(data)
		}
	}
	// TODO tracing

	// 4. Call appcallback
	clientV1 := runtimev1pb.NewAppCallbackClient(a.AppCallbackConn)
	res, err := clientV1.OnTopicEvent(ctx, envelope)

	// 5. Check result
	return retryStrategy(err, res, cloudEvent)
}

// retryStrategy returns error when the message should be redelivered
func retryStrategy(err error, res *runtimev1pb.TopicEventResponse, cloudEvent map[string]interface{}) error {
	if err != nil {
		errStatus, hasErrStatus := status.FromError(err)
		if hasErrStatus && (errStatus.Code() == codes.Unimplemented) {
			// DROP
			log.DefaultLogger.Warnf("[runtime]non-retriable error returned from app while processing pub/sub event %v: %s", cloudEvent[pubsub.IDField].(string), err)
			return nil
		}

		err = errors.New(fmt.Sprintf("error returned from app while processing pub/sub event %v: %s", cloudEvent[pubsub.IDField].(string), err))
		log.DefaultLogger.Debugf("%s", err)
		// on error from application, return error for redelivery of event
		return err
	}

	switch res.GetStatus() {
	case runtimev1pb.TopicEventResponse_SUCCESS:
		// on uninitialized status, this is the case it defaults to as an uninitialized status defaults to 0 which is
		// success from protobuf definition
		return nil
	case runtimev1pb.TopicEventResponse_RETRY:
		return errors.New(fmt.Sprintf("RETRY status returned from app while processing pub/sub event %v", cloudEvent[pubsub.IDField].(string)))
	case runtimev1pb.TopicEventResponse_DROP:
		log.DefaultLogger.Warnf("[runtime]DROP status returned from app while processing pub/sub event %v", cloudEvent[pubsub.IDField].(string))
		return nil
	}
	// Consider unknown status field as error and retry
	return errors.New(fmt.Sprintf("unknown status returned from app while processing pub/sub event %v: %v", cloudEvent[pubsub.IDField].(string), res.GetStatus()))
}

func listTopicSubscriptions(client runtimev1pb.AppCallbackClient, log log.ErrorLogger) []*runtimev1pb.TopicSubscription {
	resp, err := client.ListTopicSubscriptions(context.Background(), &emptypb.Empty{})
	if err != nil {
		log.Errorf("[runtime][listTopicSubscriptions]error after callback: %s", err)
		return make([]*runtimev1pb.TopicSubscription, 0)
	}
	if resp != nil && len(resp.Subscriptions) > 0 {
		return resp.Subscriptions
	}
	return make([]*runtimev1pb.TopicSubscription, 0)
}
