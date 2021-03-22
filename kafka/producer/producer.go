package producer

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Shopify/sarama"
	"os"
	"strings"
	"time"
)

type KafkaConf struct {
	Topic           string              `json:"Topic"`
	BrokersPeers    string              `json:"BrokersPeers"`
	SaslUser        string              `json:"SaslUser"`
	SaslPasswd      string              `json:"SaslPasswd"`
	DialTimeout     string              `json:"DialTimeout"`
	KeepAlive       string              `json:"KeepAlive"`
	MaxMessageBytes int                 `json:"MaxMessageBytes"`
	Version         sarama.KafkaVersion `json:"version"`
}

type Sender struct {
	name         string
	producer     sarama.AsyncProducer
	cfg          *sarama.Config
	Topic        string
	BrokersPeers []string
	ticker       *time.Ticker
}

func NewProducer(c *KafkaConf) (kafkaSender *Sender, err error) {

	topic := c.Topic
	if len(topic) == 0 {
		err = errors.New("topic is nil")
		return
	}
	brokers := strings.Split(c.BrokersPeers, ",")
	if len(brokers) == 0 {
		err = errors.New("brokers is nil")
		return
	}
	name := fmt.Sprintf("kafkaSender:(kafkaUrl:%s,topic:%s)", brokers, topic)
	hostName, err := os.Hostname()
	if err != nil {
		hostName = "getHostnameErr:" + err.Error()
		err = nil
	}
	sasluser := c.SaslUser
	saslpasswd := c.SaslPasswd
	cfg := sarama.NewConfig()
	cfg.Version = c.Version
	cfg.Producer.Return.Errors = true
	cfg.ClientID = hostName
	cfg.Producer.Partitioner = func(topic string) sarama.Partitioner { return sarama.NewRoundRobinPartitioner(topic) }
	if len(sasluser) > 0 && len(saslpasswd) > 0 {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.User = sasluser
		cfg.Net.SASL.Password = saslpasswd
	}
	
	if len(c.DialTimeout) == 0 {
		c.DialTimeout = "30s"
	}
	cfg.Net.DialTimeout, err = time.ParseDuration(c.DialTimeout)
	if err != nil {
		err = errors.New("DialTimeout parse failed")
		return
	}
	cfg.Net.KeepAlive, err = time.ParseDuration(c.KeepAlive)
	if err != nil {
		err = errors.New("KeepAlive parse failed")
		return
	}
	cfg.Producer.MaxMessageBytes = c.MaxMessageBytes
	producer, err := sarama.NewAsyncProducer(brokers, cfg)

	if err != nil {
		return
	}
	kafkaSender = newSender(name, brokers, topic, cfg, producer)
	return
}
func newSender(name string, brokers []string, topic string, cfg *sarama.Config, producer sarama.AsyncProducer) (k *Sender) {
	k = &Sender{
		name:         name,
		producer:     producer,
		Topic:        topic,
		BrokersPeers: brokers,
		ticker:       time.NewTicker(time.Second * cfg.Net.DialTimeout),
	}
	return
}

func (this *Sender) Name() string {
	return this.name
}
func (this *Sender) ReadMessageToErrorChan() <-chan *sarama.ProducerError {
	return this.producer.Errors()
}
func (this *Sender) Send(data interface{}) error {
	var producer = this.producer
	message, err := this.getEventMessage(data)
	if err != nil {
		return err
	}

	select {
	case producer.Input() <- message:
	case <-this.ticker.C:
		return fmt.Errorf("send kafka failed:%v", this.name)
	}
	return nil
}

func (kf *Sender) getEventMessage(event interface{}) (pm *sarama.ProducerMessage, err error) {
	value, err := json.Marshal(event)
	if err != nil {
		return
	}
	pm = &sarama.ProducerMessage{
		Topic: kf.Topic,
		Value: sarama.StringEncoder(value),
	}
	return
}