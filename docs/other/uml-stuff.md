## 1. Архитектура Encoder сервиса
```mermaid
graph TB
    subgraph "Encoder Service"
        Service[Service<br/>Основной сервис]
        Watchers[Watchers<br/>N горутин]
        Scheduler[Scheduler<br/>Планировщик]
        Backlog[Backlog<br/>Очередь задач]
        WorkerPool[Worker Pool<br/>N воркеров]
        Cleanup[Cleanup<br/>Фоновая очистка]
    end
    
    subgraph "External Services"
        Composer[Composer Service<br/>gRPC + HTTP]
        FFmpeg[FFmpeg<br/>Кодирование]
        Storage[File Storage<br/>Временные файлы]
    end
    
    Service --> Watchers
    Service --> Scheduler
    Service --> WorkerPool
    Service --> Cleanup
    
    Watchers -->|GetTask| Composer
    Watchers -->|Prefetch| Storage
    Watchers --> Backlog
    
    Backlog --> Scheduler
    Scheduler --> WorkerPool
    
    WorkerPool -->|Encode| FFmpeg
    WorkerPool -->|Upload| Composer
    WorkerPool -->|UpdateProgress| Composer
    WorkerPool -->|FinishTask| Composer
    
    Cleanup --> Storage
    
    style Service fill:#e1f5ff
    style Composer fill:#fff4e1
    style FFmpeg fill:#ffe1f5
    style Storage fill:#e1ffe1
```
## 2. Диаграмма последовательности обработки задачи
```mermaid
sequenceDiagram
    participant W as Watcher
    participant C as Composer
    participant P as Prefetch
    participant B as Backlog
    participant S as Scheduler
    participant Worker as Worker
    participant F as FFmpeg
    participant HTTP as HTTP API
    
    W->>+C: GetTask(encoder, hostname, ffmpeg_version)
    C--)-W: Response
    alt Задача получена
        W->>+P: Prefetch(task)
        P->>+C: Download(source_url)
        C--)-P: File data
        P--)-W: Local file path
        W->>B: Add task to backlog
        B->>S: Task available
        S-)Worker: Assign task (least loaded)
        Worker->>Worker: Check capacity (weight)
        alt Воркер доступен
            Worker->>+F: Encode video/audio
            loop Каждые 2 секунды
                F-)Worker: Progress
                Worker->>+C: UpdateProgress(delta)
                C--)-Worker: OK
            end
            F--)-Worker: Encoded files
            alt Видео задача
                Worker->>+F: Generate poster (if needed)
                F--)-Worker: Poster file
                Worker->>+HTTP: Upload chunks (parallel)
                HTTP--)-Worker: 200 OK
                Worker->>+HTTP: Upload poster
                HTTP--)-Worker: 200 OK
            else Аудио задача
                Worker->>+HTTP: Upload audio
                HTTP--)-Worker: 200 OK
            end
            Worker->>+C: FinishTask(success)
            C--)-Worker: OK
        else Воркер занят
            S->>S: Wait 15ms, retry
        end
    else Нет задач
        W->>W: Sleep 3s
    else Ошибка
        W->>+C: FinishTask(error)
        C--)-W: OK
    end
```
## 3. Блок-схема алгоритма работы воркера
```mermaid
flowchart TD
    Start([Получение задачи]) --> CheckType{Тип задачи?}
    
    CheckType -->|Видео| VideoStart[Видео обработка]
    CheckType -->|Аудио| AudioStart[Аудио обработка]
    
    VideoStart --> CreateVideoDir[Создать директорию<br/>WorkDir/taskID/assets/part]
    CreateVideoDir --> EncodeVideo[FFmpeg: Кодирование<br/>в несколько качеств]
    EncodeVideo --> CheckPoster{Нужен постер?}
    CheckPoster -->|Да| GeneratePoster[FFmpeg: Генерация<br/>постера]
    CheckPoster -->|Нет| UploadChunks
    GeneratePoster --> UploadChunks[Параллельная загрузка<br/>всех качеств]
    UploadChunks --> UploadPoster{Постер создан?}
    UploadPoster -->|Да| UploadPosterFile[Загрузка постера]
    UploadPoster -->|Нет| CleanupVideo
    UploadPosterFile --> CleanupVideo[Очистка временных<br/>файлов]
    CleanupVideo --> FinishSuccess[FinishTask<br/>успех]
    
    AudioStart --> CreateAudioDir[Создать директорию<br/>WorkDir/taskID/assets/part]
    CreateAudioDir --> EncodeAudio[FFmpeg: Кодирование<br/>аудио]
    EncodeAudio --> UploadAudio[Загрузка аудио файла]
    UploadAudio --> CleanupAudio[Очистка временных<br/>файлов]
    CleanupAudio --> FinishSuccess
    
    FinishSuccess --> End([Завершение])
    
    EncodeVideo -.->|Прогресс каждые 2с| UpdateProgress[UpdateProgress<br/>к Composer]
    UpdateProgress -.-> EncodeVideo
    
    style VideoStart fill:#e1f5ff
    style AudioStart fill:#ffe1f5
    style FinishSuccess fill:#e1ffe1
    style UpdateProgress fill:#fff4e1
```
## 4. Диаграмма состояний задачи в Encoder
```mermaid
stateDiagram-v2
    [*] --> ПолучениеЗадачи: Watcher запущен
    
    ПолучениеЗадачи --> ЗагрузкаФайла: GetTask успешно
    ПолучениеЗадачи --> Ожидание: NO_TASKS
    ПолучениеЗадачи --> Пропуск: SKIP_TASK
    ПолучениеЗадачи --> Ошибка: Ошибка запроса
    
    Ожидание --> ПолучениеЗадачи: Через 3 секунды
    Пропуск --> ПолучениеЗадачи: Немедленно
    
    ЗагрузкаФайла --> ВОчереди: Prefetch успешно
    ЗагрузкаФайла --> Ошибка: Ошибка загрузки
    
    ВОчереди --> Обработка: Scheduler назначил воркер
    ВОчереди --> ВОчереди: Воркеры заняты (wait 15ms)
    
    Обработка --> Кодирование: Воркер принял задачу
    Обработка --> ВОчереди: Воркер перегружен
    
    Кодирование --> ЗагрузкаРезультатов: Кодирование завершено
    Кодирование --> Ошибка: Ошибка кодирования
    
    ЗагрузкаРезультатов --> Очистка: Загрузка успешна
    ЗагрузкаРезультатов --> Ошибка: Ошибка загрузки
    
    Очистка --> Завершено: Временные файлы удалены
    Ошибка --> Завершено: FinishTask с ошибкой
    
    Завершено --> [*]
    
    note right of Кодирование
        Отправка прогресса
        каждые 2 секунды
    end note
```
## 5. Диаграмма распределения нагрузки
```mermaid
graph LR
    subgraph "Scheduler Logic"
        Task[Новая задача] --> Sort[Сортировка воркеров<br/>по InProgress]
        Sort --> Check1{Worker 1<br/>доступен?}
        Check1 -->|Да| Assign1[Назначить Worker 1]
        Check1 -->|Нет| Check2{Worker 2<br/>доступен?}
        Check2 -->|Да| Assign2[Назначить Worker 2]
        Check2 -->|Нет| CheckN{Worker N<br/>доступен?}
        CheckN -->|Да| AssignN[Назначить Worker N]
        CheckN -->|Нет| Wait[Подождать 15ms]
        Wait --> Sort
    end
    
    subgraph "Worker Capacity"
        W1[Worker 1<br/>InProgress: 0<br/>Weight: 0]
        W2[Worker 2<br/>InProgress: 1<br/>Weight: 60]
        W3[Worker N<br/>InProgress: 0<br/>Weight: 0]
    end
    
    Assign1 --> W1
    Assign2 --> W2
    AssignN --> W3
    
    style Task fill:#e1f5ff
    style Sort fill:#fff4e1
    style Wait fill:#ffe1f5
```
## 6. Диаграмма жизненного цикла Encoder сервиса
```mermaid
graph TB
    Start([Запуск сервиса]) --> Init[Инициализация]
    
    Init --> CheckFFmpeg[Проверка версии FFmpeg]
    CheckFFmpeg --> ConnectComposer[Подключение к Composer<br/>gRPC клиент]
    ConnectComposer --> CreateWorkDir[Создание рабочей<br/>директории]
    CreateWorkDir --> ResetTasks[Сброс незавершенных<br/>задач]
    ResetTasks --> CreateWorkers[Создание пула воркеров<br/>N = CPUQuota]
    CreateWorkers --> StartWatchers[Запуск Watchers<br/>N горутин]
    StartWatchers --> StartScheduler[Запуск Scheduler]
    StartScheduler --> StartCleanup[Запуск Cleanup<br/>фоновая горутина]
    StartCleanup --> Running[Сервис работает]
    
    Running --> WatchLoop[Watchers запрашивают<br/>задачи]
    Running --> ScheduleLoop[Scheduler распределяет<br/>задачи]
    Running --> CleanupLoop[Cleanup очищает<br/>старые файлы]
    
    WatchLoop --> WatchLoop
    ScheduleLoop --> ScheduleLoop
    CleanupLoop --> CleanupLoop
    
    Running --> Signal{Получен сигнал<br/>SIGINT/SIGTERM?}
    Signal -->|Нет| Running
    Signal -->|Да| Shutdown[Завершение работы]
    Shutdown --> End([Остановка])
    
    style Start fill:#e1ffe1
    style Running fill:#e1f5ff
    style Shutdown fill:#ffe1f5
    style End fill:#ffe1f5
```
## 7. Диаграмма обработки видео задачи
```mermaid
flowchart TD
    Start([Видео задача]) --> Validate[Проверка загрузки<br/>воркера]
    Validate -->|Перегружен| Reject[Отклонить задачу]
    Validate -->|Доступен| Accept[Принять задачу]
    
    Accept --> CreateDir[Создать директорию<br/>assets/part]
    CreateDir --> Encode[FFmpeg EncodeCPU<br/>Кодирование в качества]
    
    Encode --> ProgressLoop{Прогресс<br/>доступен?}
    ProgressLoop -->|Да| SendProgress[UpdateProgress<br/>к Composer]
    ProgressLoop -->|Нет| CheckEncode
    SendProgress --> CheckEncode{Кодирование<br/>завершено?}
    CheckEncode -->|Нет| ProgressLoop
    CheckEncode -->|Да| CheckPoster{Нужен<br/>постер?}
    
    CheckPoster -->|Да| SelectQuality[Выбрать максимальное<br/>качество]
    SelectQuality --> GeneratePoster[FFmpeg PosterThumbCPU]
    GeneratePoster --> UploadChunks
    CheckPoster -->|Нет| UploadChunks
    
    UploadChunks --> UploadParallel[Параллельная загрузка<br/>всех качеств<br/>HTTP POST]
    UploadParallel --> CheckPosterFile{Постер<br/>создан?}
    CheckPosterFile -->|Да| UploadPoster[Загрузка постера<br/>HTTP POST]
    CheckPosterFile -->|Нет| Cleanup
    UploadPoster --> Cleanup
    
    Cleanup[Удаление временных<br/>файлов] --> Finish[FinishTask<br/>успех]
    Reject --> End([Конец])
    Finish --> End
    
    style Start fill:#e1f5ff
    style Encode fill:#fff4e1
    style UploadParallel fill:#ffe1f5
    style Finish fill:#e1ffe1
```
## 8. Диаграмма обработки аудио задачи
```mermaid
flowchart TD
    Start([Аудио задача]) --> Validate[Проверка загрузки<br/>воркера]
    Validate -->|Перегружен| Reject[Отклонить задачу]
    Validate -->|Доступен| Accept[Принять задачу]
    
    Accept --> CreateDir[Создать директорию<br/>assets/part]
    CreateDir --> Encode[FFmpeg EncodeAudio<br/>Кодирование аудио]
    
    Encode --> CheckEncode{Кодирование<br/>завершено?}
    CheckEncode -->|Ошибка| Error[Ошибка кодирования]
    CheckEncode -->|Успех| UploadAudio
    
    UploadAudio[Загрузка аудио файла<br/>HTTP POST] --> CheckUpload{Загрузка<br/>успешна?}
    CheckUpload -->|Ошибка| Error
    CheckUpload -->|Успех| Cleanup
    
    Cleanup[Удаление временных<br/>файлов] --> Finish[FinishTask<br/>успех]
    Error --> FinishError[FinishTask<br/>с ошибкой]
    Reject --> End([Конец])
    Finish --> End
    FinishError --> End
    
    style Start fill:#e1f5ff
    style Encode fill:#fff4e1
    style UploadAudio fill:#ffe1f5
    style Finish fill:#e1ffe1
    style Error fill:#ffcccc
```
