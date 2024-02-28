#!/bin/sh

openssl req -new -key cert.key -subj "/CN=mail.ru" -sha256 | openssl x509 -req -days 3650 -CA ca.crt -CAkey ca.key -set_serial "8282829" > hck.crt
