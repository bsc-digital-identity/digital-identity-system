import json
import uuid
import os
import sys

import pika

message = {
    "event_id": str(uuid.uuid4()),
    "identity_id": "9ac41446-8945-4a7d-accf-47621830127e",
    "day": 15,
    "month": 7,
    "year": 1990,
    "schema": "12324235525474364465786756"
}

def resolve_lan_host() -> str:
    lan_host = os.environ.get("LAN_HOST_IP")
    if not lan_host:
        sys.exit("LAN_HOST_IP environment variable is required")
    return lan_host


credentials = pika.PlainCredentials('verifier_mock', 'verifier_mock')
connection = pika.BlockingConnection(
    pika.ConnectionParameters(
        host=resolve_lan_host(),
        port=5672,
        credentials=credentials
    )
)

channel = connection.channel()

channel.basic_publish(
    exchange='verifiers',
    routing_key='positive',
    body=json.dumps(message),
    properties=pika.BasicProperties(
        content_type='application/json',
        delivery_mode=2
    )
)

print(f"Sent to verified.positive queue: {message}")
connection.close()
