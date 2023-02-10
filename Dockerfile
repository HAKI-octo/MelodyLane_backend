FROM golang:1.20.0 as build

WORKDIR /src

COPY . /src

RUN CGO_ENABLED=0 GOOS=linux go build -mod=readonly -v -o melodylane

FROM alpine:3

EXPOSE 8080

COPY --from=build /src/melodylane .

CMD ["/melodylane"]




