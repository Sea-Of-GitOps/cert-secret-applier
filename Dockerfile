FROM docker.io/golang:1.23 AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o /app/main /app/cmd/*.go

FROM scratch

WORKDIR /app
COPY --from=build /app/main /app/main
COPY --from=build /app/config/config.yaml /app/config/config.yaml

EXPOSE 8080
CMD [ "./main" ]
