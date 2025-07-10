from golang:1.20

RUN mkdir /app

WORKDIR /app

COPY ./suaobra-app /app/

CMD ["/app/suaobra-app", "serve", "--http", "0.0.0.0:8090"]