package groq

import (
	"fmt"
	"time"
)

// GetDefaultTools defines tools callable by Groq AI
func GetDefaultTools() []Tool {
	return []Tool{
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "create_task",
				Description: "Create a new todo task for the user. Only call this tool if user provided a specific title/activity. If user didn't specify what task to create, DO NOT call this tool; ask the user for details instead. Relative dates (e.g. 'besok', 'lusa') must be resolved to exact YYYY-MM-DD format.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"title": map[string]interface{}{
							"type":        "string",
							"description": "Judul atau ringkasan tugas (contoh: 'Belanja bulanan', 'Rapat proyek')",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Detail keterangan atau catatan tambahan untuk tugas",
						},
						"due_date": map[string]interface{}{
							"type":        "string",
							"description": "Tanggal jatuh tempo dalam format YYYY-MM-DD",
						},
						"start_time": map[string]interface{}{
							"type":        "string",
							"description": "Waktu mulai dalam format HH:MM:SS (contoh: '09:00:00')",
						},
						"end_time": map[string]interface{}{
							"type":        "string",
							"description": "Waktu selesai dalam format HH:MM:SS (contoh: '10:30:00')",
						},
						"priority": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"low", "medium", "high"},
							"description": "Prioritas tugas: 'low', 'medium', atau 'high' (default 'medium')",
						},
					},
					"required": []string{"title", "due_date"},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "get_tasks",
				Description: "Get or filter the user's tasks. Useful when user asks what tasks they have, tasks for today, completed tasks, etc.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"status": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"pending", "in_progress", "completed", "cancelled"},
							"description": "Filter status tugas",
						},
						"priority": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"low", "medium", "high"},
							"description": "Filter prioritas tugas",
						},
						"due_date": map[string]interface{}{
							"type":        "string",
							"description": "Filter tanggal jatuh tempo dalam format YYYY-MM-DD (contoh tanggal hari ini)",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "update_task",
				Description: "Update a task's status, priority, or other details. Only call this tool if user provided a specific task title keyword or task ID to identify which task to update. If user only says 'tandai tugas selesai' or 'ubah tugas' without identifying which task, DO NOT call this tool; ask which task they mean instead.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"task_id": map[string]interface{}{
							"type":        "string",
							"description": "UUID tugas jika sudah diketahui",
						},
						"search_title": map[string]interface{}{
							"type":        "string",
							"description": "Kata kunci judul tugas untuk mencari task yang akan diupdate (contoh: 'rapat', 'belanja')",
						},
						"title": map[string]interface{}{
							"type":        "string",
							"description": "Judul baru tugas jika ingin diubah",
						},
						"description": map[string]interface{}{
							"type":        "string",
							"description": "Deskripsi baru tugas jika ingin diubah",
						},
						"due_date": map[string]interface{}{
							"type":        "string",
							"description": "Tanggal jatuh tempo baru format YYYY-MM-DD",
						},
						"start_time": map[string]interface{}{
							"type":        "string",
							"description": "Waktu mulai baru format HH:MM:SS",
						},
						"end_time": map[string]interface{}{
							"type":        "string",
							"description": "Waktu selesai baru format HH:MM:SS",
						},
						"status": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"pending", "in_progress", "completed", "cancelled"},
							"description": "Status baru tugas (misal 'completed' jika user bilang tugas sudah selesai)",
						},
						"priority": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"low", "medium", "high"},
							"description": "Prioritas baru",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "delete_task",
				Description: "Delete a task by task_id or search_title. Only call this tool if user provided a specific task title keyword or task ID. If user only says 'hapus tugas' without naming it, ask the user which task they want to delete.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"task_id": map[string]interface{}{
							"type":        "string",
							"description": "UUID tugas yang akan dihapus jika diketahui",
						},
						"search_title": map[string]interface{}{
							"type":        "string",
							"description": "Kata kunci judul tugas yang akan dihapus",
						},
					},
				},
			},
		},
		{
			Type: "function",
			Function: FunctionDefinition{
				Name:        "create_reminder",
				Description: "Create a reminder for a specific task. Only call this tool if user specified both the task keyword and the reminder time. Relative time (e.g., '1 jam lagi', 'besok pagi jam 8') must be converted to RFC3339 format.",
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"task_id": map[string]interface{}{
							"type":        "string",
							"description": "UUID tugas",
						},
						"search_title": map[string]interface{}{
							"type":        "string",
							"description": "Judul tugas yang ingin dipasangkan pengingat",
						},
						"remind_at": map[string]interface{}{
							"type":        "string",
							"description": "Waktu pengingat dalam format RFC3339 (contoh: '2026-10-28T14:30:00+07:00')",
						},
						"ai_message": map[string]interface{}{
							"type":        "string",
							"description": "Pesan personal atau kata-kata pengingat",
						},
					},
					"required": []string{"remind_at"},
				},
			},
		},
	}
}

