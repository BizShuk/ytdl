# CLAUDE.md — ytdl 技術脈絡 (Technical Context)

YouTube 下載 CLI。業務定義見 [README.md](README.md)。
Module `github.com/bizshuk/ytdl`，Go 1.26。TUI 採用 `bubbletea` 與 `lipgloss`。

## 結構與 ownership

```text
ytdl/
├── main.go              # flag 解析 + subcommand 派工 + 逐 URL 派工
└── pkg/
    ├── config/
    │   └── config.go    # config 與 data 目錄解析與建立 (~/.config/ytdl)
    ├── download/
    │   ├── quality.go   # MediaType 與 quality tier → yt-dlp format selector
    │   └── download.go  # Request 驗證、runtime 檢查、exec yt-dlp
    └── tui/
        ├── file.go      # 下載檔案讀取、格式化、系統預設開啟
        ├── model.go     # Bubble Tea Model、狀態轉換、事件迴圈
        ├── run.go       # TUI 程式進入點 (alt-screen)
        ├── view.go      # 視圖渲染（Header/Banner/Table/Footer 高度預算）
        └── width.go     # 字符寬度測量與截斷（runewidth 私有實例）
```

`pkg/download` 不認識 flag，`main.go` 不認識 yt-dlp 參數。

## 關鍵決策 (Key Decisions)

- CLI 用 **stdlib `flag`** 而非 cobra：使用者要的是 `-type` / `-qtype`
  單破折號長旗標，pflag 會把 `-type` 當成短旗標叢集解析而失敗。子命令
  如 `config` 與 `m` / `monitor` 透過 `os.Args` 前置派工與 `flag.NewFlagSet` 獨立解析。
- `config` 子命令：`ytdl config` 輸出設定目錄路徑（`~/.config/ytdl`，
  支援 `XDG_CONFIG_HOME`），若不存在則自動建立，維持 stdout 只有路徑契約。
- `monitor`（別名 `m`）子命令：採用 `bubbletea` + `lipgloss` 打造互動式
  下載清單瀏覽器，支援單鍵開啟檔案、子序列搜尋（`/`）、格式篩選（`tab`）、
  多欄位排序（`s`）與定時自動刷新，固定於 `~/.config/ytdl/data` 目錄。
- 下載引擎是 **yt-dlp**，不是純 Go library：YouTube 的簽章與節流邏輯
  變動頻繁，純 Go 實作維護成本高且高畫質仍需自行合流 (muxing)。
  ffmpeg 同時服務 mp4 合流與 mp3 轉檔，兩者在下載開始前一次檢查
  (`checkRuntime`)，避免 yt-dlp 下載完才在 postprocessor 階段失敗。
- **一個 tier 兩套刻度**：`videoHeights` 是解析度上限（tier 5 = 0 代表
  不設限），`audioBitrates` 是 MP3 bitrate。兩張表由測試釘住必須覆蓋
  `QUALITY_MIN..QUALITY_MAX` 且單調遞增——漏一格會讓 `-qtype` 靜默退回
  yt-dlp 預設。
- mp4 selector 尾端保留 `/b` fallback：只有超過上限畫質的影片仍可下載，
  不會因為找不到符合條件的 format 而整個失敗。
- 輸出契約：**stdout 只有檔案路徑**（`--print after_move:filepath`），
  進度與橫幅走 stderr（`--quiet --progress` 讓 yt-dlp 把進度改寫到
  stderr）。與 video-utils 的 stdout-only-paths 慣例一致，可直接
  `$(ytdl ...)` 取路徑。
- 輸出目錄固定為 `~/.config/ytdl/data`（不存在則自動建立），沒有 `-out`
  旗標可以改路徑：符合 unified interface 的 `data/` 慣例，也避免下載檔
  散落在使用者當下的工作目錄。
- 預設 `--no-playlist`：playlist URL 只抓所指向的單支影片，行為可預期。
- `signal.NotifyContext` 綁 `exec.CommandContext`，Ctrl-C 會連帶終止
  yt-dlp 子行程，不留孤兒。

## 開發與驗證 (Development and Verification)

```bash
gofmt -l . && go vet ./... && go test ./... -count=1   # 全離線
go build -o bin/ytdl . && ./bin/ytdl -help
```

單元測試全離線（只驗參數映射，不打網路）。真實下載屬手動 smoke test。
