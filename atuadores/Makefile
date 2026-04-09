IMAGE_NAME = daviuefs/iot-ship-atuadores

.PHONY: build run stop help run-many-lab stop-many

build:
	docker build -t $(IMAGE_NAME) -f dockerfile ..

run:
	@docker pull $(IMAGE_NAME):latest
	@read -p "Nome do container: " CONTAINER_NAME; \
	read -p "ACTUATOR_ID: " ACTUATOR_ID; \
	read -p "BROKER_ADDR: " BROKER_ADDR; \
	docker run -d --name $$CONTAINER_NAME \
		-e ACTUATOR_ID=$$ACTUATOR_ID \
		-e BROKER_ADDR=$$BROKER_ADDR \
		$(IMAGE_NAME)

stop:
	@read -p "Nome do container a parar:" CONTAINER_NAME; \
	docker stop $$CONTAINER_NAME && docker rm $$CONTAINER_NAME

# ---------- ATALHOS HOME/LAB (sem prompts) ----------

LAB_IP  = 172.16.103.8
TCP_PORT = 4343

ACTUATOR_CONTAINER_LAB  = iot-atuador-lab

ACTUATOR_ID = exaustor-01

run-lab:
	@docker pull $(IMAGE_NAME):latest
	@docker rm -f $(ACTUATOR_CONTAINER_LAB) >/dev/null 2>&1 || true
	@docker run -d --name $(ACTUATOR_CONTAINER_LAB) \
		-e ACTUATOR_ID=$(ACTUATOR_ID) \
		-e BROKER_ADDR=$(LAB_IP):$(TCP_PORT) \
		$(IMAGE_NAME)
	@docker logs -f $(ACTUATOR_CONTAINER_LAB)


stop-lab:
	@docker rm -f $(ACTUATOR_CONTAINER_LAB) >/dev/null 2>&1 || true
	@echo OK
	
# Uso:
# make run-many-lab N=10 PREFIX=atuador LAB_IP=172.16.201.9 TCP_PORT=4343
run-many-lab:
	@docker pull $(IMAGE_NAME):latest
	@N=$${N:-100}; \
	PREFIX=$${PREFIX:-atuador}; \
	BROKER_ADDR=$${BROKER_ADDR:-$(LAB_IP):$(TCP_PORT)}; \
	echo "Subindo $$N atuadores (TCP) em $$BROKER_ADDR (prefixo=$$PREFIX)"; \
	i=1; \
	while [ $$i -le $$N ]; do \
		ID=$$(printf "%s-%03d" "$$PREFIX" $$i); \
		echo "-> $$ID"; \
		docker rm -f $$ID >/dev/null 2>&1 || true; \
		docker run -d --restart unless-stopped --name $$ID \
			-e ACTUATOR_ID=$$ID \
			-e BROKER_ADDR=$$BROKER_ADDR \
			$(IMAGE_NAME) || exit 1; \
		i=$$((i+1)); \
	done; \
	echo "OK"

# Uso:
# make stop-many N=20 PREFIX=atuador
stop-many:
	@N=$${N:-100}; \
	PREFIX=$${PREFIX:-atuador}; \
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
	@echo "  make build  - Constrói a imagem $(IMAGE_NAME)"
	@echo "  make run    - Pede nome do container e variáveis de ambiente, depois roda o container"
	@echo "  make stop   - Para e remove o container"
	@echo "  make run-lab  - Roda atuador pré-configurado para ambiente LAB"
	@echo "  make stop-lab - Para atuador pré-configurado para ambiente LAB"
	@echo "  make run-many-lab - Sobe N atuadores automaticamente no LAB"
	@echo "  make stop-many - Remove N atuadores automaticamente"
	@echo "  make logs    - Pede nome do container e mostra os logs em tempo real"
