# Makefile for multi-carrier shipping stack

DOCKER_COMPOSE = docker compose

.PHONY: all start start-all build up down logs restart clean

all: up

start: up

start-all: up

build:
	$(DOCKER_COMPOSE) build

up:
	$(DOCKER_COMPOSE) up --build

down:
	$(DOCKER_COMPOSE) down

logs:
	$(DOCKER_COMPOSE) logs -f

restart: down up

clean:
	$(DOCKER_COMPOSE) down --rmi all --volumes --remove-orphans
