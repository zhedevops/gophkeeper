# gophkeeper

**Запуск сервера:**
```
make server
```

**Команды клиента:**

После ввода команд регистрации и аутентификации нужно будет ввести пароль. 

Регистрация пользователя:
```
make register login=test
```

Аутентификация пользователя:
```
make login login=test
```

Сохранение логина/пароля:
```
make create-credentials type=credentials meta="логин и пароль от ВК" login=xxx password=secret1
```

Сохранение текстовых данных:
```
make create-text type=text meta="секретный ключ от банка, где деньги лежат" data="secret very big"
```

Сохранение бинарных данных:
* Через файл:
```
make create-binary-file type=binary meta="файл совершенно секретный" file="README.md"
```
* Текстом:
```
make create-binary-data type=binary meta="файл совершенно секретный" data="важная информация"
```

Сохранение банковской карты:
```
make create-card type=card meta="моя карта" number="1234567890" holder="Ivan Ivanov" expiry=2026.06.07 cvv=123
```

Получение данных:
```
make get id=16
```

Получение списка пользовательских данных:
```
make list
```

Удаление пользовательских данных:
```
make delete id=16
```

Создание сертификата для подключения по TLS:
```
make certs
```