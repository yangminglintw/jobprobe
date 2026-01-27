# JProbe 架構文件

本文件為工程師提供技術深入說明，適用於維護或擴展 JProbe 的開發人員。

---

## 1. 概述

JProbe 是一個統一的健康檢查 CLI 工具，支援 Rundeck 作業和 HTTP 端點。架構採用 Provider 設計模式，可輕鬆擴展以支援其他作業編排系統。

```
┌─────────────────────────────────────────────────────────────────────┐
│                            CLI 層                                    │
│                    (Cobra 指令: run, list, version)                  │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          設定層                                      │
│              (YAML 載入、驗證、環境變數展開)                          │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         執行引擎                                     │
│              (作業過濾、循序執行)                                     │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        Provider 層                                   │
│  ┌───────────────┐  ┌───────────────┐  ┌───────────────┐           │
│  │    Rundeck    │  │     HTTP      │  │   (未來)      │           │
│  │   Provider    │  │   Provider    │  │   Providers   │           │
│  └───────────────┘  └───────────────┘  └───────────────┘           │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          輸出層                                      │
│                    (Console、JSON Writers)                          │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 2. 技術堆疊

| 類別 | 選擇 | 原因 |
|------|------|------|
| 語言 | Go 1.23 | 單一執行檔、跨平台、DevOps 生態系統 |
| CLI | Cobra | Go 標準、自動產生說明 |
| 設定 | YAML + Viper | 人類可讀、版本控制友善 |
| JSON Path | ohler55/ojg | 輕量、無額外依賴 |
| 顏色 | fatih/color | 跨平台終端機顏色 |

---

## 3. 套件結構

```
jobprobe/
├── cmd/                       # CLI 指令 (Cobra)
│   ├── root.go               # 根指令、全域旗標
│   ├── run.go                # Run 指令 - 執行作業
│   ├── list.go               # List 指令 - 顯示作業/環境
│   └── version.go            # Version 指令
├── internal/
│   ├── config/               # 設定載入與驗證
│   │   ├── config.go         # 核心型別 (Config, Job, Environment)
│   │   ├── loader.go         # YAML 檔案載入、環境變數展開
│   │   ├── validator.go      # 設定驗證
│   │   └── environment.go    # 環境相關輔助函式
│   ├── providers/            # Provider 註冊模式
│   │   ├── provider.go       # Provider 介面、Status、Result
│   │   ├── registry.go       # Provider 註冊表
│   │   ├── http/             # HTTP provider
│   │   │   ├── http.go       # Provider 實作
│   │   │   └── client.go     # HTTP 客戶端封裝
│   │   └── rundeck/          # Rundeck provider
│   │       ├── rundeck.go    # Provider 實作
│   │       ├── client.go     # Rundeck API 客戶端
│   │       └── types.go      # Rundeck 專用型別
│   ├── runner/               # 作業執行編排
│   │   ├── runner.go         # 主要執行器、作業過濾
│   │   ├── executor.go       # 作業執行邏輯
│   │   └── result.go         # 執行結果彙整
│   └── output/               # 輸出格式化
│       ├── output.go         # Writer 介面、ProgressAdapter
│       ├── console.go        # Console 輸出（含顏色）
│       └── json.go           # JSON 輸出
├── configs/                  # 範例設定
├── test/                     # 測試資源
│   └── mock-api/             # 整合測試用模擬 API
└── main.go                   # 程式進入點
```

---

## 4. 核心介面

### 4.1 Provider 介面

Provider 介面是核心擴展點。每種作業類型（HTTP、Rundeck 等）都實作此介面。

```go
// Provider 定義作業執行 provider 的介面。
type Provider interface {
    // Name 回傳 provider 名稱（例如 "http"、"rundeck"）。
    Name() string

    // Execute 執行作業並回傳結果。
    Execute(ctx context.Context, job config.Job, env config.Environment) (*Result, error)
}
```

**位置**: `internal/providers/provider.go:57-63`

### 4.2 Status 列舉

```go
type Status string

