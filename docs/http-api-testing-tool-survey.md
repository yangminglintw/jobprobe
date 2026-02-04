# HTTP 測試工具調查報告

## 一、場景需求（Scenario）

### 1.1 概述

我們需要測試多種類型的 API，包含同步與非同步 API。主要挑戰在於非同步 API 需要多步驟呼叫（觸發→輪詢→驗證）。

### 1.2 測試對象

| 對象 | 說明 | 特性 |
|------|------|------|
| Rundeck Job API | 觸發和監控排程任務 | 非同步執行 |
| FastAPI 應用 | 內部開發的 API 服務 | 同步/非同步 |
| 其他 HTTP API | 健康檢查、可用性監控 | 同步 |

### 1.3 非同步 API 測試情境

部分 API 採用非同步模式，測試流程涉及**多個 API 呼叫**：

```
┌─────────────────────────────────────────────────┐
│         非同步 API 測試流程                       │
├─────────────────────────────────────────────────┤
│                                                 │
│  步驟 1：觸發操作（API 呼叫 A）                   │
│  POST /api/trigger                              │
│  → 回應：task_id, status: "pending"             │
│                                                 │
│  步驟 2：等待執行                                │
│  （背景非同步執行中）                             │
│                                                 │
│  步驟 3：檢查狀態（API 呼叫 B）                   │
│  GET /api/status/{task_id}                      │
│  → 回應：status: "running" / "completed" / ...  │
│                                                 │
│  步驟 4：取得結果（API 呼叫 C）                   │
│  GET /api/result/{task_id}                      │
│  → 回應：實際執行結果                            │
│                                                 │
└─────────────────────────────────────────────────┘
```

**適用此模式的 API：**
- Rundeck Job 執行
- FastAPI 背景任務
- 任何採用「觸發→輪詢→取結果」模式的 API

**核心挑戰：**
- API A 和 API B/C 是不同的端點
- 需要從 API A 的回應擷取 ID（如 `task_id`、`execution_id`）
- 需要傳遞此 ID 給後續 API
- 需要重複呼叫 API B 直到狀態變為終態（如 completed、failed、cancelled）

---

## 二、評估標準（Standards）

### 2.1 概述

我們採用 4 個核心指標評估工具，偏好純配置方式、不撰寫程式碼。

| 指標 | 權重 | 說明 |
|------|:----:|------|
| 非同步處理能力 | 30% | 多步驟 API 工作流程 |
| 內容驗證能力 | 25% | 回應內容檢查 |
| 維護成本 | 30% | 偏好配置，避免程式碼 |
| 整合友善度 | 15% | CLI、退出碼（K8s CronJob 執行） |

### 2.2 指標 1：非同步處理能力（權重：30%）

**定義**：工具處理「多步驟 API 工作流程」的能力

| 等級 | 定義 | 具體行為 |
|:----:|------|----------|
| 1 | 無法處理 | 只能發送單一請求，無法串接 |
| 2 | 需外部腳本 | 需用 shell script 包裝多個請求 |
| 3 | 內建 retry | 可對單一 API 重試，但無法跨 API 傳遞資料 |
| 4 | 變數擷取 + retry | 可從 API A 擷取值，傳給 API B，並重試 |
| 5 | 完整工作流程 | 可定義完整流程：觸發→等待→驗證，支援條件判斷 |

**非同步 API 測試的具體需求：**

```
工具需要能夠：
1. 呼叫觸發 API
2. 從回應中擷取 task_id / execution_id
3. 用此 ID 呼叫狀態 API
4. 重複步驟 3 直到狀態為終態（如 completed、failed、cancelled）
5. 驗證最終狀態是否正確
```

### 2.3 指標 2：內容驗證能力（權重：25%）

| 等級 | 定義 | 說明 |
|:----:|------|------|
| 1 | 無驗證 | 僅檢查狀態碼 |
| 2 | 字串匹配 | 包含/不包含特定字串 |
| 3 | 正則表達式 | 支援 regex 匹配 |
| 4 | JSON 路徑 | 支援 jsonpath 取值比對 |
| 5 | 完整斷言 | 多條件、類型檢查、Schema 驗證 |

### 2.4 指標 3：維護成本（權重：30%）

