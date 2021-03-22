package consumer

import (
	"context"
	"fmt"
	"github.com/Shopify/sarama"
	"io"
	"log"
	"sync"
)

type KafkaGroupReader struct {
	Consume  sarama.ConsumerGroup
	Consumer ConsumerHandler
	topics   []string
	cancel   context.CancelFunc
}

// Consumer represents a Sarama consumer group consumer
type ConsumerHandler struct {
	lock            *sync.RWMutex
	topics          []string
	ReadBufferChan  chan *Event
	ready           chan bool
	highWaterOffset map[int32]int64
	currentOffset   map[int32]int64
}
type Event struct {
	Session sarama.ConsumerGroupSession
	Msg     *sarama.ConsumerMessage
}

func ConfigLogger(out io.Writer) {
	sarama.Logger = log.New(out, "[Sarama] ", 0)
}
func NewKafkaReader(brokers []string, groupId string, topics []string, config *sarama.Config) (krq *KafkaGroupReader, err error) {
	/*************************************************************/
	if err = config.Validate(); err != nil {
		return nil, err
	}
	var kr = &KafkaGroupReader{topics: topics}
	client, err := sarama.NewConsumerGroup(brokers, groupId, config)
	if err != nil {
		return nil, fmt.Errorf("Error creating consumer group client: %v", err)
	}
	kr.Consume = client
	kr.Consumer = ConsumerHandler{
		ready:           make(chan bool),
		ReadBufferChan:  make(chan *Event, 2048),
		highWaterOffset: make(map[int32]int64),
		currentOffset:   make(map[int32]int64),
		lock:            new(sync.RWMutex),
	}
	go kr.consume()
	return kr, nil
}

func (r *KafkaGroupReader) consume() {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel
	go func() {
		for {
			err := r.Consume.Consume(ctx, r.topics, &r.Consumer)
			if err != nil {
				log.Printf("err from consumer: %v", err)
			}
			if ctx.Err() != nil {
				return
			}
			r.Consumer.ready = make(chan bool)
		}
	}()

	<-r.Consumer.ready
}

func (r *KafkaGroupReader) ReadLine() *Event {
	return <-r.Consumer.ReadBufferChan
}

func (r *KafkaGroupReader) Close() {
	r.cancel()
	r.Consume.Close()

}

func (r *KafkaGroupReader) Lag() map[string]int64 {
	var lag = make(map[string]int64)
	var total int64
	r.Consumer.lock.RLock()
	for pt, hi := range r.Consumer.highWaterOffset {
		cur := hi - 1 - r.Consumer.currentOffset[pt]
		total += cur
		lag[fmt.Sprintf("%v", pt)] = cur
	}
	r.Consumer.lock.RUnlock()
	lag["total"] = total
	return lag
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {

	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/master/consumer_group.go#L27-L29
	//claim.HighWaterMarkOffset()
	for message := range claim.Messages() {
		consumer.ReadBufferChan <- &Event{
			Session: session,
			Msg:     message,
		}
		consumer.lock.Lock()
		consumer.currentOffset[message.Partition] = message.Offset
		consumer.highWaterOffset[claim.Partition()] = claim.HighWaterMarkOffset()
		consumer.lock.Unlock()
	}

	return nil
}