const (
    StatusPending   Status = "pending"
    StatusRunning   Status = "running"
    StatusSucceeded Status = "succeeded"
    StatusFailed    Status = "failed"
    StatusAborted   Status = "aborted"
    StatusTimedOut  Status = "timed_out"
)

func (s Status) IsTerminal() bool  // 回傳是否為終態
func (s Status) IsSuccess() bool   // 僅 succeeded 回傳 true
```

**位置**: `internal/providers/provider.go:12-36`

### 4.3 Result 型別

```go
type Result struct {
    JobName     string                 `json:"name"`
    Environment string                 `json:"environment"`
    Type        string                 `json:"type"`
    Status      Status                 `json:"status"`
    StartedAt   time.Time              `json:"started_at"`
    FinishedAt  time.Time              `json:"finished_at"`
    Duration    time.Duration          `json:"duration_ms"`
    Error       string                 `json:"error,omitempty"`
    Details     map[string]interface{} `json:"details,omitempty"`
}
```

**位置**: `internal/providers/provider.go:39-49`

### 4.4 Writer 介面

Writer 介面抽象化輸出格式。實作包含 Console（含顏色）和 JSON writers。

```go
type Writer interface {
    WriteHeader(version string)
    WriteConfigSummary(envCount, jobCount int)
    WriteJobStart(index, total int, job config.Job)
    WriteJobProgress(jobName string, status providers.Status, message string)
    WriteJobComplete(index, total int, result *providers.Result)
    WriteResult(result *runner.RunResult)
}
```

**位置**: `internal/output/output.go:11-29`

### 4.5 Registry 模式

```go
type Registry struct {
    mu        sync.RWMutex
    providers map[string]Provider
}

func (r *Registry) Register(provider Provider)
func (r *Registry) Get(name string) (Provider, error)
func (r *Registry) List() []string

var DefaultRegistry = NewRegistry()
```

**位置**: `internal/providers/registry.go:8-63`

### 4.6 ProgressHandler 介面

```go
type ProgressHandler interface {
    OnJobStart(index, total int, job config.Job)
    OnJobProgress(jobName string, status providers.Status, message string)
    OnJobComplete(index, total int, result *providers.Result)
}
```

**位置**: `internal/runner/runner.go:20-24`

---

## 5. 設計模式

| 模式 | 位置 | 用途 |
|------|------|------|
| Registry | `providers/registry.go` | Provider 外掛架構 |
| Adapter | `output/output.go:ProgressAdapter` | 橋接 Writer → ProgressHandler |
| Strategy | `cmd/run.go` | 輸出格式選擇 (console/json) |
| Factory | `NewJSONWriter()`, `NewConsoleWriter()` | Writer 實例化 |
| Template Method | `runner/runner.go:Run()` | 共用執行流程與掛鉤點 |

---

## 6. 執行流程

```
main()
  └─► cmd.Execute()
        └─► runCmd.RunE()
              │
              ├─► 1. config.Load()
              │     • 從目錄載入 YAML 檔案
              │     • 展開環境變數
              │     • 驗證設定
              │
              ├─► 2. 選擇 Writer
              │     • Console（預設）或 JSON
              │     • 依據 --output 旗標
              │
              ├─► 3. runner.NewRunner()
              │     • 使用設定初始化
              │     • 設置 executor 使用 DefaultRegistry
              │
              ├─► 4. runner.Run()
              │     │
              │     ├─► filterJobs()
              │     │     • 依 --name、--tags、--env 過濾
              │     │
              │     └─► 對每個作業:
              │           ├─► progressHandler.OnJobStart()
              │           ├─► executor.Execute()
              │           │     └─► provider.Execute()
              │           ├─► result.AddResult()
              │           └─► progressHandler.OnJobComplete()
              │
              ├─► 5. writer.WriteResult()
              │     • 寫入摘要
              │     • 顯示通過/失敗數量
              │
              └─► 6. 回傳結束碼
                    • 0: 全部通過
                    • 1: 一個以上失敗
                    • 2: 設定錯誤
                    • 3: 執行時錯誤