**定義**：偏好純配置方式，不撰寫程式碼，直接使用開源工具

| 等級 | 定義 | 說明 |
|:----:|------|------|
| 1 | 需大量程式碼 | 必須撰寫自定義程式邏輯 |
| 2 | 部分程式碼 | 核心邏輯需寫程式，配置輔助 |
| 3 | 混合模式 | 簡單場景用配置，複雜需程式 |
| 4 | 主要配置 | 大多數場景只需配置檔 |
| 5 | 純配置 | 完全透過配置定義，不需程式碼 |

### 2.5 指標 4：整合友善度（權重：15%）

| 等級 | 定義 | 說明 |
|:----:|------|------|
| 1 | 僅 GUI | 只能手動執行 |
| 2 | CLI 可用 | 可命令列執行 |
| 3 | 結構化輸出 | 支援 JSON 輸出 |
| 4 | CI/CD 整合 | 支援 JUnit 報告、退出碼 |
| 5 | 完整生態系 | API、Webhook、多種格式支援 |

### 2.6 評估公式

```
總分 = (非同步處理 × 0.30) + (內容驗證 × 0.25) +
       (維護成本 × 0.30) + (整合友善 × 0.15)
```

---

## 三、工具調查（Tool Survey）

### 3.1 概述

調查了 8 個開源工具，評估其在非同步 API 測試場景的適用性。重點考量：多步驟 API 支援、純配置方式、活躍維護狀態。

### 3.2 調查工具清單

| 分類 | 工具名稱 | 配置/程式碼 | 授權 | 商用 | GitHub Stars | 活躍度 |
|------|----------|:-----------:|------|:----:|-------------:|:------:|
| API 測試 | Hurl | 配置（.hurl） | Apache 2.0 | 可 | ~15k | 活躍 |
| API 測試 | k6 | 程式碼（JS） | AGPL-3.0 | 可 | ~26k | 活躍 |
| API 測試 | Step CI | 配置（YAML） | MPL-2.0 | 可 | ~2k | 活躍 |
| API 測試 | Tavern | 配置（YAML） | MIT | 可 | ~1k | 維護中 |
| API 測試 | Bruno | 配置（JSON+JS） | MIT | 可 | ~29k | 活躍 |
| HTTP 探測 | httpx | 配置（參數） | MIT | 可 | ~8k | 活躍 |
| 監控 | Blackbox Exporter | 配置（YAML） | Apache 2.0 | 可 | ~5k | 活躍 |
| 通用 | curl + shell | 程式碼（Shell） | MIT | 可 | N/A | N/A |

**活躍度說明：**

| 狀態 | 說明 |
|------|------|
| 活躍 | 近期有頻繁更新，社群活躍 |
| 維護中 | 有維護但更新較少 |

**配置方式說明：**

| 類型 | 說明 | 範例工具 |
|------|------|----------|
| 配置 | 使用宣告式格式（YAML/JSON/DSL） | Hurl、Step CI |
| 配置 + 腳本 | 宣告式為主，腳本輔助 | Bruno、Tavern |
| 程式碼 | 必須撰寫程式邏輯 | k6、Shell |

**授權說明：**

| 授權 | 商業使用 | 注意事項 |
|------|:--------:|----------|
| MIT | 可 | 無限制 |
| Apache 2.0 | 可 | 需保留版權聲明 |
| MPL-2.0 | 可 | 修改檔案需開源 |
| AGPL-3.0 | 可 | 僅修改並散佈 k6 本身時需開源；執行測試不受影響 |

### 3.3 工具分類

依功能分為三類：

| 類型 | 工具 | 多步驟 API | 適用場景 |
|------|------|:----------:|----------|
| API 測試 | Hurl、Step CI、k6、Tavern、Bruno | 支援 | 非同步 API 測試 |
| HTTP 探測 | httpx | 不支援 | 存活檢查、批量探測 |
| 監控 | Blackbox Exporter | 不支援 | 持續監控 |

### 3.4 多步驟 API 呼叫能力比較

