# 🚢 IoT-Ship
Este repositório simula o funcionamento de um navio integrado à Internet das Coisas.Ele é distribuído com quatro componentes principais:
- `broker`: recebe telemetria dos sensores e intermedia comandos.
- `sensores`: geram dados periódicos e enviam para o broker via UDP.
- `atuadores`: recebem comandos (ligar/desligar) via broker.
- `clientes`: interface em terminal para consultar estado e enviar comandos.

Este README é um manual de uso básico, focado em:
- conexão entre os containers;
- uso dos `Makefile` de cada pasta para usuário comum (`make run` e `make stop`);
- uso detalhado do container `clientes`.

## 1) 🔌 Visão Rápida da Conexão
- Sensores -> Broker: `UDP` (telemetria), porta configurada por `UDP_PORT` no broker.
- Cliente -> Broker: `TCP` (query, comandos, ACK, notificações), porta `TCP_PORT`.
- Atuador -> Broker: `TCP` (handshake, recebimento de comando e envio de estado).

Fluxo básico:
1. `broker` sobe escutando UDP e TCP.
2. `sensores` enviam dados no formato JSON para o UDP do broker.
3. `atuadores` e `clientes` conectam por TCP e fazem handshake com `ID` único.
4. `cliente` pede atuadores, visualiza sensores em tempo real e envia comandos.
5. `broker` confirma comandos com `ACK` ao cliente e encaminha para atuador.

Detalhes de robustez implementados no código:
- handshake para cliente/atuador (rejeita `ID` duplicado);
- `ping/pong` para detectar desconexão em conexões TCP;
- reconexão automática em cliente e atuador;
- remoção de sensor por TTL no broker (sensor é removido após 20s sem enviar dados).

## 2) 📁 Manual por Pasta/Container

### 2.1 `broker/`
Responsabilidade:
- receber sensores via UDP;
- registrar cliente e atuador via TCP;
- encaminhar atualizações para clientes;
- enviar comandos para atuadores.

Variáveis de ambiente usadas no `run`:
- `UDP_PORT` (padrão `9000`)
- `TCP_PORT` (padrão `9001`)

Comandos para usuário comum:
```bash
make run
make run-lab
```
O `make run` vai pedir:
- nome do container;
- `UDP_PORT`;
- `TCP_PORT`.

Ele automaticamente puxa a imagem do Dockerhub, cria e sobe o container.
O run-lab cria com parâmetros já definidos, para facilitação em testes. Modifique o IP e porta no Makefile para o do seu computador. 

Para parar/remover:
```bash
make stop
make stop-lab
```
Que irá pedir o nome do container a ser removido. stop-lab remove diretamente o container criado pelo run-lab.

#### Deadlines

As conexões TCP usam a lógica de ping/pong. A cada 3s, o servidor envia um pacote JSON com uma mensagem "ping". Se o cliente/atuador não responder pong em 4s, a conexão é fechada e o registro sai do map do broker. 
Para UDP, usa TTL. Se o último pacote de um sensor foi há 20s, a rotina de monitoramento remove-o do map.

Para o handshake inicial, o tempo limite é de 5 segundos.

### 2.2 `sensores/`
Responsabilidade:
- simular sensores e enviar dados periódicos para o broker (UDP).

Variáveis de ambiente usadas no `run`:
- `SENSOR_ID` (identificador)
- `SENSOR_TIPO` (`temperatura`, `combustivel` ou `porcao`)
- `BROKER_IP` (formato `host:porta`, ex.: `127.0.0.1:9000`)

Comandos para usuário comum:
```bash
make run
make run-lab
make run-many-lab

```
Para parar/remover:
```bash
make stop
make stop-lab
make stop-many
```
O make run-many-lab cria N cópias do container e sobe, cada um com um ID diferente. Use para testar muitas conexões ao broker. No caso dos sensores, eles são criados em tipos aleatórios. Modifique o IP:porta e o N no seu Makefile (padrão N=100).
O make stop-many remove os containeres criados.

Observação importante:
- no código atual, os tipos aceitos são `temperatura`, `combustivel` e `porcao` (mais em breve).

### 2.3 `atuadores/`
Responsabilidade:
- conectar-se ao broker via TCP;
- receber comando (`ligar`, `desligar`, `status`);
- manter estado atual e reportar para o broker.

Variáveis usadas no `run`:
- `ACTUATOR_ID` (identificador do atuador)
- `BROKER_ADDR` (formato `host:porta`, ex.: `127.0.0.1:9001`)

Comandos para usuário comum:
```bash
make run
make run-lab
make run-many-lab

```
Para parar/remover:
```bash
make stop
make stop-lab
make stop-many
```

