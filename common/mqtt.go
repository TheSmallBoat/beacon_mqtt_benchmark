package common

import (
	"crypto/rand"
	"fmt"
	"sync/atomic"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
)

const maxClientIdRandomSectionLength = 6

type MqttClientUnit struct {
	MqttClient MQTT.Client
	Opts       *MQTT.ClientOptions
	Topic      string
	Qos        uint8

	WML   *WorkerMetricsWithLevel
	SInfo *SubscriberInfoOnReceivedMessage
}

func newPublisherMqttOptions(mpc *MqttPublisherConf) *MQTT.ClientOptions {
	var brokerUri = fmt.Sprintf("%s://%s:%d", mpc.Scheme, mpc.Hostname, mpc.Port)
	return initMqttOptions(brokerUri, mpc.Username, mpc.Password, mpc.Cleansession, mpc.Pingtimeout, mpc.Keepalive)
}
func newSubscriberMqttOptions(msc *MqttSubscriberConf) *MQTT.ClientOptions {
	var brokerUri = fmt.Sprintf("%s://%s:%d", msc.Scheme, msc.Hostname, msc.Port)
	return initMqttOptions(brokerUri, msc.Username, msc.Password, msc.Cleansession, msc.Pingtimeout, msc.Keepalive)
}

func initMqttOptions(brokerUri string, username string, password string, cleansession bool, pingtimeout uint8, keepalive uint16) *MQTT.ClientOptions {
	opts := MQTT.NewClientOptions()

	opts.SetAutoReconnect(true)
	opts.SetCleanSession(cleansession)
	opts.SetPingTimeout(time.Duration(pingtimeout) * time.Second)
	opts.SetConnectTimeout(time.Duration(keepalive) * time.Second)

	opts.AddBroker(brokerUri)
	if username != "" {
		opts.SetUsername(username)
	}
	if password != "" {
		opts.SetPassword(password)
	}

	return opts
}

// getRandomClientId returns randomized ClientId.
func getRandomClientId(prefix string) string {
	const alphaNum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, maxClientIdRandomSectionLength)
	_, _ = rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphaNum[b%byte(len(alphaNum))]
	}
	return prefix + "-" + string(bytes)
}

// with Connects connect to the MQTT broker with Options.
func NewPublisherMqttClientUnit(mpc *MqttPublisherConf, mtc *MqttTopicConf, wml *WorkerMetricsWithLevel, level string, topicid uint16) (*MqttClientUnit, error) {
	atomic.AddUint32(&wml.numOfPublishers, 1)
	clientId := getRandomClientId(mpc.Prefixshort)
	pubTopic := fmt.Sprintf("%s/%s/%d", level, mtc.Topicroot, topicid)

	opts := newPublisherMqttOptions(mpc)
	opts.SetClientID(clientId)
	mcu := &MqttClientUnit{Opts: opts, Topic: pubTopic, Qos: mpc.Qos, WML: wml}
	err := mcu.setMqttClientHandler(mcu.PublisherOnConnect, mcu.PublisherConnectionLost)
	if err != nil {
		atomic.AddUint32(&wml.numOfPublisherOnConnectErrors, 1)
	}
	atomic.AddUint32(&wml.numOfPublishersDone, 1)
	return mcu, err
}

func NewSubscriberMqttClientUnit(si *SubscriberInfoOnReceivedMessage, msc *MqttSubscriberConf, mtc *MqttTopicConf, wml *WorkerMetricsWithLevel, level string, topicid uint16) (*MqttClientUnit, error) {
	atomic.AddUint32(&wml.numOfSubscribers, 1)
	clientId := getRandomClientId(msc.Prefixshort)
	subTopic := fmt.Sprintf("%s/%s/%d", level, mtc.Topicroot, topicid)

	opts := newSubscriberMqttOptions(msc)
	opts.SetClientID(clientId)
	mcu := &MqttClientUnit{Opts: opts, Topic: subTopic, Qos: msc.Qos, WML: wml, SInfo: si}
	err := mcu.setMqttClientHandler(mcu.SubscriberOnConnect, mcu.SubscriberConnectionLost)
	if err != nil {
		atomic.AddUint32(&wml.numOfSubscriberOnConnectErrors, 1)
	}
	atomic.AddUint32(&wml.numOfSubscribersDone, 1)
	return mcu, err
}

func (m *MqttClientUnit) setMqttClientHandler(onConn MQTT.OnConnectHandler, conLostHandler MQTT.ConnectionLostHandler) error {
	m.Opts.SetOnConnectHandler(onConn)
	m.Opts.SetConnectionLostHandler(conLostHandler)

	client := MQTT.NewClient(m.Opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	m.MqttClient = client
	return nil
}

func (m *MqttClientUnit) PublisherOnConnect(client MQTT.Client) {
	atomic.AddUint32(&m.WML.numOfPublisherOnConnect, 1)
}

func (m *MqttClientUnit) PublisherConnectionLost(client MQTT.Client, reason error) {
	atomic.AddUint32(&m.WML.numOfPublisherConnectionLost, 1)
}

func (m *MqttClientUnit) onMessageReceived(client MQTT.Client, msg MQTT.Message) {
	atomic.AddUint32(&m.WML.numOfSuccessfulMessagesSubscribed, 1)
	atomic.AddUint32(&m.SInfo.num, 1)
	//atomic.AddUint64(&m.SInfo.byte,uint64(len(msg.Payload())))

	m.SInfo.t1 = time.Now()
	if m.SInfo.t0.IsZero() {
		m.SInfo.t0 = m.SInfo.t1
	}
}

func (m *MqttClientUnit) SubscriberOnConnect(client MQTT.Client) {
	atomic.AddUint32(&m.WML.numOfSubscriberOnConnect, 1)
	if token := client.Subscribe(m.Topic, byte(m.Qos), m.onMessageReceived); token.Wait() && token.Error() != nil {
		atomic.AddUint32(&m.WML.numOfSubscriberOnConnectSubErrors, 1)
		return
	}
}

func (m *MqttClientUnit) SubscriberConnectionLost(client MQTT.Client, reason error) {
	atomic.AddUint32(&m.WML.numOfSubscriberConnectionLost, 1)
}

func (m *MqttClientUnit) Disconnect(counter *uint32) {
	if m.MqttClient.IsConnected() {
		m.MqttClient.Disconnect(250)
		atomic.AddUint32(counter, 1)
	}
}

func (m *MqttClientUnit) SubscriberExit() {
	if token := m.MqttClient.Unsubscribe(m.Topic); token.Wait() && token.Error() != nil {
		atomic.AddUint32(&m.WML.numOfFailedSubscriberOnUnsubscribe, 1)
	}
	atomic.AddUint32(&m.WML.numOfSuccessfulSubscriberOnUnsubscribe, 1)
	m.Disconnect(&m.WML.numOfSubscriberOnDisConnect)
}