| 工具 | 多步驟 API | 變數傳遞 | 條件重試 | 配置方式 |
|------|:----------:|:--------:|:--------:|----------|
| **Hurl** | 支援 | 支援 | 支援 | .hurl 純文字檔 |
| **Step CI** | 支援 | 支援 | 支援 | YAML workflow |
| **k6** | 支援 | 支援 | 支援 | JavaScript 腳本 |
| **Tavern** | 支援 | 支援 | 支援 | YAML stages |
| **Bruno** | 支援 | 支援 | 支援 | JSON 配置 + JS script |
| httpx | 不支援 | 不支援 | 不支援 | CLI 參數 |
| Blackbox | 不支援 | 不支援 | 不支援 | YAML（單一端點） |
| curl | 不支援 | 不支援 | 不支援 | 需 shell 包裝 |

### 3.5 認證機制支援比較

以下僅列出支援多步驟 API 的工具：

| 工具 | API Key | Bearer Token | OAuth 2.0 | 環境變數 |
|------|:-------:|:------------:|:---------:|:--------:|
| Hurl | ✓ | ✓ | 需多步驟 | ✓ |
| Step CI | ✓ | ✓ | ✓（內建） | ✓ |
| k6 | ✓ | ✓ | ✓（程式碼）| ✓ |
| Tavern | ✓ | ✓ | 需 plugin | ✓ |
| Bruno | ✓ | ✓ | ✓（內建） | ✓ |

### 3.6 綜合評分表

**注意：** 以下評分適用於通用 HTTP 測試場景。若專注於**非同步 API 測試**，請優先考慮「非同步處理」欄位得分 ≥ 4 的工具。

| 工具 | 非同步處理<br>(30%) | 內容驗證<br>(25%) | 維護成本<br>(30%) | 整合友善<br>(15%) | **加權總分** |
|------|:---:|:---:|:---:|:---:|:---:|
| **Hurl** | 4 | 5 | 5 | 4 | **4.55** |
| **Step CI** | 4 | 4 | 5 | 5 | **4.45** |
| **Tavern** | 4 | 4 | 4 | 4 | **4.00** |
| **k6** | 5 | 5 | 2 | 4 | **3.95** |
| **httpx** | 2 | 4 | 5 | 5 | **3.85** |
| **Bruno** | 4 | 4 | 3 | 4 | **3.70** |
| **Blackbox Exporter** | 2 | 3 | 5 | 5 | **3.60** |
| **Shell + curl** | 4 | 3 | 1 | 5 | **3.00** |

### 3.7 工具詳細評估（多步驟 API 工具優先）

#### 3.7.1 Hurl（總分：4.55）

| 資訊 | 內容 |
|------|------|
| 官方網站 | https://hurl.dev |
| GitHub | https://github.com/Orange-OpenSource/hurl |
| 配置方式 | .hurl 純文字檔（宣告式） |

**多步驟 API 範例：**

```hurl
# 步驟 1：觸發任務
POST https://api.example.com/tasks
Content-Type: application/json
{"action": "run"}
HTTP 200
[Captures]
task_id: jsonpath "$.id"

# 步驟 2：輪詢狀態
GET https://api.example.com/tasks/{{task_id}}/status
[Options]
retry: 10
retry-interval: 5000
HTTP 200
[Asserts]
jsonpath "$.status" == "completed"
```

**評分：**

| 指標 | 分數 | 說明 |
|------|:----:|------|
| 非同步處理 | 4 | 支援變數擷取 + retry |
| 內容驗證 | 5 | 完整斷言（JSONPath、regex、Schema） |
| 維護成本 | 5 | 純配置 |
| 整合友善 | 4 | 支援 JUnit、退出碼 |

---

#### 3.7.2 Step CI（總分：4.45）

| 資訊 | 內容 |
|------|------|
| 官方網站 | https://stepci.com |
| GitHub | https://github.com/stepci/stepci |
| 配置方式 | YAML 檔案（宣告式） |

**多步驟 API 範例：**

```yaml
version: "1.1"
name: Async API Test
tests:
  async-workflow:
    steps:
      # 步驟 1：觸發任務
      - name: Trigger Task
        http:
          url: https://api.example.com/tasks
          method: POST
          json:
            action: run
        captures:
          task_id:
            jsonpath: $.id

      # 步驟 2：輪詢狀態
      - name: Check Status
        http:
          url: https://api.example.com/tasks/${{captures.task_id}}/status
        check:
          jsonpath:
            $.status: completed
        options:
          retries: 10
          retryInterval: 5000
```

