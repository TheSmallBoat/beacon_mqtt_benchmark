package common

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	tool "awesomeProject/beacon/mqtt-benchmark-sn/tool"
	log "github.com/Sirupsen/logrus"
)

const (
	publisherMessageChannelSize               = 8
	checkPublishWorkerProgressIntervalSeconds = 2
)

type MessageChannelWithClient struct {
	MessageChan chan string //the channel for the task of generating messages
}

type PublisherSignal struct {
	StopForSubscribeChan *chan bool //the channel for stop signal

	StartJobsFlag bool      //the flag for starting jobs
	StartJobsTime time.Time //the begin time for running jobs
	EndJobsTime   time.Time //the end time for running jobs

	NumPublisherFinished uint32 //the number of publisher who have finished it's tasks
	NumPublisher         uint32 //the number of publisher
	IsFinished           bool   //the flag of the finish status
}

// Millisecond Level
type PublisherRunTimeStats struct {
	regular *tool.Stats
	special *tool.Stats
	supreme *tool.Stats
}

func NewPublishRunTimeStats() *PublisherRunTimeStats {
	return &PublisherRunTimeStats{
		&tool.Stats{},
		&tool.Stats{},
		&tool.Stats{},
	}
}

func (p *PublisherRunTimeStats) Summary() (string, string, string) {
	var regular, special, supreme = "", "", ""
	if p.regular.Size() > 0 {
		regular = p.regular.ReportStatsInfo("[REGULAR] Statistical information about publishing time (ms) of each message")
	}
	if p.special.Size() > 0 {
		special = p.special.ReportStatsInfo("[SPECIAL] Statistical information about publishing time (ms) of each message")
	}
	if p.supreme.Size() > 0 {
		supreme = p.supreme.ReportStatsInfo("[SUPREME] Statistical information about publishing time (ms) of each message")
	}
	return regular, special, supreme
}

func (p *PublisherRunTimeStats) calThroughputWithLevel(ts *tool.Stats) string {
	fast := 1e3 / ts.Min()
	mean := 1e3 / ts.Mean()
	slow := 1e3 / ts.Max()
	throughput := fmt.Sprintf("Publish Throughput => Fastest : %.0f msg/sec, Mean: %.0f msg/sec, Slowest: %.0f msg/sec \n", fast, mean, slow)
	return throughput
}

func (p *PublisherRunTimeStats) LogInfoForSummary() {
	line := "~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~ "
	regular, special, supreme := p.Summary()
	if regular != "" {
		log.Info(regular + p.calThroughputWithLevel(p.regular) + line)
	}
	if special != "" {
		log.Info(special + p.calThroughputWithLevel(p.special) + line)
	}
	if supreme != "" {
		log.Info(supreme + p.calThroughputWithLevel(p.special) + line)
	}
}

func StartPublisherWorker(pts *PublisherRunTimeStats, wm *WorkerMetrics, wg *sync.WaitGroup, mpc *MqttPublisherConf, mtc *MqttTopicConf, sch *chan bool, ctd *CalculateTargetData) {
	ps := &PublisherSignal{
		sch,
		false,
		time.Now(),
		time.Now(),
		0,
		3 * uint32(mtc.Numofeachlevel) * uint32(mpc.Publisherseachtopic),
		false,
	}

	newPublisherWorkerWithLevel(pts.regular, wm.Regular, wg, ps, mpc, mtc, "*Regular*")
	newPublisherWorkerWithLevel(pts.special, wm.Special, wg, ps, mpc, mtc, "*Special*")
	newPublisherWorkerWithLevel(pts.supreme, wm.Supreme, wg, ps, mpc, mtc, "*Supreme*")

	checkPublishFinishStatus(wg, ps, wm, ctd)
}

func newPublisherWorkerWithLevel(ts *tool.Stats, wml *WorkerMetricsWithLevel, wg *sync.WaitGroup, ps *PublisherSignal, mpc *MqttPublisherConf, mtc *MqttTopicConf, level string) {
	log.Infof("Start publisher worker for [%s] ... ", level)

	for tid := uint16(0); tid < mtc.Numofeachlevel; tid++ {
		for idx := uint8(0); idx < mpc.Publisherseachtopic; idx++ {
			startPublisherWorker(ts, wg, ps, mpc, mtc, wml, level, tid)
		}
	}
}

func startPublisherWorker(ts *tool.Stats, wg *sync.WaitGroup, ps *PublisherSignal, mpc *MqttPublisherConf, mtc *MqttTopicConf, wml *WorkerMetricsWithLevel, level string, tid uint16) {
	go func() {
		mcwc := &MessageChannelWithClient{make(chan string, publisherMessageChannelSize)}

		mcu, err := NewPublisherMqttClientUnit(mpc, mtc, wml, level, tid)
		if err == nil {
			generateMessages(mcu, wg, mpc, mcwc, ps)
			publisherRunningTask(ts, mcu, wg, mcwc, ps)
		}
	}()
}

