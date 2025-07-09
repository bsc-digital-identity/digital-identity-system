#!/bin/bash
set -e

echo "⏳ Waiting for RabbitMQ to start..."
until rabbitmqctl status > /dev/null 2>&1; do
    sleep 2
done

echo "✅ RabbitMQ is up. Creating user..."

rabbitmqctl add_user guest guest || true
rabbitmqctl set_user_tags guest administrator
rabbitmqctl set_permissions -p / guest ".*" ".*" ".*"

echo "🎉 User 'guest' created."
