package common

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
)

type WorkerMetrics struct {
	Regular *WorkerMetricsWithLevel
	Special *WorkerMetricsWithLevel
	Supreme *WorkerMetricsWithLevel

	StopDebugInfo bool

	PublishThroughputTime   float64 //the millisecond for running all the publish jobs
	SubscribeThroughputTime float64 //the millisecond for receiving the messages
}

type WorkerMetricsWithLevel struct {
	numOfPublishers     uint32
	numOfPublishersDone uint32

	numOfSubscribers     uint32
	numOfSubscribersDone uint32

	numOfMessagesGenerated uint32

	numOfSuccessfulMessagesPublished uint32
	numOfFailedMessagesPublished     uint32

	numOfSuccessfulMessagesSubscribed uint32

	numOfPublisherOnConnect       uint32
	numOfPublisherOnConnectErrors uint32
	numOfPublisherConnectionLost  uint32
	numOfPublisherOnDisConnect    uint32

	numOfSubscriberOnConnect          uint32
	numOfSubscriberOnConnectErrors    uint32
	numOfSubscriberOnConnectSubErrors uint32
	numOfSubscriberConnectionLost     uint32
	numOfSubscriberOnDisConnect       uint32

	numOfSuccessfulSubscriberOnUnsubscribe uint32
	numOfFailedSubscriberOnUnsubscribe     uint32
}

func NewWorkerMetrics() *WorkerMetrics {
	return &WorkerMetrics{
		&WorkerMetricsWithLevel{},
		&WorkerMetricsWithLevel{},
		&WorkerMetricsWithLevel{},
		false,
		1e-6,
		1e-6,
	}
}

func (wml *WorkerMetricsWithLevel) getErrorInfo() string {
	info := ""
	if wml.numOfSubscriberOnConnectErrors > 0 {
		info += fmt.Sprintf("[numOfSubscriberOnConnectErrors: %d] ", wml.numOfSubscriberOnConnectErrors)
	}
	if wml.numOfSubscriberOnConnectSubErrors > 0 {
		info += fmt.Sprintf("[numOfSubscriberOnConnectSubErrors: %d] ", wml.numOfSubscriberOnConnectSubErrors)
	}
	if wml.numOfSubscriberConnectionLost > 0 {
		info += fmt.Sprintf("[numOfSubscriberConnectionLost: %d] ", wml.numOfSubscriberConnectionLost)
	}
	if wml.numOfPublisherOnConnectErrors > 0 {
		info += fmt.Sprintf("[numOfPublisherOnConnectErrors: %d] ", wml.numOfPublisherOnConnectErrors)
	}
	if wml.numOfPublisherConnectionLost > 0 {
		info += fmt.Sprintf("[numOfPublisherConnectionLost: %d] ", wml.numOfPublisherConnectionLost)
	}
	return info
}

func (wml *WorkerMetricsWithLevel) GetInfo(level string) string {
	info := fmt.Sprintf("Worker Metrics [%s] ... %+v ", level, wml)
	es := wml.getErrorInfo()
	if es != "" {
		info += fmt.Sprintf("\n* ... Worker Metrics [%s] found error => %s ", level, es)
	}
	return info
}
func (wm *WorkerMetrics) GetInfo() string {
	info := wm.Regular.GetInfo("Regular")
	info += wm.Regular.GetInfo("Special")
	info += wm.Regular.GetInfo("Supreme")
	return info
}