**評分：**

| 指標 | 分數 | 說明 |
|------|:----:|------|
| 非同步處理 | 4 | 支援變數擷取 + retry |
| 內容驗證 | 4 | JSONPath、regex |
| 維護成本 | 5 | 純配置 |
| 整合友善 | 5 | 完整生態系（CI/CD、報告） |

---

#### 3.7.3 Tavern（總分：4.00）

| 資訊 | 內容 |
|------|------|
| 官方網站 | https://taverntesting.github.io |
| GitHub | https://github.com/taverntesting/tavern |
| 配置方式 | YAML 檔案（需 Python 環境） |

**多步驟 API 範例：**

```yaml
test_name: Async API Test

stages:
  # 步驟 1：觸發任務
  - name: Trigger task
    request:
      url: https://api.example.com/tasks
      method: POST
      json:
        action: run
    response:
      status_code: 200
      save:
        json:
          task_id: id

  # 步驟 2：輪詢狀態（需配合 pytest retry plugin）
  - name: Check status
    request:
      url: https://api.example.com/tasks/{task_id}/status
    response:
      status_code: 200
      json:
        status: completed
```

**評分：**

| 指標 | 分數 | 說明 |
|------|:----:|------|
| 非同步處理 | 4 | 支援變數擷取，retry 需 plugin |
| 內容驗證 | 4 | JSONPath、regex |
| 維護成本 | 4 | 主要配置，需 Python 環境 |
| 整合友善 | 4 | pytest 整合、JUnit 報告 |

---

#### 3.7.4 k6（總分：3.95）

| 資訊 | 內容 |
|------|------|
| 官方網站 | https://k6.io |
| GitHub | https://github.com/grafana/k6 |
| 配置方式 | JavaScript 程式碼 |

**多步驟 API 範例：**

```javascript
import http from 'k6/http';
import { check, sleep } from 'k6';

export default function() {
  // 步驟 1：觸發任務
  let triggerRes = http.post(
    'https://api.example.com/tasks',
    JSON.stringify({ action: 'run' }),
    { headers: { 'Content-Type': 'application/json' } }
  );

  let taskId = triggerRes.json('id');

  // 步驟 2：輪詢狀態
  let status = 'running';
  let retries = 10;

  while (status !== 'completed' && retries > 0) {
    sleep(5);
    let statusRes = http.get(
      `https://api.example.com/tasks/${taskId}/status`
    );
    status = statusRes.json('status');
    retries--;
  }

  check(status, { 'task completed': (s) => s === 'completed' });
}
```

**評分：**

| 指標 | 分數 | 說明 |
|------|:----:|------|
| 非同步處理 | 5 | 完整工作流程（程式碼彈性高） |
| 內容驗證 | 5 | 完整斷言（程式碼可實現任何驗證） |
| 維護成本 | 2 | 需撰寫 JavaScript |
| 整合友善 | 4 | 支援 JUnit、JSON 輸出 |

---

#### 3.7.5 Bruno（總分：3.70）

| 資訊 | 內容 |
|------|------|
| 官方網站 | https://www.usebruno.com |
| GitHub | https://github.com/usebruno/bruno |
| 配置方式 | JSON 配置 + JavaScript script |

**多步驟 API 範例：**

```javascript
// 步驟 1：觸發任務（Post-request script）
bru.setVar("task_id", res.body.id);

