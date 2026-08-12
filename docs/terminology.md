# 術語表 (Terminology)

| 術語               | 定義                                                                        |
| ------------------ | --------------------------------------------------------------------------- |
| `MediaType`        | 輸出容器格式，僅 `mp3` 與 `mp4` 兩值；對應 `-type` 旗標                     |
| `quality tier`     | `-qtype` 的 1~5 數字等級，5 為最高、預設 3；同一數字在 mp3/mp4 對應不同刻度 |
| `tier ceiling`     | mp4 的 tier 語意：解析度上限而非精確值，選取「不超過該高度的最佳串流」      |
| `uncapped`         | mp4 tier 5，`videoHeights` 值為 `0`，代表交由 yt-dlp 選最佳畫質             |
| `format selector`  | 傳給 yt-dlp `--format` 的串流挑選運算式，例如 `bv*[height<=720]+ba/b`       |
| `remux`            | 不重新編碼、僅更換容器（mp4）；與 mp3 的 transcode（重新編碼）不同         |
| `output contract`  | stdout 只輸出檔案路徑、其餘訊息一律 stderr 的規約                           |
| `runtime check`    | 下載前確認 `yt-dlp` 與 `ffmpeg` 存在於 `PATH` 的前置檢查                    |