```

---

## 7. Provider 實作

### 7.1 HTTP Provider

**位置**: `internal/providers/http/`

**執行流程**:
1. 從環境基礎 URL + 作業路徑建立請求 URL
2. 設定方法 (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS)
3. 新增標頭（來自環境 + 作業）
4. 新增認證 (bearer, basic, api_key)
5. 設定請求主體（如有）
6. 執行請求（含逾時）
7. 執行斷言

**斷言**:
- `status_code`: 預期 HTTP 狀態碼
- `json`: 回應主體的 JSONPath 斷言
- `max_duration`: 最大可接受回應時間

**認證類型**:
| 類型 | 標頭 |
|------|------|
| bearer | `Authorization: Bearer <token>` |
| basic | `Authorization: Basic <base64(user:pass)>` |
| api_key | 可設定的標頭名稱 |

### 7.2 Rundeck Provider

**位置**: `internal/providers/rundeck/`

**執行流程**:
1. 連接 Rundeck API (版本 41+)
2. 以選項觸發作業執行
3. 以可設定間隔輪詢執行狀態
4. 等待終態或逾時
5. 執行斷言

**輪詢**:
- 預設間隔：10 秒（可設定）
- 透過 ProgressCallback 回報進度
- 遵守 context 取消

**狀態對應**:
| Rundeck 狀態 | JProbe 狀態 |
|--------------|-------------|
| succeeded | StatusSucceeded |
| failed | StatusFailed |
| aborted | StatusAborted |
| timedout | StatusTimedOut |
| running | StatusRunning |

**斷言**:
- `status`: 預期最終狀態（通常為 "succeeded"）
- `max_duration`: 最大可接受執行時間

---

## 8. 設定結構

### 8.1 Config 結構

```go
type Config struct {
    Defaults     Defaults               `yaml:"defaults"`
    Output       OutputConfig           `yaml:"output"`
    Environments map[string]Environment `yaml:"environments"`
    Jobs         []Job                  `yaml:"jobs"`
}
```

**位置**: `internal/config/config.go:7-12`

### 8.2 檔案組織

```
configs/
├── config.yaml          # 預設值和輸出設定
├── environments.yaml    # 目標環境
└── jobs/
    ├── http-checks.yaml # HTTP 健康檢查
    └── rundeck-jobs.yaml # Rundeck 作業
```

載入器會合併目錄中所有 YAML 檔案。

### 8.3 環境類型

| 類型 | 欄位 |
|------|------|
| http | url, headers, auth (bearer/basic/api_key) |
| rundeck | url, api_version, auth (token) |

### 8.4 作業類型

| 類型 | 欄位 |
|------|------|
| http | method, path, headers, body, assertions |
| rundeck | job_id, project, options, timeout, poll_interval, assertions |

### 8.5 驗證規則

- `defaults.timeout` > 0（如有設定）
- `environment.type` 必須為 "http" 或 "rundeck"
- `job.name` 必須唯一
- `job.environment` 必須參照有效環境
- `job.type` 必須與環境類型相符

---

## 9. 認證架構

所有憑證值都支援 `${ENV_VAR}` 展開以確保設定安全。

| 認證類型 | Provider | 實作 |
|----------|----------|------|
| bearer | HTTP | `Authorization: Bearer <token>` 標頭 |
| basic | HTTP | `Authorization: Basic <base64>` 標頭 |
| api_key | HTTP/Rundeck | 自訂標頭與 API 金鑰值 |
| token | Rundeck | `X-Rundeck-Auth-Token` 標頭 |

**安全注意事項**:
- 憑證應始終使用環境變數
- 切勿將 token 儲存在設定檔中
- 憑證不會記錄在輸出中

---

## 10. 錯誤處理

### 10.1 狀態流程

```
StatusPending → StatusRunning → StatusSucceeded
                             → StatusFailed
                             → StatusAborted
                             → StatusTimedOut
