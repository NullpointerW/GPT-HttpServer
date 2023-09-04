FROM golang:1.20.2  
WORKDIR /opt
ADD .  /opt

ENV  GOPROXY=https://goproxy.cn,direct 

RUN go build -o gpt3.5 ./gptcli.go  

EXPOSE 8080

CMD ["/opt/gpt3.5"]

