# Log Detect Backend 文件總覽

本目錄是 `log-detect-backend` 的正式交付文件入口，整理後分為「核心規格」、「操作手冊」、「補充設計」與「歷史文件」四類。

## 核心規格

核心技術規格集中在 [`specs/`](./specs/)：

- [規格文件導覽](./specs/README.md)
- [01-系統概述](./specs/01-%E7%B3%BB%E7%B5%B1%E6%A6%82%E8%BF%B0.md)
- [02-架構設計](./specs/02-%E6%9E%B6%E6%A7%8B%E8%A8%AD%E8%A8%88.md)
- [03-API規格](./specs/03-API%E8%A6%8F%E6%A0%BC.md)
- [04-資料庫設計](./specs/04-%E8%B3%87%E6%96%99%E5%BA%AB%E8%A8%AD%E8%A8%88.md)
- [05-安全與部署](./specs/05-%E5%AE%89%E5%85%A8%E8%88%87%E9%83%A8%E7%BD%B2.md)
- [快速參考](./specs/%E5%BF%AB%E9%80%9F%E5%8F%83%E8%80%83.md)

## 操作手冊

YML 模式的交付與測試說明集中在 [`manuals/`](./manuals/)：

- [YML 模式操作與設定手冊](./manuals/yml-mode-operation-manual.md)
- [YML 模式測試 Checklist](./manuals/yml-mode-test-checklist.md)

## API 與 Swagger

- [OpenAPI 規格](./openapi.yml)
- [Swagger JSON](./swagger.json)
- [Swagger YAML](./swagger.yaml)

互動式文件：

- `http://<server>:8006/swagger/index.html`

## 補充設計文件

這些文件屬於功能設計補充，適合需要追蹤設計背景與演進的開發/維運人員閱讀：

- [Feature Toggles 與 YML 同步](./features/feature-toggles-yml-sync.md)
- [HA Group 設計](./features/ha-group.md)
- [YML Hot Reload / Disabled Devices / Alert Policy 設計](./features/yml-hot-reload-disabled-devices-alert-policy-design.md)

## 歷史文件

以下文件保留作為歷史參考，不建議當成最新規格的唯一依據：

- [Auth System Legacy](./legacy/auth-system-legacy.md)
- [General Troubleshooting Legacy](./legacy/general-troubleshooting-legacy.md)
- [History Data Management Legacy](./legacy/history-data-management-legacy.md)

## 閱讀建議

不同角色可優先閱讀：

- 架構/交付：`specs/`
- 維運/測試：`manuals/` + `specs/05-安全與部署.md`
- API 整合：`specs/03-API規格.md` + `openapi.yml`
- 問題排查：先看 `specs/`，再視需要看 `features/` 與 `legacy/`
