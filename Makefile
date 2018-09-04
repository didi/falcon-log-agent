TARGET = falcon-log-agent
PACK = ./falcon-log-agent cfg/ control

all: build pack

build:
	go build -ldflags "-X main.GitCommit=`git rev-parse --short HEAD`" -o $(TARGET)

pack:
	tar -zcvf falcon-log-agent.tar.gz $(PACK)
