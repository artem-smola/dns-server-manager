# dns-server-manager

Клиент-серверное приложение на Go для управления DNS-серверами через `resolv.conf`.

## Возможности

- добавить DNS-сервер;
- удалить DNS-сервер;
- получить список DNS-серверов;
- healthcheck сервера;
- CLI-клиент.

## Сборка

```bash
go build -o bin/dns-server ./cmd/server
go build -o bin/dns-client ./cmd/client
```

## Запуск

### Сервер

```bash
./bin/dns-server --addr :8080 --resolv-conf /etc/resolv.conf --log-file ./dns-server.log
```

Если сервер работает с системным файлом `/etc/resolv.conf`, обычно требуется запуск с `sudo`:

```bash
sudo ./bin/dns-server --addr 127.0.0.1:8080 --resolv-conf /etc/resolv.conf --log-file ./dns-server.log
```

Перед изменениями рекомендуется сделать бэкап:

```bash
sudo cp /etc/resolv.conf /etc/resolv.conf.bak
```

Если нужно откатить изменения:

```bash
sudo cp /etc/resolv.conf.bak /etc/resolv.conf
```

Флаги сервера:

- `--addr` — адрес прослушивания (по умолчанию `:8080`);
- `--resolv-conf` — путь к файлу `resolv.conf` (по умолчанию `/etc/resolv.conf`);
- `--log-file` — путь к файлу логов (если не указан, лог в stdout).

### Клиент

```bash
./bin/dns-client --help
./bin/dns-client --server http://127.0.0.1:8080 list
./bin/dns-client --server http://127.0.0.1:8080 add 8.8.8.8
./bin/dns-client --server http://127.0.0.1:8080 remove 8.8.8.8
```

Флаги клиента:

- `--server` — базовый URL dns-manager сервера (по умолчанию `http://127.0.0.1:8080`).
- `--help` — показать справку по CLI-клиенту (команды и доступные флаги).

## API

Сервер предоставляет REST API.

- `GET /dns`
- `POST /dns` с JSON вида `{"server":"8.8.8.8"}`
- `DELETE /dns` с JSON вида `{"server":"8.8.8.8"}`
- `GET /healthz`

## Unit-тесты

```bash
go test ./...
```