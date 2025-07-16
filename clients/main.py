import pika
import json


message = {
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
channel.basic_publish(exchange='verifiers', routing_key='verified', body=message, properties=pika.BasicProperties(content_type='application/json'))
print("Message sent")
connection.close()