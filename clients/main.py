import pika

connection = pika.BlockingConnection(pika.ConnectionParameters(host='localhost', port=5672))
channel = connection.channel()
channel.queue_declare(queue='test')
channel.basic_publish(exchange='', routing_key='test', body='Hello World!')
print("Message sent")
connection.close()