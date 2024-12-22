# Yandex Metric first iteration

## Запустить сервер для сбора метрик в память

В первом инкременте реализован простейший сервер, который принимает метрики и сохраняет их в памяти.

Используется одна ручка `/update/` для отправки метрик. Метрики принимаются методи PUSH в параметрах строки.

Для проверки VET
go vet -vettool=$(which statictest) ./...

Для сборки сервера используется команда 'go build -ldflags "-X main.buildVersion=1.0.0 -X main.buildDate=$(date +%Y-%m-%d) -X main.buildCommit=$(git rev-parse HEAD)" -o ./cmd/server/server ./cmd/server'
Данная команда устанавливает версию, дату и коммит в бинарник.
