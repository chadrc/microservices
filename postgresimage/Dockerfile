# Image to initialize users and tables for micro service app
FROM postgres

ENV POSTGRES_USER=postgres
ENV POSTGRES_PASSWORD=postgres
ENV POSTGRES_DB=micro_services

COPY . /docker-entrypoint-initdb.d