# APM — Awesome Package Manager for Windows

<p align="center">
  <strong>🚀 Быстрый и удобный пакетный менеджер для Windows</strong>
</p>

```
     █████╗ ██████╗ ███╗   ███╗
    ██╔══██╗██╔══██╗████╗ ████║
    ███████║██████╔╝██╔████╔██║
    ██╔══██║██╔═══╝ ██║╚██╔╝██║
    ██║  ██║██║     ██║ ╚═╝ ██║
    ╚═╝  ╚═╝╚═╝     ╚═╝     ╚═╝
```

## 📋 О проекте

APM (Awesome Package Manager) — это CLI-утилита для установки, удаления и поиска программ в Windows. Написана на Go, компилируется в один `.exe` без зависимостей.

**Особенности APM:**
- ⚡ **Параллельная загрузка**: Скачивайте десятки пакетов одновременно.
- 🛡️ **Проверка SHA256**: Встроенная проверка целостности файлов.
- 📦 **Пользовательские репозитории**: Создавайте и делитесь своими списками пакетов.
- **30+ популярных пакетов** уже в реестре: Chrome, Firefox, VLC, VS Code, Node.js, Telegram, Discord и многое другое.

## ⚡ Быстрый старт

```bash
# Поиск пакета
apm search obs
# Вывод: obs-studio

apm search google
# Вывод: google-chrome, google-drive

# Установка
apm install obs-studio
apm install google-chrome

# Удаление
apm remove obs-studio

# Список установленных
apm list

# Все доступные пакеты
apm list-all
```

## 🛠 Команды

| Команда | Алиас | Описание |
|---------|-------|----------|
| `apm install <package>` | `apm i` | Установить пакет |
| `apm remove <package>` | `apm rm` | Удалить пакет |
| `apm search <query>` | `apm s` | Поиск пакетов |
| `apm list` | `apm ls` | Установленные пакеты |
| `apm list-all` | | Все доступные пакеты |
| `apm update` | `apm u` | Обновить локальный реестр пакетов с GitHub |
| `apm upgrade` | | Обновить сам пакетный менеджер APM |
| `apm upgrade-all` | | Обновить все установленные пакеты |
| `apm export <file>` | | Экспортировать список пакетов |
| `apm import <file>` | | Установить пакеты из списка |
| `apm repo add <name> <url>` | | Добавить пользовательский репозиторий |
| `apm repo rm <name>` | | Удалить пользовательский репозиторий |
| `apm repo ls` | | Список пользовательских репозиториев |
| `apm selfdestruct`| | Полностью удалить APM |
| `apm version` | `apm -v` | Версия APM |
| `apm help` | `apm -h` | Справка |

## [📦 Доступные пакеты](https://github.com/Tiwerrrr/apm/blob/main/data/registry.json)

## 🔧 Сборка из исходников

```bash
# Требуется Go 1.22+
git clone https://github.com/Tiwerrrr/apm.git
cd apm
go build -o apm.exe .
```

## 📁 Структура проекта

```
APM/
├── main.go                        # Точка входа CLI
├── go.mod                         # Go модуль
├── data/
│   └── registry.json              # Реестр пакетов (30+ программ)
└── internal/
    ├── commands/
    │   ├── install.go             # Команда install
    │   ├── remove.go              # Команда remove
    │   ├── search.go              # Команда search
    │   ├── update.go              # Команды update
    │   ├── upgrade.go             # Команды upgrade / upgrade-all
    │   ├── export.go              # Команды export / import
    │   ├── selfdestruct.go        # Команда selfdestruct
    │   └── list.go                # Команды list / list-all
    ├── config/
    │   └── config.go              # Пути и база установленных пакетов
    ├── console/
    │   └── console.go             # Цветной вывод, таблицы, прогресс-бар
    ├── downloader/
    │   └── downloader.go          # Скачивание файлов с прогресс-баром
    ├── installer/
    │   └── installer.go           # Silent-установка и удаление
    └── registry/
        └── registry.go            # Загрузка и поиск по реестру
```

## ⚙️ Как это работает

1. **Search** — ищет по ID, имени, тегам и описанию с ранжированием результатов
2. **Install** — скачивает установщик с прогресс-баром → запускает тихую установку (`/S`, `/quiet`) → записывает в базу
3. **Remove** — находит деинсталлятор в реестре Windows → запускает тихое удаление → убирает из базы

## 📜 Лицензия

MIT
