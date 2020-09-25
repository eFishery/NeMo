FROM golang:1.15-alpine as builder

WORKDIR /app
COPY ./ .

RUN go build -o NeMo

FROM golang:1.15-alpine

WORKDIR /app

RUN mkdir -p /app/.build/sessions/
RUN mkdir -p /app/coral/

COPY --from=builder /app/NeMo ./NeMo

COPY ./coral/ /app/coral/

RUN touch /app/.build/commands.json
RUN touch /app/.build/schedules.json
RUN touch /app/.build/greetings.json

ENTRYPOINT ["./NeMo"]