// 步驟 2：輪詢狀態
// GET https://api.example.com/tasks/{{task_id}}/status
// 需搭配 bru.runRequest() 實現 retry 邏輯
```

**評分：**

| 指標 | 分數 | 說明 |
|------|:----:|------|
| 非同步處理 | 4 | 支援 `bru.runRequest()` + 變數傳遞 |
| 內容驗證 | 4 | JavaScript assertions |
| 維護成本 | 3 | 混合模式（JSON 配置 + JS script） |
| 整合友善 | 4 | CLI 支援、JUnit reporter |

---

#### 3.7.6 httpx（總分：3.85）

| 資訊 | 內容 |
|------|------|
| 官方網站 | https://projectdiscovery.io |
| GitHub | https://github.com/projectdiscovery/httpx |
| 配置方式 | CLI 參數 |

**多步驟 API 範例：** 不支援

httpx 設計用於單一請求探測，不支援多步驟工作流程。

**適用場景**：批量探測、存活檢查

**評分：**

| 指標 | 分數 | 說明 |
|------|:----:|------|
| 非同步處理 | 2 | 需外部腳本 |
| 內容驗證 | 4 | 支援狀態碼、內容匹配 |
| 維護成本 | 5 | 純 CLI 參數 |
| 整合友善 | 5 | JSON 輸出、退出碼 |

---

#### 3.7.7 Blackbox Exporter（總分：3.60）

| 資訊 | 內容 |
|------|------|
| 官方網站 | https://prometheus.io |
| GitHub | https://github.com/prometheus/blackbox_exporter |
| 配置方式 | YAML 檔案 |

**多步驟 API 範例：** 不支援

Blackbox Exporter 設計用於持續監控單一端點，不支援多步驟工作流程。

**適用場景**：持續監控、健康檢查

**評分：**

| 指標 | 分數 | 說明 |
|------|:----:|------|
| 非同步處理 | 2 | 需外部腳本 |
| 內容驗證 | 3 | 支援 regex |
| 維護成本 | 5 | 純配置 |
| 整合友善 | 5 | Prometheus 整合 |

---

#### 3.7.8 Shell Script + curl（總分：3.00）

| 資訊 | 內容 |
|------|------|
| 說明 | bash/shell 腳本配合 curl |
| 配置方式 | 程式碼（Shell Script） |

**多步驟 API 範例：**

```bash
#!/bin/bash

# 步驟 1：觸發任務
RESPONSE=$(curl -s -X POST https://api.example.com/tasks \
  -H "Content-Type: application/json" \
  -d '{"action": "run"}')

TASK_ID=$(echo $RESPONSE | jq -r '.id')

# 步驟 2：輪詢狀態
RETRIES=10
while [ $RETRIES -gt 0 ]; do
  STATUS=$(curl -s "https://api.example.com/tasks/${TASK_ID}/status" \
    | jq -r '.status')

  if [ "$STATUS" == "completed" ]; then
    echo "Task completed!"
    exit 0
  fi

  sleep 5
  RETRIES=$((RETRIES - 1))
done

echo "Task failed or timeout"
exit 1
```

**評分：**

| 指標 | 分數 | 說明 |
|------|:----:|------|
| 非同步處理 | 4 | 可實現，但需自行撰寫邏輯 |
| 內容驗證 | 3 | 需配合 jq 等工具 |
| 維護成本 | 1 | 需大量程式碼 |
| 整合友善 | 5 | 退出碼、可整合任何環境 |

---

## 四、結論與建議

### 4.1 概述

根據評估結果，**Hurl** 和 **Step CI** 是最適合的工具選擇（純配置、支援多步驟 API）。

### 4.2 依場景選擇工具

| 場景 | 建議工具 | 原因 |
|------|----------|------|
| **非同步 API 測試** | Hurl 或 Step CI | 純配置、支援變數傳遞、retry |
| **快速探測** | httpx | 高速、純 CLI 參數 |
| **持續監控** | Blackbox Exporter | 純配置、Prometheus 整合 |

### 4.3 工具定位

```
┌─────────────────────────────────────────────────┐
│              工具定位                            │
├─────────────────────────────────────────────────┤
│                                                 │
│  非同步 API 測試                                 │
│  ├─ 首選：Hurl（純配置、變數擷取 + retry）        │
│  └─ 替代：Step CI（YAML 配置、CI/CD 友善）       │
│                                                 │
│  端點探測                                        │
│  └─ httpx（批量探測、存活檢查）                   │
│                                                 │
│  持續監控                                        │
│  └─ Blackbox Exporter（Prometheus 整合）        │
│                                                 │
└─────────────────────────────────────────────────┘
```

### 4.4 下一步行動

1. **選擇測試工具**：建議選擇 Hurl（純配置、適合非同步 API）
2. **建立 POC**：針對一個非同步 API 建立 .hurl 測試檔
3. **驗證流程**：確認變數擷取和 retry 機制符合需求

---

*文件產生日期：2026-02-02*