```

### 10.2 終態

```go
func (s Status) IsTerminal() bool {
    switch s {
    case StatusSucceeded, StatusFailed, StatusAborted, StatusTimedOut:
        return true
    }
    return false
}
```

### 10.3 錯誤傳播

1. Provider 錯誤 → Result.Error 欄位
2. 結果彙整至 RunResult
3. 結束碼由摘要決定：
   - 全部通過 → 0
   - 任一失敗 → 1
   - 設定錯誤 → 2
   - 執行時錯誤 → 3

---

## 11. 擴展 JProbe

### 11.1 新增 Provider

1. 在 `internal/providers/{name}/` 下建立套件
2. 實作 Provider 介面
3. 在 `init()` 中註冊：

```go
package myprovider

import "github.com/user/jobprobe/internal/providers"

func init() {
    providers.Register(New())
}

type MyProvider struct{}

func New() *MyProvider {
    return &MyProvider{}
}

func (p *MyProvider) Name() string {
    return "myprovider"
}

func (p *MyProvider) Execute(ctx context.Context, job config.Job, env config.Environment) (*providers.Result, error) {
    // 實作
}
```

4. 在 `main.go` 中匯入套件以執行 init()：
```go
import _ "github.com/user/jobprobe/internal/providers/myprovider"
```

### 11.2 新增輸出格式

1. 在 `internal/output/` 下建立檔案
2. 實作 Writer 介面
3. 新增工廠函式
4. 更新 `cmd/run.go` 以支援新格式旗標：

```go
case "myformat":
    writer = output.NewMyFormatWriter(os.Stdout)
```

### 11.3 新增斷言類型

1. 將欄位新增至 `config.Assertions` 結構
2. 更新 provider 的斷言檢查邏輯
3. 在 `internal/config/validator.go` 中新增驗證

---

## 12. 依賴套件

| 套件 | 版本 | 用途 |
|------|------|------|
| spf13/cobra | v1.10.2 | CLI 框架 |
| spf13/viper | latest | 設定管理 |
| gopkg.in/yaml.v3 | v3.0.1 | YAML 解析 |
| ohler55/ojg | latest | JSON path 斷言 |
| fatih/color | latest | 終端機顏色 |

---

## 13. 建置系統

### 13.1 Makefile 目標

```bash
make build          # 建置當前平台的執行檔
make build-all      # 建置所有平台 (linux, darwin, windows)
make test           # 執行單元測試
make test-coverage  # 執行測試並產生覆蓋率報告
make lint           # 執行 golangci-lint
make docker-build   # 建置 Docker 映像檔
make clean          # 移除建置產物
```

### 13.2 版本注入

版本資訊在建置時透過 ldflags 注入：

```bash
go build -ldflags "-X cmd.Version=0.1.0 -X cmd.Commit=$(git rev-parse HEAD) -X cmd.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

### 13.3 Docker 建置

使用 Alpine 基礎的多階段建置：

```dockerfile
FROM golang:1.23-alpine AS builder
# 建置靜態執行檔

FROM alpine:3.19
# 複製執行檔、加入 ca-certificates
```

---

## 14. 測試

### 14.1 單元測試

```bash
go test ./...
go test -cover ./...
```

### 14.2 整合測試

位於 `test/` 目錄，包含模擬 API 伺服器。

```bash
# 啟動模擬 API
go run test/mock-api/main.go &

# 執行整合測試
go test -tags=integration ./...
```

### 14.3 測試覆蓋率

目標：核心套件 80%+ 覆蓋率。

---

## 15. 未來考量

### 計畫中的增強功能

1. **平行執行**
   - Goroutine 池用於並行作業執行
   - 可設定並行數限制
   - 執行緒安全的結果彙整

2. **重試機制**
   - 每個作業可設定重試次數
   - 指數退避
   - 針對特定錯誤類型重試

3. **額外 Providers**
   - Jenkins：作業觸發和狀態輪詢
   - Airflow：DAG 觸發和任務監控
   - Kubernetes：Job/CronJob 執行

4. **指標輸出**
   - Prometheus 格式用於監控整合
   - OpenMetrics 相容

5. **通知 Webhooks**
   - Slack、Teams、Discord 整合
   - 可設定通知規則
