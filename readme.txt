
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o beacon_mbt -a -ldflags '-extldflags "-static"' .
~~~~~~~~~~~~~~~~
docker build . --tag beacon/mbt:1.06
~~~~~~~~~~~~~~~~
docker run -d --name mbt beacon/mbt:1.06
docker log -f mbt
~~~~~~~~~~~~~~~~
temporary command:
 docker run -ti --rm --name mbt -v ~/go/src/awesomeProject/beacon:/opt/beacon alpine:3.11.2 /bin/sh
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

sample report
~~~~~~~~~~~~~~~~
Larrys-MacBook-Pro:mqtt-benchmark-sn abechua$ ./mqtt-benchmark-sn run -c ./conf/mqtt-benchmark.ini
INFO[2020-01-06T08:40:51+13:00] Loading configuration information from './conf/mqtt-benchmark.ini'
INFO[2020-01-06T08:40:51+13:00] Configuration information ...
[general] => {Debug:false}, [mqtt-topic] => {Topicroot:Benchmark Numofeachlevel:5}, [mqtt-publisher] => {Scheme:tcp Hostname:127.0.0.1 Port:1883 Cleansession:true Qos:0 Pingtimeout:1 Keepalive:60 Username:x Password: Prefixname:PubBenchmark Prefixshort:PB Publisherseachtopic:3 Messageseachpublisher:500}, [mqtt-subscriber] => {Scheme:tcp Hostname:127.0.0.1 Port:1883 Cleansession:true Qos:0 Pingtimeout:1 Keepalive:60 Username:x Password: Prefixname:SubBenchmark Prefixshort:SB Subscribereachtopic:2}

INFO[2020-01-06T08:40:51+13:00] Calculated data based on configuration information will be ...
 -------------------------------------------------------------------------------------------------------------------
 [Topics: 15], [publishers (Qos:0): 45 -> messages: 22,500], [subscribers (Qos:0): 30 <- messages: 45,000]
 -------------------------------------------------------------------------------------------------------------------

INFO[2020-01-06T08:40:51+13:00] Start subscriber worker for [Regular] ...
INFO[2020-01-06T08:40:51+13:00] Start subscriber worker for [Special] ...
INFO[2020-01-06T08:40:51+13:00] Start subscriber worker for [Supreme] ...
INFO[2020-01-06T08:40:52+13:00] Start publisher worker for [Regular] ...
INFO[2020-01-06T08:40:52+13:00] Start publisher worker for [Special] ...
INFO[2020-01-06T08:40:52+13:00] Start publisher worker for [Supreme] ...
INFO[2020-01-06T08:40:56+13:00] All [30] subscribers ready to go ...

INFO[2020-01-06T08:40:57+13:00] All publishers ready to go ...

INFO[2020-01-06T08:40:58+13:00] Publisher all [45] workers have finished their tasks ... 477.352 ms ...

WARN[2020-01-06T08:40:58+13:00] Subscribers ready to exit due to receive a stop signal, please wait ...
INFO[2020-01-06T08:40:59+13:00] Subscribers have received 45000 messages, expected number is 45000 ...
INFO[2020-01-06T08:40:59+13:00] Subscribers have unsubscribed their topics and disconnected [30] connections.

