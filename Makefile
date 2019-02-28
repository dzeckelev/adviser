DOCKERCMD=docker-compose

all: demon
build:
	$(DOCKERCMD) build --force-rm
run: build
	$(DOCKERCMD) up
demon: build
	$(DOCKERCMD) up -d
stop:
	$(DOCKERCMD) down --rmi local