func checkPublishFinishStatus(wg *sync.WaitGroup, ps *PublisherSignal, wm *WorkerMetrics, ctd *CalculateTargetData) {
	wg.Add(1)
	timeSticker := time.NewTicker(time.Second * time.Duration(checkPublishWorkerProgressIntervalSeconds))

	go func() {
		defer wg.Done()
		log.Debugf("Publisher workers begin wait finish status signal and check pass status every %d second. ", checkPublishWorkerProgressIntervalSeconds)

		var pass, exit = false, false

		for {
			select {
			case <-timeSticker.C:
				if !pass {
					pass, exit = checkPublisherWorkerMetricsProcess(wm)
					if pass {
						log.Info("All publishers ready to go ... \n ")
						ps.StartJobsFlag = true
						ps.StartJobsTime = time.Now()
					}
					if exit {
						log.Warn("Publishers ready to exit after 8 seconds due to finding errors ... \n ")
						wm.StopDebugInfo = true
						time.Sleep(8 * time.Second)
						return
					}
				} else {
					wm.GetPubMessages()
					log.Infof("[%d, %d, %d] messages have published ... [%.2f%%]", wm.Regular.numOfSuccessfulMessagesPublished, wm.Special.numOfSuccessfulMessagesPublished, wm.Supreme.numOfSuccessfulMessagesPublished, float32(wm.GetPubMessages())/float32(ctd.MessagesForPublish)*100)
					log.Infof("[%d, %d, %d] messages have received ... [%.2f%%]", wm.Regular.numOfSuccessfulMessagesSubscribed, wm.Special.numOfSuccessfulMessagesSubscribed, wm.Supreme.numOfSuccessfulMessagesSubscribed, float32(wm.GetSubMessages())/float32(ctd.MessagesForSubscribe)*100)
					log.Infof("[%d // %d] publisher workers have finished their tasks ... ", ps.NumPublisherFinished, ps.NumPublisher)
					if ps.IsFinished {
						wm.PublishThroughputTime = float64(ps.EndJobsTime.Sub(ps.StartJobsTime).Nanoseconds()) / 1e6
						log.Infof("Publisher all [%d] workers have finished their tasks ... %.3f ms ... \n ", ps.NumPublisherFinished, wm.PublishThroughputTime)
						wm.StopDebugInfo = true
						return
					}
				}
			}
		}
	}()
}

func generateMessages(mcu *MqttClientUnit, wg *sync.WaitGroup, mpc *MqttPublisherConf, mcwc *MessageChannelWithClient, ps *PublisherSignal) {
	wg.Add(1)

	go func() {
		defer wg.Done()

		for {
			if ps.StartJobsFlag {
				for mid := 0; mid < int(mpc.Messageseachpublisher); mid++ {
					if mpc.Enablestaticmessage {
						mcwc.MessageChan <- "{Beacon_MBT}"
					} else {
						mcwc.MessageChan <- fmt.Sprintf("{\"time\":%d,\"mid\":%d,\"cid\":%s}", time.Now().Unix(), mid, mcu.Opts.ClientID)
					}
					atomic.AddUint32(&mcu.WML.numOfMessagesGenerated, 1)
					time.Sleep(time.Duration(mpc.Intervalmessagesgeneration) * time.Microsecond)
				}
				mcwc.MessageChan <- "{DONE}"
				return
			} else {
				time.Sleep(1 * time.Second)
			}
		}

	}()
}

func publisherRunningTask(ts *tool.Stats, mcu *MqttClientUnit, wg *sync.WaitGroup, mcwc *MessageChannelWithClient, ps *PublisherSignal) {
	wg.Add(1)

	go func() {
		defer wg.Done()
		t0 := time.Now()
		num := 0

		for {
			if ps.StartJobsFlag {
				select {
				case msg, ok := <-mcwc.MessageChan:
					if ok {
						if msg != "{DONE}" {
							mcu.publishInfo(msg)
							num++
						} else {
							atomic.AddUint32(&ps.NumPublisherFinished, 1)
							if ps.NumPublisherFinished >= ps.NumPublisher {
								ps.IsFinished = true
								ps.EndJobsTime = time.Now()
								*ps.StopForSubscribeChan <- true
							}
							tpm := float64(time.Since(t0).Nanoseconds()) / float64(num) / 1e6 // Statistics information about publishing time of each message
							ts.Update(tpm)
							mcu.Disconnect(&mcu.WML.numOfPublisherOnDisConnect)
							return
						}
					}
				}
			} else {
				time.Sleep(1 * time.Second)
				t0 = time.Now()
			}
		}
	}()
}

func (m *MqttClientUnit) publishInfo(payload string) {
	topic := m.Topic
	qos := m.Qos
	if token := m.MqttClient.Publish(topic, qos, false, payload); token.Wait() && token.Error() != nil {
		atomic.AddUint32(&m.WML.numOfFailedMessagesPublished, 1)
	}
	atomic.AddUint32(&m.WML.numOfSuccessfulMessagesPublished, 1)
}

func checkPublisherWorkerMetricsProcess(wm *WorkerMetrics) (bool, bool) {
	pass := false
	exit := false

	if wm.Regular.numOfPublishersDone > 0 && wm.Special.numOfPublishersDone > 0 && wm.Supreme.numOfPublishersDone > 0 && wm.Regular.numOfPublishersDone == wm.Regular.numOfPublishers && wm.Special.numOfPublishersDone == wm.Special.numOfPublishers && wm.Supreme.numOfPublishersDone == wm.Supreme.numOfPublishers {
		if wm.Regular.numOfPublisherOnConnect == wm.Regular.numOfPublishers && wm.Special.numOfPublisherOnConnect == wm.Special.numOfPublishers && wm.Supreme.numOfPublisherOnConnect == wm.Supreme.numOfPublishers {
			pass = true
		}
		if wm.Regular.numOfPublisherOnConnectErrors == wm.Regular.numOfPublishers && wm.Special.numOfPublisherOnConnectErrors == wm.Special.numOfPublishers && wm.Supreme.numOfPublisherOnConnectErrors == wm.Supreme.numOfPublishers {
			exit = true
		}
	}
	return pass, exit
}
