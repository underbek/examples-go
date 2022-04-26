## Пример сервиса, который создает и возвращает пользователя

### Dependency
```shell
git clone https://github.com/googleapis/googleapis

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

### Generate simple
```shell
protoc \
  --go_out=proto \
  --go-grpc_out=proto \
  proto/*.proto
```

### Evans
```shell
evans --port 8000 -r repl
```

### Generate with gateway
```shell
protoc -I. -I./googleapis \
  --go_out=proto \
  --go-grpc_out=proto \
  --grpc-gateway_out=logtostderr=true:proto \
  --openapiv2_out=logtostderr=true,repeated_path_param_separator=ssv:./api \
  proto/*.proto
```

### Совместимость

1. Не редактируте теги (поля можно перименовыать, т.к. имеет значение только сам тег)
2. Добовляйте новые поля (старый код не будет их использовать - нужно это учитывать)
3. Обновляйте версию
4. Помечайте старые поля как deprecated
5. Обновляйте версию
6. Удаляйте поля
7. Обновляйте версию

### Тестируем из консоли
CLI
```shell
brew install grpc
grpc ls localhost:8000
```

### Ссылки
* [grpc gateway](https://github.com/grpc-ecosystem/grpc-gateway)
* [grpc gateway на русском](https://habr.com/ru/post/496574/)
* [альтернатива генерации http из grpc (clay)](https://github.com/utrack/clay)
* [работа с ошибками](https://jbrandhorst.com/post/grpc-errors/)
*[grpc mock](https://github.com/tokopedia/gripmock)
* [grpc middlewares](https://github.com/grpc-ecosystem/go-grpc-middleware)
* [grpc prometheus](https://github.com/grpc-ecosystem/go-grpc-prometheus)
* [grpc-awesome](https://github.com/grpc-ecosystem/awesome-grpc)
* [micro](https://github.com/micro/micro)
* [evans](https://github.com/ktr0731/evans)

* [про хелсчеки](https://github.com/grpc/grpc/blob/master/doc/health-checking.md)
* [готовый хендлер](https://pkg.go.dev/google.golang.org/grpc/health)
* [проба для кубера](https://github.com/grpc-ecosystem/grpc-health-probe)
