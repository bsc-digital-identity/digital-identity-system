#!/bin/sh
set -e

rabbitmq-server &
RABBIT_PID=$!

until rabbitmqctl status > /dev/null 2>&1; do
  echo "Waiting for RabbitMQ..."
  sleep 2
done

/add_user.sh

trap "echo '💡 Caught signal, stopping...'; rabbitmqctl stop; exit 0" TERM INT

wait $RABBIT_PID
