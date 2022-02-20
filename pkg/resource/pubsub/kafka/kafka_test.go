package kafka

// import (
// 	"context"
// 	"testing"

// 	cloudevents "github.com/cloudevents/sdk-go"
// )

// func TestProduct(t *testing.T) {
// 	kafkaIns, err := newKafkaPubsub("test", map[string]interface{}{
// 		"topic":   "test",
// 		"group":   "core-nodes",
// 		"brokers": []string{"localhost:9092"},
// 		"timeout": 20,
// 	})
// 	if nil != err {
// 		t.Fatal(err)
// 	}

// 	ev := cloudevents.NewEvent()
// 	ev.SetID("ev-123")
// 	ev.SetType("test")
// 	ev.SetSource("Test")
// 	ev.SetData("{}")
// 	if err = kafkaIns.Send(context.Background(), ev); nil != err {
// 		t.Fatal(err)
// 	}
// }

// func TestConsume(t *testing.T) {
// 	kafkaIns, err := newKafkaPubsub("test", map[string]interface{}{
// 		"topic":   "test",
// 		"group":   "core-nodes",
// 		"brokers": []string{"localhost:9092"},
// 		"timeout": 20,
// 	})
// 	if nil != err {
// 		t.Fatal(err)
// 	}

// 	kafkaIns.Received(context.Background(), func(ctx context.Context, ev cloudevents.Event) error {
// 		t.Log("received event: ", ev)
// 		return nil
// 	})
// }

// func TestKafkaPartition(t *testing.T) {
// 	fmt.Printf("producer_test\n")
// 	config := sarama.NewConfig()
// 	config.Producer.RequiredAcks = sarama.WaitForAll
// 	config.Producer.Partitioner = sarama.NewRandomPartitioner
// 	config.Producer.Return.Successes = true
// 	config.Producer.Return.Errors = true
// 	config.Version = sarama.V0_11_0_2

// 	producer, err := sarama.NewAsyncProducer([]string{"localhost:9092"}, config)
// 	if err != nil {
// 		fmt.Printf("producer_test create producer error :%s\n", err.Error())
// 		return
// 	}

// 	defer producer.AsyncClose()

// 	// send message
// 	msg := &sarama.ProducerMessage{
// 		Topic: "test",
// 		Key:   sarama.StringEncoder("go_test"),
// 	}

// 	value := "this is message"
// 	for {
// 		fmt.Scanln(&value)
// 		msg.Value = sarama.ByteEncoder(value)
// 		fmt.Printf("input [%s]\n", value)

// 		// send to chain
// 		producer.Input() <- msg

// 		select {
// 		case suc := <-producer.Successes():
// 			fmt.Printf("offset: %d,  timestamp: %s", suc.Offset, suc.Timestamp.String())
// 		case fail := <-producer.Errors():
// 			fmt.Printf("err: %s\n", fail.Err.Error())
// 		}
// 	}
// }
