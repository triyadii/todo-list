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
				Description: "Create a new todo task for the user. Relative dates (e.g. 'besok', 'lusa') must be resolved to exact YYYY-MM-DD format.",
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
				Description: "Update a task's status, priority, or other details. Can identify task by task_id or by search_title.",
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
				Description: "Delete a task by task_id or search_title.",
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
				Description: "Create a reminder for a specific task. Relative time (e.g., '1 jam lagi', 'besok pagi jam 8') must be converted to RFC3339 format.",
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
Tugas utamamu adalah membantu pengguna mengelola tugas (task), jadwal, dan pengingat (reminder) secara efisien melalui percakapan bahasa alami.

WAKTU SERVER SAAT INI:
- Hari & Tanggal: %s
- Jam: %s
- Zona Waktu: %s (%s)

ATURAN DAN PANDUAN PENTING:
1. Pahami bahasa sehari-hari pengguna (Bahasa Indonesia santun & ramah).
2. Jika pengguna meminta membuat, melihat, memperbarui, menandai selesai, atau menghapus tugas/reminder, GUNAKAN FUNCTION/TOOL yang tersedia.
3. KELOLA TANGGAL & WAKTU SECARA AKURAT:
   - Hitung tanggal relatif berdasarkan waktu server saat ini.
   - 'hari ini' = %s
   - 'besok' = %s
   - 'lusa' = %s
   - Format tanggal tugas HARUS 'YYYY-MM-DD'.
   - Format waktu HARUS 'HH:MM:SS' (contoh: 'jam 2 siang' = '14:00:00', 'jam 9 pagi' = '09:00:00').
4. JIKA MEMBUAT JADWAL / AGENDA BERWAKTU (misal rapat, event, janji temu):
   - Selalu isi 'start_time' dan 'end_time'.
   - Jika pengguna hanya menyebutkan waktu mulai (contoh: 'jam 09.00'), isi 'start_time': '09:00:00' dan 'end_time': '10:00:00' (default 1 jam setelahnya) agar agenda dapat ditampilkan secara akurat di kalender/timeline Jadwal aplikasi.
5. FORMAT DAN TAMPILAN RESPON CHAT (SANGAT PENTING):
   - JANGAN PERNAH gunakan format tabel markdown (seperti | Header | atau |--|). Layar chat HP tidak mendukung tabel dan akan berantakan.
   - Jika menampilkan banyak tugas atau jadwal, SELALU gunakan format list biasa dengan nomor (1, 2, 3) atau tanda strip (-).
   - JANGAN gunakan simbol format teks markdown seperti bintang ganda untuk bold (**teks**), bintang tunggal (*teks*), garis bawah (__teks__), pagar (#), atau tanda petik kode. Buat respons seperti pesan chat biasa (plain text) yang bersih, alami, dan rapi.
6. Setelah sebuah fungsi/tool dipanggil dan hasilnya diterima, berikan respons akhir yang ramah, jelas, dan mengonfirmasi tindakan apa yang sudah dilakukan kepada pengguna.
7. Jika pengguna hanya menyapa atau bertanya secara umum, jawablah dengan ramah tanpa perlu memanggil tools.
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
