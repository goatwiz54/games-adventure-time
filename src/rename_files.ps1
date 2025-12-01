# .go と .txt の両方を対象とする
# -Include を使用して、カレントディレクトリのみを検索
#$files = Get-ChildItem -Path "." -Include "*.go", "*.txt" | Where-Object { -not $_.PSIsContainer }
$files = Get-ChildItem -Path ".\*.go", ".\*.txt" | Where-Object { -not $_.PSIsContainer }


foreach ($file in $files) {
        # 1行目を読み込む
    # -Raw を使用して、ファイル全体を単一の文字列として取得し、
    # Select-Object -First 1 で安全に1行目を取得
    $firstLine = (Get-Content $file.FullName -TotalCount 1)
    
    # パターンを定義
    $goPattern = '^// filename:\s*(.+)$'
    $txtPattern = '^\s*<!--\s*filename:\s*(.+?)\s*-->$'
    
    $targetName = $null
    
    # ファイル名の抽出
    if ($firstLine -match $goPattern) {
        $targetName = $matches[1].Trim()
    }
    elseif ($firstLine -match $txtPattern) {
        $targetName = $matches[1].Trim()
    }

    # 名前が取得できた場合のみ処理を続行
    if ($targetName -ne $null -and $file.Name -ne $targetName) {
        # ファイル名のみを取得（ディレクトリパスは無視）
        $fileName = [System.IO.Path]::GetFileName($targetName)

        # 同じディレクトリ内でリネーム
        $destination = Join-Path $file.DirectoryName $fileName
        Rename-Item -Path $file.FullName -NewName $fileName -Force
        Write-Host "Renamed: $($file.Name) -> $fileName" -ForegroundColor Cyan
    }
}

Write-Host "Done." -ForegroundColor Green
