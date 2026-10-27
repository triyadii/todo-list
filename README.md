# 📝 TodoList Backend API (Go Fiber + GORM + PostgreSQL)

Backend RESTful API untuk aplikasi Todo List cerdas berbasis Go (Golang), menggunakan web framework **Fiber v2**, **GORM ORM**, dan database **PostgreSQL**. Dilengkapi dengan dukungan AI context chat, smart reminder scheduler, dan FCM token untuk push notification.

---

## 🏗️ Arsitektur Proyek (Clean & Modular Architecture)

```
todo-backend/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point aplikasi & graceful shutdown
├── config/
│   ├── config.go                  # Load env & default config
│   └── database.go                # Koneksi PostgreSQL & auto-migration GORM
├── internal/
│   ├── handler/                   # HTTP Controller / Request Handlers
│   │   ├── auth_handler.go
│   │   ├── task_handler.go
│   │   ├── reminder_handler.go
│   │   └── chat_handler.go
│   ├── middleware/                # JWT Auth & Logging Middlewares
│   │   └── auth_middleware.go
│   ├── models/                    # GORM Models & Enums
│   │   ├── user.go
│   │   ├── task.go
│   │   ├── reminder.go
│   │   └── chat_message.go
│   ├── repository/                # Database Queries Layer
│   │   ├── user_repository.go
│   │   ├── task_repository.go
│   │   ├── reminder_repository.go
│   │   └── chat_repository.go
│   ├── routes/                    # Definisi & routing endpoint API
│   │   └── routes.go
│   └── service/                   # Business Logic Layer
│       ├── auth_service.go
│       ├── task_service.go
│       ├── reminder_service.go
│       └── chat_service.go
├── pkg/
│   └── utils/                     # Helper (JWT, Bcrypt, Standard JSON Response)
│       ├── jwt.go
│       ├── password.go
│       └── response.go
├── migrations/
│   └── 000001_init_schema.up.sql  # Skema DDL SQL murni dengan indeks & constraint
├── .env.example                   # Contoh konfigurasi environment
├── .env                           # Konfigurasi aktif
└── README.md
```

---

## 🗄️ Desain Tabel Database

### 1. Tabel `users`
Menyimpan identitas pengguna, zona waktu, dan token push notification.
| Kolom | Tipe Data | Deskripsi |
|---|---|---|
| `id` | UUID (PK) | Primary Key (Default: `gen_random_uuid()`) |
| `name` | VARCHAR(255) | Nama lengkap pengguna |
| `email` | VARCHAR(255) | Email unik (Indexed) |
| `password_hash` | VARCHAR(255) | Password terenkripsi bcrypt |
| `timezone` | VARCHAR(100) | Default: `'Asia/Jakarta'` |
| `fcm_token` | TEXT (nullable) | Token Firebase Cloud Messaging untuk push notification mobile |
| `created_at` | TIMESTAMPTZ | Waktu pendaftaran |
| `updated_at` | TIMESTAMPTZ | Waktu pembaruan profil |

### 2. Tabel `tasks`
Menyimpan agenda kegiatan pengguna.
| Kolom | Tipe Data | Deskripsi |
|---|---|---|
| `id` | UUID (PK) | Primary Key (Default: `gen_random_uuid()`) |
| `user_id` | UUID (FK) | Relasi ke `users.id` (Indexed, ON DELETE CASCADE) |
| `title` | VARCHAR(255) | Judul kegiatan |
| `description` | TEXT (nullable) | Rincian kegiatan |
| `due_date` | DATE | Tanggal kegiatan (`YYYY-MM-DD`, Indexed) |
| `start_time` | TIME (nullable) | Jam mulai kegiatan (`HH:MM:SS`) |
| `end_time` | TIME (nullable) | Jam selesai kegiatan (`HH:MM:SS`) |
| `status` | VARCHAR(50) | Enum: `'pending'`, `'in_progress'`, `'completed'`, `'cancelled'` |
| `priority` | VARCHAR(50) | Enum: `'low'`, `'medium'`, `'high'` |
| `created_at` | TIMESTAMPTZ | Waktu pembuatan |
| `updated_at` | TIMESTAMPTZ | Waktu pembaruan |

### 3. Tabel `reminders`
Mengatur jadwal pengingat untuk kegiatan/task tertentu. Satu task bisa memiliki multiple reminder (misal: H-1 hari dan H-30 menit).
| Kolom | Tipe Data | Deskripsi |
|---|---|---|
| `id` | UUID (PK) | Primary Key (Default: `gen_random_uuid()`) |
| `task_id` | UUID (FK) | Relasi ke `tasks.id` (Indexed, ON DELETE CASCADE) |
| `remind_at` | TIMESTAMPTZ | Waktu eksekusi pengingat (Indexed) |
| `is_sent` | BOOLEAN | Status pengiriman pengingat (Default: `false`, Indexed) |
| `ai_message` | TEXT (nullable) | Pesan pengingat personalisasi buatan AI |
| `sent_at` | TIMESTAMPTZ (nullable) | Waktu pengingat berhasil dikirim |
| `created_at` | TIMESTAMPTZ | Waktu reminder dibuat |
| `updated_at` | TIMESTAMPTZ | Waktu reminder diupdate |

