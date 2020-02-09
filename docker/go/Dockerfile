FROM golang:1.6

EXPOSE 80

WORKDIR /app

RUN apt-get update && apt install -y sox

#RUN go get github.com/go-redis/redis && go get github.com/joho/godotenv
#RUN go build -o Controller/controller Controller/controller.go

RUN Controller/controller &>/dev/null &
RUN Demons/CreateMessageDemon/createMessageDemon_linux &>/dev/null &
RUN Demons/sendMailDemon/sendMailDemon_linux &>/dev/null &
