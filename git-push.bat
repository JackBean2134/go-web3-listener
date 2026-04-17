@echo off
chcp 65001 >nul
echo ========================================
echo Git 初始化和提交脚本
echo ========================================
echo.

echo [1/6] 初始化Git仓库...
git init
if errorlevel 1 (
    echo Git初始化失败！
    pause
    exit /b 1
)
echo ✓ Git仓库初始化成功
echo.

echo [2/6] 添加所有文件到暂存区...
git add .
if errorlevel 1 (
    echo 添加文件失败！
    pause
    exit /b 1
)
echo ✓ 文件已添加到暂存区
echo.

echo [3/6] 创建初始提交...
git commit -m "Initial commit: Go Web3 Listener - BSC blockchain transfer monitoring service

Features:
- Multi-contract monitoring (USDT/BTCB/BNB)
- RPC node pool with health check and auto-failover
- Real-time transaction processing (10s polling)
- MySQL persistence with idempotent deduplication
- RESTful API for querying transfers
- Alert system (DingTalk & Email)
- High availability design with retry mechanisms"
if errorlevel 1 (
    echo 提交失败！
    pause
    exit /b 1
)
echo ✓ 初始提交创建成功
echo.

echo [4/6] 添加远程仓库...
git remote remove origin 2>nul
git remote add origin https://github.com/JackBean2134/go-web3-listener.git
if errorlevel 1 (
    echo 添加远程仓库失败！
    pause
    exit /b 1
)
echo ✓ 远程仓库添加成功
echo.

echo [5/6] 设置主分支...
git branch -M main
echo ✓ 主分支设置成功
echo.

echo [6/6] 推送到GitHub...
echo 注意：如果远程仓库已有内容，将使用强制推送
git push -u origin main --force
if errorlevel 1 (
    echo.
    echo 推送失败！尝试先拉取远程内容...
    git pull origin main --allow-unrelated-histories
    if errorlevel 1 (
        echo 拉取失败！请手动解决冲突后再次推送
        pause
        exit /b 1
    )
    git push -u origin main
    if errorlevel 1 (
        echo 推送仍然失败！请检查网络连接和GitHub凭据
        pause
        exit /b 1
    )
)
echo.
echo ========================================
echo ✓ 成功推送到GitHub！
echo 仓库地址: https://github.com/JackBean2134/go-web3-listener
echo ========================================
pause
