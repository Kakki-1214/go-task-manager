# Go & Next.js Task Management App

フルスタック開発の技術実証として構築したタスク管理アプリケーションです。
バックエンドにGo (Gin), フロントエンドにNext.js (TypeScript)を採用し、モダンで型安全な開発環境を実現しています。

## 🛠 技術スタック (Tech Stack)

### Backend
- **Language:** Go 1.25
- **Framework:** Gin Web Framework
- **ORM:** GORM
- **Database:** MySQL 8.0
- **Environment:** Docker / Docker Compose (Hot Reload対応)

### Frontend
- **Framework:** Next.js 14+ (App Router)
- **Language:** TypeScript
- **Styling:** Tailwind CSS
- **Communication:** REST API

## 🚀 環境構築 (Setup)

以下の手順で、ローカル環境にてアプリケーションを起動できます。

## 前提条件
- Docker / Docker Desktop がインストールされていること
- Node.js (v18以上) がインストールされていること

## 1. リポジトリのクローン
```bash
git clone [https://github.com/](https://github.com/)【あなたのID】/task-manager-demo.git
cd task-manager-demo

2. バックエンドの起動 (Docker)
APIサーバーとデータベースを立ち上げます。

Bash

cd backend
docker-compose up --build
API Server: http://localhost:8080

Database: MySQL (Port 3307 -> 3306)

3. フロントエンドの起動 (Local)
別のターミナルを開き、フロントエンドを起動します。

Bash

cd frontend
npm install
npm run dev
Frontend App: http://localhost:3000

📦 機能一覧
タスク一覧取得 (GET /tasks)

タスク作成 (POST /tasks) - バリデーション付き

タスク削除 (DELETE /tasks/:id)

レスポンシブUI (Tailwind CSS)

🏗 アーキテクチャの特徴
コンテナ技術: バックエンド開発環境をDocker化し、環境差異を排除。Airによるホットリロードを導入し、開発効率を最適化。

CORS制御: Ginミドルウェアにてクロスオリジン通信を適切に制御。

エラーハンドリング: DB接続のリトライ処理などを実装し、堅牢性を確保。


※ **注意:** `【あなたのID】` の部分は、実際のあなたのGitHub IDに書き換えてください。

---

### 手順 2: 更新をGitHubに反映

VSCodeのターミナルで、変更をGitHubにプッシュします。

```powershell
git add README.md
git commit -m "Add documentation"

git push

# Go & Next.js Task Manager

A full-stack task management application built with Go (Gin) and Next.js, deployed on cloud infrastructure.

## 🚀 Features

- **User Authentication:** Secure JWT-based signup and login system.
- **Task Management:** Create, read, and delete tasks.
- **Data Isolation:** Users can only access their own tasks.
- **Responsive UI:** Built with Tailwind CSS.

## 🛠 Tech Stack

| Category | Technology |
| --- | --- |
| **Frontend** | Next.js (TypeScript), Tailwind CSS |
| **Backend** | Go (Gin Framework) |
| **Database** | MySQL (Aiven Cloud) |
| **Infra** | Docker, Render, Vercel |

## 🏗 Architecture

- **Frontend:** Hosted on Vercel. Consumes REST API.
- **Backend:** Containerized Go application hosted on Render.
- **Database:** Managed MySQL on Aiven.

## 🔒 Security

- Passwords are hashed using **bcrypt** before storage.
- API endpoints are protected using custom **JWT middleware**.
- **CORS** policies configured for secure cross-origin requests.

## 👨‍💻 Author

Keigo Kakizawa (Japan Institute of Technology)
