package common

import (
	"fmt"
	"sync"
	"time"

	tool "awesomeProject/beacon/mqtt-benchmark-sn/tool"
	log "github.com/Sirupsen/logrus"
)

const (
	checkSubscriberWorkerDoneWaitSeconds        = 2
	checkSubscriberWorkerDoneWaitNumber         = 150
	checkSubscriberMessagesNumberChangeInterval = 1
)

// Record each subscriber's info on receive message
type SubscriberInfoOnReceivedMessage struct {
	t0  time.Time
	t1  time.Time
	num uint32
	//byte uint64
}

type SubscriberThroughputStats struct {
	regular *tool.Stats
	special *tool.Stats
	supreme *tool.Stats

	regularSInfo *[]*SubscriberInfoOnReceivedMessage
	specialSInfo *[]*SubscriberInfoOnReceivedMessage
	supremeSInfo *[]*SubscriberInfoOnReceivedMessage
}

func NewSubscriberThroughputStats() *SubscriberThroughputStats {
	regularSInfo := make([]*SubscriberInfoOnReceivedMessage, 0)
	specialSInfo := make([]*SubscriberInfoOnReceivedMessage, 0)
	supremeSInfo := make([]*SubscriberInfoOnReceivedMessage, 0)

	return &SubscriberThroughputStats{
		regular:      &tool.Stats{},
		special:      &tool.Stats{},
		supreme:      &tool.Stats{},
		regularSInfo: &regularSInfo,
		specialSInfo: &specialSInfo,
		supremeSInfo: &supremeSInfo,
	}
}

func processSubscriberThroughputStats(sip *[]*SubscriberInfoOnReceivedMessage, ts *tool.Stats) (time.Time, time.Time) {
	var t0 time.Time
	var t1 time.Time

	for _, si := range *sip {
		if si.t1.After(t1) || t1.IsZero() {
			t1 = si.t1
		}
		if si.t0.Before(t0) || t0.IsZero() {
			t0 = si.t0
		}
		if si.num > 1 {
			tms := si.t1.Sub(si.t0).Seconds() * 1e3 / float64(si.num-1)
			ts.Update(tms)
		}
	}
	return t0, t1
}

func getBefore(t0 time.Time, t1 time.Time) time.Time {
	if t0.Before(t1) {
		return t0
	} else {
		return t1
	}
}
func getAfter(t0 time.Time, t1 time.Time) time.Time {
	if t0.After(t1) {
		return t0
	} else {
		return t1
	}
}

func (s *SubscriberThroughputStats) doneStats() float64 {
	regularT0, regularT1 := processSubscriberThroughputStats(s.regularSInfo, s.regular)
	specialT0, specialT1 := processSubscriberThroughputStats(s.specialSInfo, s.special)
	supremeT0, supremeT1 := processSubscriberThroughputStats(s.supremeSInfo, s.supreme)

	t0 := getBefore(getBefore(regularT0, specialT0), supremeT0)
	t1 := getAfter(getAfter(regularT1, specialT1), supremeT1)
	throughputTime := t1.Sub(t0).Seconds() * 1e3
	return throughputTime
}

func (s *SubscriberThroughputStats) Summary() (string, string, string) {
	var regular, special, supreme = "", "", ""
	if s.regular.Size() > 0 {
		regular = s.regular.ReportStatsInfo("[REGULAR] Statistical information about receiving time (ms) of each message")
	}
	if s.special.Size() > 0 {
		special = s.special.ReportStatsInfo("[SPECIAL] Statistical information about receiving time (ms) of each message")
	}
	if s.supreme.Size() > 0 {
		supreme = s.supreme.ReportStatsInfo("[SUPREME] Statistical information about receiving time (ms) of each message")
	}
	return regular, special, supreme
}

func (s *SubscriberThroughputStats) calThroughputWithLevel(ts *tool.Stats) string {
	fast := 1e3 / ts.Min()
	mean := 1e3 / ts.Mean()
	slow := 1e3 / ts.Max()
	throughput := fmt.Sprintf("Subscribe Throughput => Fastest : %.0f msg/sec, Mean: %.0f msg/sec, Slowest: %.0f msg/sec \n", fast, mean, slow)
	return throughput
}

func (s *SubscriberThroughputStats) LogInfoForSummary(wm *WorkerMetrics) {
	line := "~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ "
	wm.SubscribeThroughputTime = s.doneStats()
	regular, special, supreme := s.Summary()
	if regular != "" {
		log.Info(regular + s.calThroughputWithLevel(s.regular) + line)
	}
	if special != "" {
		log.Info(special + s.calThroughputWithLevel(s.special) + line)
	}
	if supreme != "" {
		log.Info(supreme + s.calThroughputWithLevel(s.special) + line)
	}
}

func StartSubscriberWorker(sts *SubscriberThroughputStats, wm *WorkerMetrics, wg *sync.WaitGroup, msc *MqttSubscriberConf, mtc *MqttTopicConf, sch *chan bool, ctd *CalculateTargetData) {
	var subs = make([]*MqttClientUnit, 0)

	newSubscriberWorkerWithLevel(sts.regularSInfo, wm.Regular, &subs, msc, mtc, "*Regular*")
	newSubscriberWorkerWithLevel(sts.specialSInfo, wm.Special, &subs, msc, mtc, "*Special*")
	newSubscriberWorkerWithLevel(sts.supremeSInfo, wm.Supreme, &subs, msc, mtc, "*Supreme*")

	waitStopSignal(wg, sch, &subs, wm, ctd)
}

