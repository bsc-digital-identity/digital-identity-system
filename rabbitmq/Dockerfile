FROM rabbitmq:4.1.2-management

COPY rabbitmq.conf /etc/rabbitmq/rabbitmq.conf

COPY add_user.sh /add_user.sh

COPY run_rabbitmq.sh /run_rabbitmq.sh

RUN chmod +x /add_user.sh /run_rabbitmq.sh

# AMPQ and managment ports
EXPOSE 5672 15672

CMD ["/run_rabbitmq.sh"]