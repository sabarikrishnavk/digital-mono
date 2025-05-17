FROM devopsfaith/krakend:2.6.1

COPY krakend.json /etc/krakend/krakend.json

EXPOSE 8080

CMD [ "krakend", "run", "-d", "-c", "/etc/krakend/krakend.json" ]