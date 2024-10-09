FROM golang:1.19

WORKDIR /work

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /reactor-test-build

EXPOSE 8080

CMD [ "/reactor-test-build" ]