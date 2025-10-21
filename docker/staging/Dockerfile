from golang:1.24

RUN mkdir /app

WORKDIR /app

COPY ./suaobra-app /app/

CMD ["/app/suaobra-app", "serve", "--http", "0.0.0.0:8090"]