import pika

connection = pika.BlockingConnection(
    pika.ConnectionParameters(
        host='localhost',
        port=5672,
        credentials=pika.PlainCredentials('guest', 'guest')))

channel = connection.channel()
channel.basic_publish(exchange='verifiers', routing_key='verified', body='Hello World!')
print("Message sent")
connection.close()