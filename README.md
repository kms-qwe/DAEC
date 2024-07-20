# DAEC

## Что это? 

Распределенный вычислитель арифметических выражений

### Описание

Пользователь хочет считать арифметические выражения. Он вводит строку `2 + 2 * 2` и хочет получить в ответ `6`. Но наши операции сложения и умножения (также деления и вычитания) выполняются "очень-очень" долго. Поэтому вариант, при котором пользователь делает http-запрос и получает в качетсве ответа результат, невозможна. Более того, вычисление каждой такой операции в нашей "альтернативной реальности" занимает "гигантские" вычислительные мощности. Соответственно, каждое действие мы должны уметь выполнять отдельно и масштабировать эту систему можем добавлением вычислительных мощностей в нашу систему в виде новых "машин". Поэтому пользователь может с какой-то периодичностью уточнять у сервера "не посчиталость ли выражение"? Если выражение наконец будет вычислено - то он получит результат. Помните, что некоторые части арфиметического выражения можно вычислять параллельно.

### Как работает?
Запускается HTTP сервер (auth), который обрабатывает запросы пользователей. 
Запускается gRPC сервер (оркесратор) и gRPC клиент (агент); оркестратор из бд берет выражение и отправляет на вычисление агенту, агент вычисляет и отправляет результат оркестратору. 

Auth имеет следующие эндпоинты для пользователя:
- Регистрация пользователя
```commandline
curl --location 'localhost:8080/api/v1/register' \
--header 'Content-Type: application/json' \
--data '{
      "login": "login",
      "password": "password"

}'
``` 
- Получение JWT Токена

```commandline
curl --location --request GET 'localhost:8080/api/v1/login' \
--header 'Content-Type: application/json' \
--data '{
      "login": "login",
      "password": "password"
}'

``` 

- Добавление вычисления арифметического выражения
 
```commandline
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--header 'Authorization: Bearer YOUR_JWT_TOKEN' \
--data '{
      "expression": "1 + 1"
}'
```
 
- Получение списка выражений
 
```commandline
curl --location 'localhost:8080/api/v1/expressions' \
--header 'Authorization: Bearer YOUR_JWT_TOKEN' 

```

- Получение выражения по его идентификатору
 
```commandline
curl --location 'localhost:8080/api/v1/expression?id=1' \
--header 'Authorization: Bearer YOUR_JWT_TOKEN' 
 
```



### Параллелизм вычислений

Выражения считается быстрее, если расставить скобки, например 2 + 2 + 2 + 2 будет выполняться последовательно, т.к. операции считаются равноправными. Но (2 + 2) + (2 + 2) будет считаться в два раза быстрее, т.к. выражение распадается на два независимых подвыражения. 

## Деплой

### 1 Клонирования репозитория

```commandline
git clone https://github.com/kms-qwe/yandex-lyceum-go
```
### 2 Reset database
```commandline
./dbreset.sh
```
### 3 Запуск всех приложений (в трех терминалах, чтобы логи писались)
```commandline
./runAuth.sh
./runOrch.sh
./runAgent.sh
```

## Тесты

- Valid cases
    1. 4 + (-2) + 5 * 6

    2. 2 + 2 + 2 + 2

    3. 2 + 2 * 4 + 3 - 4 + 5

    4. (23 + 125) - 567 * 23

    5. -3 +6

- Invalid cases
    1. 4 / 0

    2. 45 + x - 5

    3. 45 + 4*

    4. ---4 + 5

    5. 52 * 3 /
