func (wm *WorkerMetrics) checkWillData(ctd *CalculateTargetData) (string, string, string, string, string, string, string) {
	numPublishers := int(wm.Regular.numOfPublishers + wm.Special.numOfPublishers + wm.Supreme.numOfPublishers)
	numSubscribers := int(wm.Regular.numOfSubscribers + wm.Special.numOfSubscribers + wm.Supreme.numOfSubscribers)
	numPubConnections := int(wm.Regular.numOfPublisherOnConnect + wm.Special.numOfPublisherOnConnect + wm.Supreme.numOfPublisherOnConnect)
	numSubConnections := int(wm.Regular.numOfSubscriberOnConnect + wm.Special.numOfSubscriberOnConnect + wm.Supreme.numOfSubscriberOnConnect)
	numPubOnDisConnect := int(wm.Regular.numOfPublisherOnDisConnect + wm.Special.numOfPublisherOnDisConnect + wm.Supreme.numOfPublisherOnDisConnect)
	numSubOnDisConnect := int(wm.Regular.numOfSubscriberOnDisConnect + wm.Special.numOfSubscriberOnDisConnect + wm.Supreme.numOfSubscriberOnDisConnect)
	numSubOnUnsubscribe := int(wm.Regular.numOfSuccessfulSubscriberOnUnsubscribe + wm.Special.numOfSuccessfulSubscriberOnUnsubscribe + wm.Supreme.numOfSuccessfulSubscriberOnUnsubscribe)
	numPubMessages := int(wm.Regular.numOfSuccessfulMessagesPublished + wm.Special.numOfSuccessfulMessagesPublished + wm.Supreme.numOfSuccessfulMessagesPublished)
	numSubMessages := int(wm.Regular.numOfSuccessfulMessagesSubscribed + wm.Special.numOfSuccessfulMessagesSubscribed + wm.Supreme.numOfSuccessfulMessagesSubscribed)

	var pub, sub, pubCon, subCon, subUnsub, pubMsg, subMsg = "", "", "", "", "", "", ""

	if ctd.Publishers != numPublishers {
		pub = fmt.Sprintf("[IMPERFECT, %d != (Target) %s]", numPublishers, FormatCommas(ctd.Publishers))
	} else {
		pub = fmt.Sprintf("[PERFECT, %d == (Target) %s]", numPublishers, FormatCommas(ctd.Publishers))
	}

	if ctd.Subscribers != numSubscribers {
		sub = fmt.Sprintf("[IMPERFECT, %d != (Target) %s]", numSubscribers, FormatCommas(ctd.Subscribers))
	} else {
		sub = fmt.Sprintf("[PERFECT, %d == (Target) %s]", numSubscribers, FormatCommas(ctd.Subscribers))
	}

	if ctd.Publishers != numPubConnections || ctd.Publishers != numPubOnDisConnect {
		pubCon = fmt.Sprintf("[IMPERFECT, %d/%d != (Target) %s]", numPubConnections, numPubOnDisConnect, FormatCommas(ctd.Publishers))
	} else {
		pubCon = fmt.Sprintf("[PERFECT, %d/%d == (Target) %s]", numPubConnections, numPubOnDisConnect, FormatCommas(ctd.Publishers))
	}

	if ctd.Subscribers != numSubConnections || ctd.Subscribers != numSubOnDisConnect {
		subCon = fmt.Sprintf("[IMPERFECT, %d/%d != (Target) %s]", numSubConnections, numSubOnDisConnect, FormatCommas(ctd.Subscribers))
	} else {
		subCon = fmt.Sprintf("[PERFECT, %d/%d == (Target) %s]", numSubConnections, numSubOnDisConnect, FormatCommas(ctd.Subscribers))
	}

	if ctd.Subscribers != numSubOnUnsubscribe {
		subUnsub = fmt.Sprintf("[IMPERFECT, %d != (Target) %s]", numSubOnUnsubscribe, FormatCommas(ctd.Subscribers))
	} else {
		subUnsub = fmt.Sprintf("[PERFECT, %d == (Target) %s]", numSubOnUnsubscribe, FormatCommas(ctd.Subscribers))
	}

	if ctd.MessagesForPublish != numPubMessages {
		pubMsg = fmt.Sprintf("[IMPERFECT, %d != (Target) %s]", numPubMessages, FormatCommas(ctd.MessagesForPublish))
	} else {
		pubMsg = fmt.Sprintf("[PERFECT, %d == (Target) %s]", numPubMessages, FormatCommas(ctd.MessagesForPublish))
	}

	if ctd.MessagesForSubscribe != numSubMessages {
		subMsg = fmt.Sprintf("[IMPERFECT, %d != (Target) %s]", numSubMessages, FormatCommas(ctd.MessagesForSubscribe))
	} else {
		subMsg = fmt.Sprintf("[PERFECT, %d == (Target) %s]", numSubMessages, FormatCommas(ctd.MessagesForSubscribe))
	}

	return pub, sub, pubCon, subCon, subUnsub, pubMsg, subMsg
}

