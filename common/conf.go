package common

import (
	"fmt"
)

type Config struct {
	General        GeneralConf        `gcfg:"general"`
	MqttTopic      MqttTopicConf      `gcfg:"mqtt-topic"`
	MqttPublisher  MqttPublisherConf  `gcfg:"mqtt-publisher"`
	MqttSubscriber MqttSubscriberConf `gcfg:"mqtt-subscriber"`
}

type GeneralConf struct {
	Debug      bool
	Brokername string
}

type MqttTopicConf struct {
	Topicroot      string
	Numofeachlevel uint16 //the number of topics for each level of the pools.
}

type MqttPublisherConf struct {
	Scheme                     string
	Hostname                   string
	Port                       uint
	Cleansession               bool
	Qos                        uint8
	Pingtimeout                uint8
	Keepalive                  uint16
	Username                   string
	Password                   string
	Prefixname                 string
	Prefixshort                string
	Publisherseachtopic        uint8
	Messageseachpublisher      uint16
	Enablestaticmessage        bool
	Intervalmessagesgeneration uint16
}

type MqttSubscriberConf struct {
	Scheme              string
	Hostname            string
	Port                uint
	Cleansession        bool
	Qos                 uint8
	Pingtimeout         uint8
	Keepalive           uint16
	Username            string
	Password            string
	Prefixname          string
	Prefixshort         string
	Subscribereachtopic uint8
}

type CalculateTargetData struct {
	Topics               int
	Publishers           int
	PublishQos           int
	MessagesForPublish   int
	Subscribers          int
	SubscribeQos         int
	MessagesForSubscribe int
}

func (cfg *Config) CalculateData() *CalculateTargetData {
	numTopics := int(cfg.MqttTopic.Numofeachlevel * 3)
	numPublishers := int(cfg.MqttPublisher.Publisherseachtopic) * numTopics
	numMessagesForPublish := int(cfg.MqttPublisher.Messageseachpublisher) * numPublishers
	numSubscribers := int(cfg.MqttSubscriber.Subscribereachtopic) * numTopics
	numMessagesForSubscribe := int(cfg.MqttPublisher.Messageseachpublisher) * int(cfg.MqttPublisher.Publisherseachtopic) * numSubscribers

	return &CalculateTargetData{
		numTopics,
		numPublishers,
		int(cfg.MqttPublisher.Qos),
		numMessagesForPublish,
		numSubscribers,
		int(cfg.MqttSubscriber.Qos),
		numMessagesForSubscribe,
	}
}

func (cfg *Config) GetConfInfo() string {
	info := fmt.Sprintf("Configuration information ... \n[general] => %+v, [mqtt-topic] => %+v, [mqtt-publisher] => %+v, [mqtt-subscriber] => %+v \n ", cfg.General, cfg.MqttTopic, cfg.MqttPublisher, cfg.MqttSubscriber)
	return info
}

func (cfg *Config) CalculateInfo() string {
	ctd := cfg.CalculateData()
	line := "-------------------------------------------------------------------------------------------------------------------"
	info := fmt.Sprintf("Calculated data based on configuration information will be ... \n %s \n [Topics: %s], [publishers (Qos:%d): %s -> messages: %s], [subscribers (Qos:%d): %s <- messages: %s] \n %s \n ", line, FormatCommas(ctd.Topics), ctd.PublishQos, FormatCommas(ctd.Publishers), FormatCommas(ctd.MessagesForPublish), ctd.SubscribeQos, FormatCommas(ctd.Subscribers), FormatCommas(ctd.MessagesForSubscribe), line)
	return info
}