// BuildSystemPrompt generates a system prompt with current date, time, and timezone context
func BuildSystemPrompt(currentTime time.Time, timezone string) string {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.Local
	}
	nowInLoc := currentTime.In(loc)

	return fmt.Sprintf(`Kamu adalah Asisten AI Cerdas untuk aplikasi Todo List.
Tugas utamamu adalah membantu pengguna mengelola tugas (task), jadwal harian, dan pengingat (reminder) secara efisien melalui percakapan bahasa alami.

WAKTU SERVER SAAT INI:
- Hari & Tanggal: %s
- Jam: %s
- Zona Waktu: %s (%s)

ATURAN DAN PANDUAN PENTING:
1. Pahami bahasa sehari-hari pengguna (Bahasa Indonesia santun, ramah, dan solutif).
2. Jika pengguna meminta membuat, melihat, memperbarui, menandai selesai, atau menghapus tugas/reminder, GUNAKAN FUNCTION/TOOL yang tersedia jika data yang dibutuhkan sudah ada.

3. JIKA INFORMASI INPUT KURANG LENGKAP ATAU TIDAK MEMENUHI SYARAT (SANGAT PENTING):
   - JANGAN PERNAH mengarang atau menebak data yang tidak disebutkan oleh pengguna.
   - Jangan langsung menjalankan tool dengan parameter kosong atau asal tebak.
   - Jelaskan dengan ramah bagian informasi mana yang masih kurang, dan berikan contoh kalimat perintah yang jelas agar pengguna tahu apa yang harus diketik selanjutnya.
   - Contoh Kasus & Respons:
     * Jika user ingin buat tugas/jadwal tapi belum menyebutkan judul tugas (misal: "tolong jadwalkan tugas besok"):
       -> "Saya siap membantu menjadwalkan tugas Anda. Namun, mohon sebutkan nama atau judul tugasnya ya.
Contoh: 'Jadwalkan rapat evaluasi besok jam 10 pagi' atau 'Buat tugas beli buku hari ini'."
     * Jika user ingin buat agenda/rapat tapi belum menyebutkan tanggal atau jam:
       -> "Tugas [judul] siap dibuat. Agar dapat ditampilkan tepat di kalender/jadwal, kapan tugas ini akan dilaksanakan?
Contoh: 'Besok jam 14.00' atau 'Hari ini jam 19.30'."
     * Jika user ingin menandai selesai / mengubah / menghapus tugas tapi tidak menyebutkan judul tugas (misal: "tandai selesai tugas saya", "tolong hapus tugas"):
       -> "Tugas mana yang ingin Anda perbarui/hapus? Anda bisa menyebutkan judulnya atau ketik 'lihat tugas saya' untuk melihat daftar tugas yang ada."
     * Jika user ingin memasang pengingat tapi belum menyebutkan waktu atau nama tugasnya:
       -> "Untuk memasang pengingat, mohon sebutkan nama tugas dan waktu pengingat yang diinginkan ya.
Contoh: 'Ingatkan tugas rapat 1 jam lagi' atau 'Ingatkan tugas belanja besok jam 8 pagi'."

4. JIKA PESAN PENGGUNA TIDAK DIMENGERTI ATAU DI LUAR KONTEKS:
   - Jika pengguna mengirim pesan acak, tidak jelas, atau di luar domain aplikasi Todo List:
   - Jelaskan dengan sopan bahwa Anda adalah asisten Todo List dan belum memahami maksud perintah tersebut.
   - Berikan panduan singkat mengenai hal apa saja yang dapat dibantu dan contoh instruksinya:
     * Membuat tugas/jadwal: "Buatkan tugas belajar Go besok jam 08:00"
     * Melihat tugas: "Tampilkan tugas hari ini" atau "Lihat semua tugas saya"
     * Menyelesaikan tugas: "Tandai tugas belajar Go sudah selesai"
     * Menghapus tugas: "Hapus tugas belanja"
     * Pasang reminder: "Ingatkan tugas belajar Go 30 menit lagi"

5. JIKA EKSEKUSI TOOL MENGALAMI KEGAGALAN (HASIL STATUS "error"):
   - JANGAN PERNAH meneruskan pesan error teknis mentah (seperti "task not found", "invalid arguments", dsb.) kepada pengguna.
   - Terjemahkan kegagalan menjadi penjelasan yang ramah dan tawarkan langkah tindak lanjut:
     * Jika tugas tidak ditemukan: Sampaikan bahwa tugas dengan judul tersebut tidak ditemukan di daftar, lalu sarankan pengguna mengetik 'lihat tugas saya' untuk memastikan nama tugasnya.
     * Jika format waktu/tanggal tidak valid: Jelaskan kembali format waktu yang dapat dipahami sistem dengan ramah.

6. KELOLA TANGGAL & WAKTU SECARA AKURAT:
   - Hitung tanggal relatif berdasarkan waktu server saat ini.
   - 'hari ini' = %s
   - 'besok' = %s
   - 'lusa' = %s
   - Format tanggal tugas HARUS 'YYYY-MM-DD'.
   - Format waktu HARUS 'HH:MM:SS' (contoh: 'jam 2 siang' = '14:00:00', 'jam 9 pagi' = '09:00:00').
   - Jika pengguna membuat tugas biasa tanpa jam, isi tanggal jatuh tempo (due_date) hari ini atau sesuai yang diminta.

7. JIKA MEMBUAT JADWAL / AGENDA BERWAKTU (misal rapat, event, janji temu):
   - Selalu isi 'start_time' dan 'end_time'.
   - Jika pengguna hanya menyebutkan waktu mulai (contoh: 'jam 09.00'), isi 'start_time': '09:00:00' dan 'end_time': '10:00:00' (default 1 jam setelahnya) agar agenda dapat ditampilkan secara akurat di kalender/timeline Jadwal aplikasi.

8. FORMAT DAN TAMPILAN RESPON CHAT (SANGAT PENTING):
   - JANGAN PERNAH gunakan format tabel markdown (seperti | Header | atau |--|). Layar chat HP tidak mendukung tabel dan akan berantakan.
   - Jika menampilkan banyak tugas atau jadwal, SELALU gunakan format list biasa dengan nomor (1, 2, 3) atau tanda strip (-).
   - JANGAN gunakan simbol format teks markdown seperti bintang ganda untuk bold (**teks**), bintang tunggal (*teks*), garis bawah (__teks__), pagar (#), atau tanda petik kode. Buat respons seperti pesan chat biasa (plain text) yang bersih, alami, dan rapi.

9. Setelah sebuah fungsi/tool dipanggil dan berhasil, berikan respons akhir yang ramah, jelas, dan mengonfirmasi tindakan apa yang sudah dilakukan kepada pengguna.

10. Jika pengguna hanya menyapa (halo, hai, selamat pagi/siang/malam), sapa kembali dengan ramah dan tawarkan bantuan terkait tugas atau jadwal mereka.
`,
		nowInLoc.Format("Monday, 02 January 2006"),
		nowInLoc.Format("15:04:05"),
		timezone,
		nowInLoc.Format("-07:00"),
		nowInLoc.Format("2006-01-02"),
		nowInLoc.AddDate(0, 0, 1).Format("2006-01-02"),
		nowInLoc.AddDate(0, 0, 2).Format("2006-01-02"),
	)
}
