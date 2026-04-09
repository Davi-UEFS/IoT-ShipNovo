IMAGE_NAME = daviuefs/iot-ship-broker

.ONESHELL:

.PHONY: build run stop help

build:
	docker build -t $(IMAGE_NAME) -f dockerfile ..

run:
	docker pull $(IMAGE_NAME):latest
	@read -p "Nome do container: " CONTAINER_NAME; \
	read -p "UDP_PORT [9000]: " UDP_PORT; \
	UDP_PORT=$${UDP_PORT:-9000}; \
	read -p "TCP_PORT [9001]: " TCP_PORT; \
	TCP_PORT=$${TCP_PORT:-9001}; \
	docker run -d --name $$CONTAINER_NAME \
		-p $$UDP_PORT:$$UDP_PORT/udp \
		-p $$TCP_PORT:$$TCP_PORT \
		-e UDP_PORT=$$UDP_PORT \
		-e TCP_PORT=$$TCP_PORT \
		$(IMAGE_NAME)

stop:
	@read -p "Nome do container a parar: " CONTAINER_NAME; \
	docker stop $$CONTAINER_NAME && docker rm $$CONTAINER_NAME

# ---------- ATALHOS HOME/LAB (sem prompts) ----------

HOME_IP = 192.168.18.13
LAB_IP  = 172.16.103.11
UDP_PORT = 4242
TCP_PORT = 4343

BROKER_CONTAINER_LAB  = iot-broker-lab

run-lab:
	@docker pull $(IMAGE_NAME):latest
	@docker rm -f $(BROKER_CONTAINER_LAB) >/dev/null 2>&1 || true
	@docker run -d --restart unless-stopped --name $(BROKER_CONTAINER_LAB) \
		-p $(UDP_PORT):$(UDP_PORT)/udp \
		-p $(TCP_PORT):$(TCP_PORT) \
		-e UDP_PORT=$(UDP_PORT) \
		-e TCP_PORT=$(TCP_PORT) \
		$(IMAGE_NAME)
	@docker logs -f $(BROKER_CONTAINER_LAB)

stop-lab:
	@docker rm -f $(BROKER_CONTAINER_LAB) >/dev/null 2>&1 || true
	@echo "OK"

logs:
	@read -p "Nome do container para ver logs: " CONTAINER_NAME; \
	docker logs -f $$CONTAINER_NAME

help:
	@echo "Comandos disponíveis:"
	@echo "  make build  - Constrói a imagem $(IMAGE_NAME)"
	@echo "  make run    - Pede nome do container e portas, depois roda o container"
	@echo "  make stop   - Para e remove o container"
	@echo "  make help   - Mostra esta ajuda"
	@echo "  make run-lab  - Roda broker pré-configurado para ambiente LAB"
	@echo "  make stop-lab - Para broker pré-configurado para ambiente LAB"