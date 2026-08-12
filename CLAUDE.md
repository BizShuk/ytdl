# CLAUDE.md — ytdl 技術脈絡 (Technical Context)

YouTube 下載 CLI。業務定義見 [README.md](README.md)。
Module `github.com/bizshuk/ytdl`，Go 1.26，**零第三方依賴**（無 go.sum）。

## 結構與 ownership

```text
ytdl/
├── main.go              # flag 解析 + 逐 URL 派工；process-level 決策只放這裡
└── pkg/download/
    ├── quality.go       # MediaType 與 quality tier → yt-dlp format selector
    └── download.go      # Request 驗證、runtime 檢查、exec yt-dlp
```

`pkg/download` 不認識 flag，`main.go` 不認識 yt-dlp 參數。

## 關鍵決策 (Key Decisions)

- CLI 用 **stdlib `flag`** 而非 cobra：使用者要的是 `-type` / `-qtype`
  單破折號長旗標，pflag 會把 `-type` 當成短旗標叢集解析而失敗。單一命令
  也不需要子命令樹。
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
- 預設 `--no-playlist`：playlist URL 只抓所指向的單支影片，行為可預期。
- `signal.NotifyContext` 綁 `exec.CommandContext`，Ctrl-C 會連帶終止
  yt-dlp 子行程，不留孤兒。

## 開發與驗證 (Development and Verification)

```bash
gofmt -l . && go vet ./... && go test ./... -count=1   # 全離線
go build -o bin/ytdl . && ./bin/ytdl -help
```

單元測試全離線（只驗參數映射，不打網路）。真實下載屬手動 smoke test。
