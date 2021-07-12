LDFLAGS_f2=-ldflags '-w -s $(LDFLAGS)'

all: build
build:
	CGO_ENABLED=0 go build -trimpath $(LDFLAGS_f2) -o telebot
release:
	CGO_ENABLED=0 GOOS=linux go build -trimpath $(LDFLAGS_f2) -o telebot