### 4. Tabel `chat_messages`
Menyimpan riwayat percakapan antara pengguna dan asisten AI agar memiliki konteks memori interaksi sebelumnya.
| Kolom | Tipe Data | Deskripsi |
|---|---|---|
| `id` | UUID (PK) | Primary Key (Default: `gen_random_uuid()`) |
| `user_id` | UUID (FK) | Relasi ke `users.id` (Indexed, ON DELETE CASCADE) |
| `role` | VARCHAR(50) | Role pesan: `'user'`, `'assistant'`, atau `'system'` |
| `content` | TEXT | Konten teks percakapan |
| `created_at` | TIMESTAMPTZ | Waktu pesan dibuat (Indexed) |

---

## 🚀 Cara Menjalankan

### 1. Persiapan Database PostgreSQL
Buat database di PostgreSQL:
```sql
CREATE DATABASE todolist_db;
```

### 2. Atur File `.env`
Sesuaikan kredensial di file `.env`:
```env
APP_PORT=8080
APP_ENV=development

DB_HOST=127.0.0.1
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=todolist_db
DB_SSLMODE=disable
DB_TIMEZONE=Asia/Jakarta

JWT_SECRET=rahasia_jwt_super_aman_2026
JWT_EXPIRE_HOURS=72
```

### 3. Jalankan Aplikasi
Aplikasi secara otomatis melakukan **Auto-Migrate** tabel ke PostgreSQL saat pertama kali dinyalakan:
```bash
# Menggunakan air (live-reload) dari root:
air

# Atau menggunakan go run:
go run cmd/api/main.go
```

---

## 🐳 Menjalankan dengan Docker

Proyek ini telah dilengkapi [Dockerfile](file:///var/www/html/golang/todo-list/Dockerfile) (*multi-stage build*, image ultra-ringan berbasis Alpine) dan [docker-compose.yml](file:///var/www/html/golang/todo-list/docker-compose.yml).

### Menghubungkan ke Database PostgreSQL yang Ada di Host:
Secara default di `docker-compose.yml`, aplikasi menggunakan konfigurasi `network_mode: "host"` (sangat direkomendasikan di Linux). Ini memungkinkan container mengakses PostgreSQL di host via `127.0.0.1:5432` tanpa perlu mengubah konfigurasi `pg_hba.conf` PostgreSQL Anda.

```bash
# Build dan jalankan container
docker compose up -d --build

# Melihat log aplikasi
docker compose logs -f

# Menghentikan container
docker compose down
```

---

## 📖 Dokumentasi Swagger UI

Setelah aplikasi berjalan, dokumentasi interaktif Swagger UI dapat langsung diakses melalui browser:

👉 **[http://localhost:8080/swagger/](http://localhost:8080/swagger/)**

### Cara Autentikasi di Swagger UI:
1. Jalankan request `POST /api/v1/auth/login` melalui Swagger UI.
2. Salin token JWT dari response (`token`).
3. Klik tombol hijau **Authorize** (ikon gembok) di kanan atas Swagger UI.
4. Masukkan dengan format: `Bearer <token_anda>`
5. Klik **Authorize**. Sekarang seluruh endpoint yang dilindungi JWT dapat diuji langsung dari Swagger UI!

### Memperbarui Dokumentasi Swagger:
Jika ada perubahan endpoint atau anotasi Swagger di handler:
```bash
swag init -g cmd/api/main.go -o docs
```

---

## 📡 Daftar Endpoint API

### Public Endpoints
- `GET /health` : Cek status kesehatan server
- `POST /api/v1/auth/register` : Registrasi akun baru (menerima `name`, `email`, `password`, `timezone`, `fcm_token`)
- `POST /api/v1/auth/login` : Login user & peroleh Bearer JWT token

### Protected Endpoints (Header: `Authorization: Bearer <TOKEN>`)

#### User & Profil
- `GET /api/v1/user/profile` : Mengambil data profil user yang sedang login
- `PUT /api/v1/user/profile` : Update profil (nama, timezone, fcm_token)

#### User Management (Kelola Pengguna)
- `POST /api/v1/users` : Menambah / membuat user baru (`name`, `email`, `password`, `timezone`, `fcm_token`)
- `GET /api/v1/users` : Mengambil seluruh daftar pengguna
- `GET /api/v1/users/:id` : Mengambil detail pengguna berdasarkan ID (UUID)
- `DELETE /api/v1/users/:id` : Menghapus pengguna (cascade ke tasks, reminders, chat messages)

#### Tasks / Activities
- `POST /api/v1/tasks` : Membuat agenda kegiatan baru
- `GET /api/v1/tasks` : Mengambil seluruh task user (Support query params: `?status=pending&priority=high&due_date=2026-10-28`)
- `GET /api/v1/tasks/:id` : Mengambil detail task beserta relasi reminders
- `PUT /api/v1/tasks/:id` : Memperbarui task
- `DELETE /api/v1/tasks/:id` : Menghapus task (cascade ke reminders terkait)

#### Reminders
- `POST /api/v1/reminders` : Menjadwalkan pengingat (`task_id`, `remind_at`, `ai_message`)
- `GET /api/v1/tasks/:taskId/reminders` : Mengambil semua pengingat pada task tertentu
- `DELETE /api/v1/reminders/:id` : Menghapus reminder

#### Chat Messages (AI Context)
- `POST /api/v1/chat/messages` : Menyimpan pesan chat (`role`: `user`/`assistant`, `content`)
- `GET /api/v1/chat/messages?limit=50` : Mengambil riwayat percakapan untuk konteks AI
- `DELETE /api/v1/chat/messages` : Menghapus riwayat chat user
