IMAGE_NAME = daviuefs/iot-ship-sensores

.PHONY: build run stop logs help run-one run-many stop-many

build:
	docker build -t $(IMAGE_NAME) -f dockerfile ..

run:
	docker pull $(IMAGE_NAME):latest
	@read -p "Nome do container: " CONTAINER_NAME; \
	read -p "SENSOR_ID: " SENSOR_ID; \
	read -p "SENSOR_TIPO (temperatura|combustivel|porcao): " SENSOR_TIPO; \
	read -p "BROKER_IP (host:port): " BROKER_IP; \
	docker run -it --name $$CONTAINER_NAME \
		-e SENSOR_ID=$$SENSOR_ID \
		-e SENSOR_TIPO=$$SENSOR_TIPO \
		-e BROKER_IP=$$BROKER_IP \
		$(IMAGE_NAME)

# ---------- NOVOS TARGETS (automação) ----------

# Uso:
# make run-one CONTAINER_NAME=sensor-001 SENSOR_ID=sensor-001 SENSOR_TIPO=temperatura BROKER_IP=172.16.103.1:9000
run-one:
	@docker pull $(IMAGE_NAME):latest
	@test -n "$(CONTAINER_NAME)" || (echo "CONTAINER_NAME obrigatório" && exit 1)
	@test -n "$(SENSOR_ID)" || (echo "SENSOR_ID obrigatório" && exit 1)
	@test -n "$(SENSOR_TIPO)" || (echo "SENSOR_TIPO obrigatório" && exit 1)
	@test -n "$(BROKER_IP)" || (echo "BROKER_IP obrigatório" && exit 1)
	docker run -d --name $(CONTAINER_NAME) \
		-e SENSOR_ID=$(SENSOR_ID) \
		-e SENSOR_TIPO=$(SENSOR_TIPO) \
		-e BROKER_IP=$(BROKER_IP) \
		$(IMAGE_NAME)

# Uso:
# make run-many N=10 BROKER_IP=172.16.103.1:9000 PREFIX=sensor
# Tipos válidos no seu código: temperatura, combustivel, porcao
run-many-lab:
	docker pull $(IMAGE_NAME):latest
	@N=$${N:-20}; \
	PREFIX=$${PREFIX:-sensor}; \
	BROKER_IP=$${BROKER_IP:-172.16.103.8:4242}; \
	echo "Subindo $$N sensores (UDP) em $$BROKER_IP (prefixo=$$PREFIX)"; \
	i=1; \
	while [ $$i -le $$N ]; do \
		ID=$$(printf "%s-%03d" "$$PREFIX" $$i); \
		case $$(( ( $$i - 1 ) % 3 )) in \
			0) TIPO=temperatura ;; \
			1) TIPO=combustivel ;; \
			2) TIPO=porcao ;; \
		esac; \
		echo "-> $$ID (tipo=$$TIPO)"; \
		docker run -d --restart unless-stopped --name $$ID \
			-e SENSOR_ID=$$ID \
			-e SENSOR_TIPO=$$TIPO \
			-e BROKER_IP=$$BROKER_IP \
			$(IMAGE_NAME) || exit 1; \
		i=$$((i+1)); \
	done; \
	echo "OK"

stop:
	@read -p "Nome do container a parar: " CONTAINER_NAME; \
	docker stop $$CONTAINER_NAME && docker rm $$CONTAINER_NAME

# Uso:
# make stop-many N=100 PREFIX=sensor
stop-many:
	@N=$${N:-20}; \
	PREFIX=$${PREFIX:-sensor}; \
	i=1; \
	while [ $$i -le $$N ]; do \
		ID=$$(printf "%s-%03d" "$$PREFIX" $$i); \
		echo "-> parando $$ID"; \
		docker rm -f $$ID >/dev/null 2>&1 || true; \
		i=$$((i+1)); \
	done; \
	echo "OK"

logs:
	@read -p "Nome do container para ver logs: " CONTAINER_NAME; \
	docker logs -f $$CONTAINER_NAME

help:
	@echo "Comandos disponíveis:"
	@echo "  make build     - Constrói a imagem $(IMAGE_NAME)"
	@echo "  make run       - Interativo (1 sensor)"
	@echo "  make run-many-lab - Sobe N sensores automaticamente"
	@echo "  make stop-many - Remove N sensores automaticamente"