func newSubscriberWorkerWithLevel(sip *[]*SubscriberInfoOnReceivedMessage, wml *WorkerMetricsWithLevel, subsp *[]*MqttClientUnit, msc *MqttSubscriberConf, mtc *MqttTopicConf, level string) {
	log.Infof("Start subscriber worker for [%s] ... ", level)
	for tid := uint16(0); tid < mtc.Numofeachlevel; tid++ {
		for idx := uint8(0); idx < msc.Subscribereachtopic; idx++ {
			startSubscriberWorker(sip, subsp, msc, mtc, wml, level, tid)
		}
	}
}

func startSubscriberWorker(sip *[]*SubscriberInfoOnReceivedMessage, subsp *[]*MqttClientUnit, msc *MqttSubscriberConf, mtc *MqttTopicConf, wml *WorkerMetricsWithLevel, level string, tid uint16) {
	go func() {
		si := &SubscriberInfoOnReceivedMessage{}
		mcu, err := NewSubscriberMqttClientUnit(si, msc, mtc, wml, level, tid)
		if err == nil {
			*subsp = append(*subsp, mcu)
			*sip = append(*sip, si)
		}
	}()
}

func waitStopSignal(wg *sync.WaitGroup, sc *chan bool, subsp *[]*MqttClientUnit, wm *WorkerMetrics, ctd *CalculateTargetData) {
	wg.Add(1)
	timeSticker := time.NewTicker(time.Second * 1)

	go func() {
		defer wg.Done()
		log.Debug("Subscriber workers begin wait stop signal and check pass status every 1 second. ")

		var pass, exit = false, false
		var flag, lsm = false, 0

		for {
			if !pass {
				select {
				case <-timeSticker.C:
					pass, exit = wm.CheckSubscriberWorkerMetricsProcess()
					if pass {
						log.Infof("All [%d] subscribers ready to go ... \n ", len(*subsp))
					}
					if exit {
						log.Warn("Subscribers ready to exit after 8 seconds due to finding errors ... \n ")
						wm.StopDebugInfo = true
						time.Sleep(8 * time.Second)
						return
					}
				}
			} else {
				select {
				case stop, ok := <-*sc:
					if ok && stop {
						log.Warn("Subscribers ready to exit due to receive a stop signal, please wait ... ")
						var wait = 0
						for {
							wait++
							time.Sleep(time.Duration(checkSubscriberWorkerDoneWaitSeconds) * time.Second)
							log.Infof("Subscribers have received [%d // %d] messages ... [%.2f%%]", wm.GetSubMessages(), ctd.MessagesForSubscribe, float32(wm.GetSubMessages())/float32(ctd.MessagesForSubscribe)*100)
							flag, lsm = wm.CheckSubscriberWorkerDone(ctd, lsm)
							if flag || wait > checkSubscriberWorkerDoneWaitNumber {
								subscriberStopTask(*subsp)
								return
							}
						}
					}
				}
			}
		}
	}()
}

func subscriberStopTask(subs []*MqttClientUnit) {
	for _, mcu := range subs {
		mcu.SubscriberExit()
	}
	log.Infof("Subscribers have unsubscribed their topics and disconnected [%d] connections. \n ", len(subs))
}

func (wm *WorkerMetrics) CheckSubscriberWorkerMetricsProcess() (bool, bool) {
	pass := false
	exit := false

	if wm.Regular.numOfSubscribersDone > 0 && wm.Special.numOfSubscribersDone > 0 && wm.Supreme.numOfSubscribersDone > 0 && wm.Regular.numOfSubscribersDone == wm.Regular.numOfSubscribers && wm.Special.numOfSubscribersDone == wm.Special.numOfSubscribers && wm.Supreme.numOfSubscribersDone == wm.Supreme.numOfSubscribers {
		if wm.Regular.numOfSubscriberOnConnect == wm.Regular.numOfSubscribers && wm.Special.numOfSubscriberOnConnect == wm.Special.numOfSubscribers && wm.Supreme.numOfSubscriberOnConnect == wm.Supreme.numOfSubscribers {
			pass = true
		}
		if wm.Regular.numOfSubscriberOnConnectErrors == wm.Regular.numOfSubscribers && wm.Special.numOfSubscriberOnConnectErrors == wm.Special.numOfSubscribers && wm.Supreme.numOfSubscriberOnConnectErrors == wm.Supreme.numOfSubscribers {
			exit = true
		}
	}
	return pass, exit
}

func (wm *WorkerMetrics) CheckSubscriberWorkerDone(ctd *CalculateTargetData, last int) (bool, int) {
	flag := false
	if wm.GetSubMessages() >= ctd.MessagesForSubscribe {
		flag = true
	} else {
		if wm.GetSubMessages() >= int(float32(ctd.MessagesForSubscribe)*0.4) && (wm.GetSubMessages()-last) < checkSubscriberMessagesNumberChangeInterval {
			flag = true
		}
	}
	return flag, wm.GetSubMessages()
}
