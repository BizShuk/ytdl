# 2026-08-12 — 本機 yt-dlp 只取得 360p，畫質分級無法端到端驗證

## 現象

`ytdl -type mp4 -qtype 1/3/5` 對同一支影片全部產出 640x360。以
`yt-dlp -F` 直接檢查，多解析度影片（如 `aqz-KE-bpKQ`，實際有 4K）也只列出
單一 format `18`（640x360 mp4）。

## 根因

不是 format selector 的問題。yt-dlp `2026.02.04` 對本機 session 命中 YouTube
的 SABR-only 實驗，warning 明示：

```text
Some android_vr client https formats have been skipped as they are missing a
URL. YouTube may have enabled the SABR-only streaming experiment.
```

`--extractor-args youtube:player_client=tv|ios|web` 三種 client 全數回傳 0 個
mp4/webm format，代表這是抽取層的環境條件，不是 `-qtype` 映射的缺陷。

## 影響與處置

- `-qtype` 的上限語意（ceiling + 尾端 `/b` fallback）在此情況下仍正確：找不到
  符合上限的串流時退回唯一可用畫質，不會整個失敗——所以症狀是「畫質一律
  360p」而非報錯。
- 單元測試只釘住參數映射，不打網路，因此不受此環境條件影響。

## 結果（2026-08-31：已解除）

天花板隨 yt-dlp 升級消失，本機版本自 `2026.07.04` 升到 `2026.08.19`
（由 Homebrew 安裝，`brew upgrade yt-dlp`，`不是` pip —— 上面原本寫的 pip
路徑在這台機器上並不存在）。同一支 `aqz-KE-bpKQ` 現在列得出到 3840x2160，
`-qtype` 的分級因此能端到端驗證。

期間曾以 `--extractor-args youtube:player_client=web_embedded,default` 繞開
SABR 與 403，該釘選`已一併移除`：新版預設挑的 client 自己就拿得到完整
format 清單，實測釘選與預設列出的解析度完全相同。留著一個沒有作用的釘選
，只會在下一次 YouTube 改動時被誤認成仍在生效的防線。

`修法就是升級抽取器本身`，不是在參數上疊變通 —— YouTube 的簽章與節流是
移動標靶，任何在本地猜 client 的寫法都只在某個版本區間內成立。
