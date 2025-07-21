# Go-Ragger

[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Demo](https://img.shields.io/badge/demo-live-green)](https://ragger.fondorg.ru/)

Go-Ragger - это веб-приложение для организации базы знаний на основе RAG (Retrieval-Augmented Generation) AI.

## 🚀 Демо

Работающее демо доступно [здесь](https://ragger.fondorg.ru/)

## ✨ Возможности

- 🔐 Регистрация и авторизация пользователей
- 📂 Загрузка, просмотр и скачивание документов (PDF, DOCX, XLSX, PPTX, JPG, PNG и др.)
- 🤖 ИИ-ассистент для ответов на вопросы по документам

## 🛠️ Технологический стек

- **Модульная структура**: [Uber FX](https://uber-go.github.io/fx/)
- **Бизнес-логика**: [Eino Framework](https://www.cloudwego.io/docs/eino/)
- **Web-фреймворк**: [Data-Star](https://data-star.dev/)
- **Векторная БД**: [Milvus](https://milvus.io/)
- **Чанкинг документов**: [Chunkr.ai](https://chunkr.ai/)
- **Автоматизация задач**: [Taskfile](https://taskfile.dev/)
- **Kuberentes-разработка**: [DevSpace](https://www.devspace.sh/)

## 📝 Как это работает

1. Пользователь загружает документы в хранилище
2. Документы разбиваются на чанки "умными методами" с помощью сервиса Chunkr.ai
3. Чанки векторизуются с помощью LLM и сохраняются в Milvus
4. При запросе к ИИ-ассистенту:
   - Векторная БД ищет релевантные запросу чанки
   - Найденные чанки подаются в LLM как контекст для формирования ответа

## ⚙️ Требования

Для работы приложения необходимо:

- Доступ по API к сервису [chunkr.ai](https://chunkr.ai/) (можно развернуть локально - [инструкции](https://github.com/lumina-ai-inc/chunkr/tree/main/kube))
- Доступ по API к OpenAI-совместимому API (например, [Ollama](https://ollama.com/))
- Установленный [DevSpace](https://www.devspace.sh/)
- Работающий Kubernetes-кластер (можно использовать [Kind](https://kind.sigs.k8s.io/))

## 🚀 Запуск проекта

### Первоначальная настройка

При первом запуске Devspace запросит необходимые переменные:

- Имя Docker Hub аккаунта
- Виртуальный хост для Ingress
- Модель для эмбеддинга (рекомендуется дефолтная)
- Модель для чата (рекомендуется дефолтная)
- URL OpenAI-совместимого сервиса
- API-ключ OpenAI-совместимого сервиса
- URL API сервиса Chunkr
- API-ключ сервиса Chunkr

Devspace сохранит введенные ответы для последующих запусков.

Для сброса переменных:
```bash
 devspace reset vars
 ```


### Разработка
1. Запустить devspace deploy:
```bash
devspace deploy --namespace ragger 
```
- создаст заданный namespace если не существует
- соберет докер-образы
- развернет prod деплоймент в заданном кластере и неймспейсе вместе с дополнительными сервисами (postgresql и milvus)

2. Запустить devspace dev:
```bash
devspace dev
```
- запустит терминал в контейнере приложения.
 
3. В терминале DevSpace выполнить:
```bash
task reset-all      # Сгенерировать код, резетнуть и засидить бд
task                # Запустить приложение в dev-режиме с hot-reload
```

4. Приложение будет доступно по адресу: http://localhost:8383
#### Полезные команды разработки (в Taskfile.yml):
```bash
task db:seed # создать сиды в БД
task sqlc:generate # сгенерировать SQLC код
task reset-all # сбросить и засидить БД, перегенерировать код
```

### Деплоймент
1. Запустить devspace:
```bash
devspace reset vars # сбросить переменные devspace на всякий случай
devspace deploy --namespace ragger # запустить prod деплоймент в namespace ragger
```
- соберет prod-образ приложения и образ миграций
- развернет в заданном кластере и неймспейсе вместе с дополнительными сервисам (postgresql и milvus)
- запустит init-конейнер для применения миграций
- запустит приложение
- приложение будет доступно по адресу виртуального хоста

## 📄 Лицензия

MIT License - см. [LICENSE](LICENSE)
