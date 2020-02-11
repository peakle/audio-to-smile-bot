FROM golang:1.6

EXPOSE 80

WORKDIR /app

RUN apt-get update && apt install -y sox

#RUN go get github.com/go-redis/redis && go get github.com/joho/godotenv
#RUN go build -o Controller/controller Controller/controller.go

RUN Controller/controller &>/dev/null &
RUN Demons/CreateMessageDemon/createMessageDemon &>/dev/null &
RUN Demons/SendMailDemon/sendMailDemon &>/dev/null &
