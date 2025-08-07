import pika
import json

message = {
    "identity_id": "admin-guid-here",
    "day": 15,
    "month": 7,
    "year": 1990,
    "schema": "12324235525474364465786756"
}

credentials = pika.PlainCredentials('verifier_mock', 'verifier_mock')
connection = pika.BlockingConnection(
    pika.ConnectionParameters(
        host='localhost',
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