### 2.4 🖥️ `clientes/`
Responsabilidade:
- interface de operação do sistema;
- escutar mensagens de sensores/atuadores;
- enviar consultas e comandos para o broker;
- exibir notificações em tempo real.

Variáveis usadas no `run`:
- `CLIENT_ID` (ID único do cliente no broker)
- `BROKER_ADDR` (formato `host:porta`, ex.: `127.0.0.1:9001`)

Comandos para usuário comum:
```bash
make run
make run-lab
make run-many-lab
```
Para parar/remover:
```bash
make stop
make stop-lab
make stop-many
```

#### 📋 Como usar o menu do cliente
Ao iniciar, o cliente conecta no broker por TCP, faz handshake e abre o menu:
1. `Pedir atuadores`
2. `Mostrar atuadores`
3. `Enviar comando para atuador`
4. `Mostrar sensores`
5. `Ver notificações`
6. `Sair`


<img width="383" height="200" alt="Captura de tela de 2026-04-09 13-50-49" src="https://github.com/user-attachments/assets/bc698238-8323-45e0-8d24-53d373b9a604" />
</br>

**1) Pedir atuadores**: pede ao broker a lista de atuadores conectados à ele. Antes do broker responder o cliente, ele atualiza todos os atuadores no map.

**2) Mostrar atuadores**: Mostra todos os atuadores conectados, seu estado e última ação.

**3) Enviar comandos para atuador**: Envia comando a um atuador. Você deve digitar o ID do atuador desejado e depois a ação. O minimenu irá aguardar o ACK do broker, informando se obteve sucesso ou não.

</br>
<img width="360" height="205" alt="Captura de tela de 2026-04-09 13-56-17" src="https://github.com/user-attachments/assets/10d75e21-bfca-4c20-95b6-ee794e4521fe" />
</br>
<img width="360" height="205" alt="Captura de tela de 2026-04-09 13-56-22" src="https://github.com/user-attachments/assets/ea267b55-4dff-4ef2-8dfb-2b613ae26ec4" />
</br>
<img width="443" height="200" alt="Captura de tela de 2026-04-09 13-56-38" src="https://github.com/user-attachments/assets/70f2d949-289c-44a3-8ae8-9bc29b57a413" />
</br></br>

**4) Mostrar sensores**: Mostra todos os sensores conectados. Não é necessário pedir ao broker, eles chegam automaticamente.

**5) Ver notificações**: Exibe as notificações. Ao chegar uma, a cor muda para amarelo e um som é tocado.

<img width="360" height="205" alt="Captura de tela de 2026-04-09 13-53-14" src="https://github.com/user-attachments/assets/671e66e9-7348-4fcb-b5f0-680f65cb8af4" />\
</br>

**6) Sair:** Encerra o progama. Também desconecta do broker.

Fluxo recomendado de uso:
1. Suba `broker`, pelo menos um `atuador`, e um ou mais `sensores`.
2. Suba o `cliente`.
3. No cliente, use `1) Pedir atuadores`.
4. Use `2) Mostrar atuadores` para confirmar IDs e estado.
5. Use `3) Enviar comando para atuador` para `ligar`/`desligar`.
6. Use `4) Mostrar sensores` para acompanhar telemetria em tempo real.
7. Use `5) Ver notificações` para eventos de adição/remoção e atualizações.

#### ⚠️ Observações
O "cache"/map local de atuadores só atualiza ao realizar a opção 1. Lembre-se de utilizá-la antes e após enviar um comando. **(fix em breve)**.

Fique sempre atento às notificações. Elas podem conter informações críticas, como remoção de sensores/atuadores.

#### 🔍 Conectividade/transporte
- Toda requisição importante do cliente gera um `request_id` único. Ele é usado para os ACKs serem válidos para apenas aquele ID.
- O broker responde com `ACK` (`ok=true/false`) para confirmar ou recusar.
- O cliente possui reconexão automática se o broker cair/voltar.
- O menu é interrompido por eventos de notificação para não perder atualizações.

#### ❌ Erros comuns no cliente
- `ID` duplicado: o broker rejeita handshake. Troque `CLIENT_ID` e suba um novo container.
- broker indisponível: o cliente está tentando reconectar. Aguarde até que estabilize.
- atuador não encontrado: confira se o atuador está rodando e use `Pedir atuadores` antes de enviar comando.


## 3) 🗂️ Estrutura das Pastas
- `broker/`: lógica de roteamento, sessões TCP, TTL e ACK.
- `sensores/`: simulação e envio UDP.
- `atuadores/`: estado do atuador e processamento de comandos.
- `clientes/`: UI em terminal, notificações, visualização e comandos.
- `shared/`: structs e funções compartilhadas (handshake, escrita JSON, retry).
