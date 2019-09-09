#!/bin/sh
cd /opt/kafka

bin/kafka-server-start.sh config/server-1.properties &>/dev/null &
bin/kafka-server-start.sh config/server-2.properties &>/dev/null &
bin/kafka-server-start.sh config/server-3.properties &>/dev/null &

bin/kafka-console-producer.sh --broker-list localhost:9093,localhost:9094,localhost:9095 --topic foo