func (wm *WorkerMetrics) GetSummaryInfo(ctd *CalculateTargetData) string {
	pub, sub, pubCon, subCon, subUnsub, pubMsg, subMsg := wm.checkWillData(ctd)

	info := fmt.Sprint("Worker Metrics Summary Information ... ")
	info += "\n********************************************************************************************************************************************************* "
	info += fmt.Sprintf("\n Publishers'  amount       ...   Regular [%d], Special [%d], Supreme [%d] ... (Qos:%d) ... %s", wm.Regular.numOfPublishers, wm.Special.numOfPublishers, wm.Supreme.numOfPublishers, ctd.PublishQos, pub)
	info += fmt.Sprintf("\n Subscribers' amount       ...   Regular [%d], Special [%d], Supreme [%d] ... (Qos:%d) ... %s", wm.Regular.numOfSubscribers, wm.Special.numOfSubscribers, wm.Supreme.numOfSubscribers, ctd.SubscribeQos, sub)
	info += fmt.Sprintf("\n Publishers'  connection   ...   Regular [%d/(E)%d,%d,%d], Special [%d/(E)%d,%d,%d], Supreme [%d/(E)%d,%d,%d]   ... %s", wm.Regular.numOfPublisherOnConnect, wm.Regular.numOfPublisherOnConnectErrors, wm.Regular.numOfPublisherConnectionLost, wm.Regular.numOfPublisherOnDisConnect, wm.Special.numOfPublisherOnConnect, wm.Special.numOfPublisherOnConnectErrors, wm.Special.numOfPublisherConnectionLost, wm.Special.numOfPublisherOnDisConnect, wm.Supreme.numOfPublisherOnConnect, wm.Supreme.numOfPublisherOnConnectErrors, wm.Supreme.numOfPublisherConnectionLost, wm.Supreme.numOfPublisherOnDisConnect, pubCon)
	info += fmt.Sprintf("\n Subscribers' connection   ...   Regular [%d/(E)%d,%d,%d,%d], Special [%d/(E)%d,%d,%d,%d], Supreme [%d/(E)%d,%d,%d,%d]   ... %s", wm.Regular.numOfSubscriberOnConnect, wm.Regular.numOfSubscriberOnConnectErrors, wm.Regular.numOfSubscriberConnectionLost, wm.Regular.numOfSubscriberOnConnectSubErrors, wm.Regular.numOfSubscriberOnDisConnect, wm.Special.numOfSubscriberOnConnect, wm.Special.numOfSubscriberOnConnectErrors, wm.Special.numOfSubscriberConnectionLost, wm.Special.numOfSubscriberOnConnectSubErrors, wm.Special.numOfSubscriberOnDisConnect, wm.Supreme.numOfSubscriberOnConnect, wm.Supreme.numOfSubscriberOnConnectErrors, wm.Supreme.numOfSubscriberConnectionLost, wm.Supreme.numOfSubscriberOnConnectSubErrors, wm.Supreme.numOfSubscriberOnDisConnect, subCon)
	info += fmt.Sprintf("\n Subscribers' unsubscribe  ...   Regular [%d/(F)%d], Special [%d/(F)%d], Supreme [%d/(F)%d]   ... %s", wm.Regular.numOfSuccessfulSubscriberOnUnsubscribe, wm.Regular.numOfFailedSubscriberOnUnsubscribe, wm.Special.numOfSuccessfulSubscriberOnUnsubscribe, wm.Special.numOfFailedSubscriberOnUnsubscribe, wm.Supreme.numOfSuccessfulSubscriberOnUnsubscribe, wm.Supreme.numOfFailedSubscriberOnUnsubscribe, subUnsub)
	info += fmt.Sprintf("\n Publishers'  messages     ...   Regular [%d,%d/(F)%d], Special [%d,%d/(F)%d], Supreme [%d,%d/(F)%d]   ... %s", wm.Regular.numOfMessagesGenerated, wm.Regular.numOfSuccessfulMessagesPublished, wm.Regular.numOfFailedMessagesPublished, wm.Special.numOfMessagesGenerated, wm.Special.numOfSuccessfulMessagesPublished, wm.Special.numOfFailedMessagesPublished, wm.Supreme.numOfMessagesGenerated, wm.Supreme.numOfSuccessfulMessagesPublished, wm.Supreme.numOfFailedMessagesPublished, pubMsg)
	info += fmt.Sprintf("\n Subscribers' messages     ...   Regular [%d], Special [%d], Supreme [%d]   ... %s", wm.Regular.numOfSuccessfulMessagesSubscribed, wm.Special.numOfSuccessfulMessagesSubscribed, wm.Supreme.numOfSuccessfulMessagesSubscribed, subMsg)
	info += "\n********************************************************************************************************************************************************* "

	tsi := wm.GetThroughputSummaryInfo()
	if tsi != "" {
		info += tsi
	}
	return info
}

func (wm *WorkerMetrics) GetThroughputSummaryInfo() string {
	var info = ""
	numPubMessages := float64(wm.GetPubMessages())
	numSubMessages := float64(wm.GetSubMessages())
	if numPubMessages > 0 || numSubMessages > 0 {
		pubThroughput := int(1e3 * numPubMessages / wm.PublishThroughputTime)
		subThroughput := int(1e3 * numSubMessages / wm.SubscribeThroughputTime)
		info += "\n Benchmark Summary Information : "
		info += fmt.Sprintf("\n Publishers'  Throughput : %s msg/sec, Time: %.2f ms", FormatCommas(pubThroughput), wm.PublishThroughputTime)
		info += fmt.Sprintf("\n Subscribers' Throughput : %s msg/sec, Time: %.2f ms", FormatCommas(subThroughput), wm.SubscribeThroughputTime)
		info += "\n********************************************************** \n "
	}
	return info
}

func (wm *WorkerMetrics) DebugInfo() {
	log.Debug("Worker metrics will be checked every 10 second. ")
	timeSticker := time.NewTicker(time.Second * 10)

	for {
		if wm.StopDebugInfo {
			log.Warn("Stopping the worker metrics debug info now ... \n ")
			return
		} else {
			select {
			case <-timeSticker.C:
				log.Debug(wm.Regular.GetInfo("Regular"))
				log.Debug(wm.Special.GetInfo("Special"))
				log.Debug(wm.Supreme.GetInfo("Supreme"))
			}
		}
	}
}

func (wm *WorkerMetrics) GetSubMessages() int {
	return int(wm.Regular.numOfSuccessfulMessagesSubscribed + wm.Special.numOfSuccessfulMessagesSubscribed + wm.Supreme.numOfSuccessfulMessagesSubscribed)
}

func (wm *WorkerMetrics) GetPubMessages() int {
	return int(wm.Regular.numOfSuccessfulMessagesPublished + wm.Special.numOfSuccessfulMessagesPublished + wm.Supreme.numOfSuccessfulMessagesPublished)
}