INFO[2020-01-06T08:40:59+13:00] [REGULAR] Statistical information about publishing time (ms) of each message ......
Size[15], Min[0.87], Mean[0.88], Max[0.91]
PopulationVariance[0.00], SampleVariance[0.00], PopulationStandardDeviation[0.01], SampleStandardDeviation[0.01]
PopulationSkew[0.67], SampleSkew[0.74], PopulationKurtosis[-0.41], SampleKurtosis[-0.06]
Publish Throughput => Fastest : 1154 msg/sec, Mean: 1133 msg/sec, Slowest: 1104 msg/sec
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
INFO[2020-01-06T08:40:59+13:00] [SPECIAL] Statistical information about publishing time (ms) of each message ......
Size[15], Min[0.87], Mean[0.88], Max[0.90]
PopulationVariance[0.00], SampleVariance[0.00], PopulationStandardDeviation[0.01], SampleStandardDeviation[0.01]
PopulationSkew[0.59], SampleSkew[0.66], PopulationKurtosis[-0.31], SampleKurtosis[0.09]
Publish Throughput => Fastest : 1146 msg/sec, Mean: 1131 msg/sec, Slowest: 1107 msg/sec
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
INFO[2020-01-06T08:40:59+13:00] [SUPREME] Statistical information about publishing time (ms) of each message ......
Size[15], Min[0.86], Mean[0.89], Max[0.91]
PopulationVariance[0.00], SampleVariance[0.00], PopulationStandardDeviation[0.01], SampleStandardDeviation[0.01]
PopulationSkew[-0.01], SampleSkew[-0.01], PopulationKurtosis[0.88], SampleKurtosis[1.80]
Publish Throughput => Fastest : 1146 msg/sec, Mean: 1131 msg/sec, Slowest: 1107 msg/sec
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
INFO[2020-01-06T08:40:59+13:00] [REGULAR] Statistical information about receiving time (ms) of each message ......
Size[10], Min[0.63], Mean[0.70], Max[0.75]
PopulationVariance[0.00], SampleVariance[0.00], PopulationStandardDeviation[0.05], SampleStandardDeviation[0.06]
PopulationSkew[-0.42], SampleSkew[-0.50], PopulationKurtosis[-1.77], SampleKurtosis[-2.17]
Subscribe Throughput => Fastest : 1597 msg/sec, Mean: 1432 msg/sec, Slowest: 1339 msg/sec
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
INFO[2020-01-06T08:40:59+13:00] [SPECIAL] Statistical information about receiving time (ms) of each message ......
Size[10], Min[0.62], Mean[0.71], Max[0.75]
PopulationVariance[0.00], SampleVariance[0.00], PopulationStandardDeviation[0.05], SampleStandardDeviation[0.05]
PopulationSkew[-1.36], SampleSkew[-1.62], PopulationKurtosis[0.07], SampleKurtosis[1.09]
Subscribe Throughput => Fastest : 1613 msg/sec, Mean: 1410 msg/sec, Slowest: 1339 msg/sec
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
INFO[2020-01-06T08:40:59+13:00] [SUPREME] Statistical information about receiving time (ms) of each message ......
Size[10], Min[0.63], Mean[0.71], Max[0.74]
PopulationVariance[0.00], SampleVariance[0.00], PopulationStandardDeviation[0.04], SampleStandardDeviation[0.04]
PopulationSkew[-1.41], SampleSkew[-1.67], PopulationKurtosis[0.11], SampleKurtosis[1.17]
Subscribe Throughput => Fastest : 1613 msg/sec, Mean: 1410 msg/sec, Slowest: 1339 msg/sec
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~
INFO[2020-01-06T08:40:59+13:00] Worker Metrics Summary Information ...
*********************************************************************************************************************************************************
 Publishers'  amount       ...   Regular [15], Special [15], Supreme [15] ... (Qos:0) ... [PERFECT, 45 == (Target) 45]
 Subscribers' amount       ...   Regular [10], Special [10], Supreme [10] ... (Qos:0) ... [PERFECT, 30 == (Target) 30]
 Publishers'  connection   ...   Regular [15/(E)0,0,15], Special [15/(E)0,0,15], Supreme [15/(E)0,0,15]   ... [PERFECT, 45/45 == (Target) 45]
 Subscribers' connection   ...   Regular [10/(E)0,0,0,10], Special [10/(E)0,0,0,10], Supreme [10/(E)0,0,0,10]   ... [PERFECT, 30/30 == (Target) 30]
 Subscribers' unsubscribe  ...   Regular [10/(F)0], Special [10/(F)0], Supreme [10/(F)0]   ... [PERFECT, 30 == (Target) 30]
 Publishers'  messages     ...   Regular [7500,7500/(F)0], Special [7500,7500/(F)0], Supreme [7500,7500/(F)0]   ... [PERFECT, 22500 == (Target) 22,500]
 Subscribers' messages     ...   Regular [15000], Special [15000], Supreme [15000]   ... [PERFECT, 45000 == (Target) 45,000]
*********************************************************************************************************************************************************
 Benchmark Summary Information :
 Publishers'  Throughput : 47,135 msg/sec, Time: 477.35 ms
 Subscribers' Throughput : 39,547 msg/sec, Time: 1137.87 ms
**********************************************************
