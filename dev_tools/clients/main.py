import pika
import json


message = {
    "identity_id": "admin-guid-here",
    "schema_id": "12324235525474364465786756",
    "birth_day": 15,
    "birth_month": 7,
    "birth_year": 1990
}

message = json.dumps(message)

connection = pika.BlockingConnection(
    pika.ConnectionParameters(
        host='localhost',
        port=5672,
        credentials=pika.PlainCredentials('guest', 'guest')))

channel = connection.channel()
channel.basic_publish(exchange='identity', routing_key='identity.verified', body=message, properties=pika.BasicProperties(content_type='application/json'))
print("Message sent")
connection.close()