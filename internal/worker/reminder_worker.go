package worker

import (
	"context"
	"fmt"
	"log"
	"time"

	"todo-backend/internal/service"
	"todo-backend/pkg/fcm"
)

type ReminderWorker struct {
	reminderService service.ReminderService
	fcmClient       *fcm.Client
	interval        time.Duration
}

func NewReminderWorker(reminderService service.ReminderService, fcmClient *fcm.Client, interval time.Duration) *ReminderWorker {
	if interval <= 0 {
		interval = 15 * time.Second
	}
	return &ReminderWorker{
		reminderService: reminderService,
		fcmClient:       fcmClient,
		interval:        interval,
	}
}

// Start begins the background reminder checking loop until context is canceled
func (w *ReminderWorker) Start(ctx context.Context) {
	log.Printf("[INFO] [REMINDER WORKER] Background reminder checker started (polling interval: %v)", w.interval)
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	// Perform initial check immediately
	w.checkReminders(time.Now())

	for {
		select {
		case <-ctx.Done():
			log.Println("[INFO] [REMINDER WORKER] Background reminder checker stopped gracefully")
			return
		case now := <-ticker.C:
			w.checkReminders(now)
		}
	}
}

func (w *ReminderWorker) checkReminders(now time.Time) {
	reminders, err := w.reminderService.GetPendingReminders(now)
	if err != nil {
		log.Printf("[ERROR] [REMINDER WORKER] Failed to query pending reminders: %v", err)
		return
	}

	if len(reminders) == 0 {
		return
	}

	log.Printf("[INFO] [REMINDER WORKER] Found %d due reminder(s) to process at %s", len(reminders), now.Format("2006-01-02 15:04:05"))

	for _, r := range reminders {
		taskTitle := "Tugas"
		userName := "User"
		userEmail := "-"
		var fcmToken *string

		if r.Task != nil {
			taskTitle = r.Task.Title
			if r.Task.User != nil {
				userName = r.Task.User.Name
				userEmail = r.Task.User.Email
				fcmToken = r.Task.User.FCMToken
			}
		}

		aiNote := ""
		if r.AIMessage != nil {
			aiNote = *r.AIMessage
		}

		log.Printf("[INFO] [REMINDER TRIGGERED] ⏰ Task: '%s' | User: %s (%s) | RemindAt: %s | Note: %s",
			taskTitle, userName, userEmail, r.RemindAt.Format("15:04:05"), aiNote)

		if w.fcmClient != nil && fcmToken != nil && *fcmToken != "" {
			log.Printf("[INFO] [REMINDER WORKER] Sending Push Notification to FCM Token: %s...", *fcmToken)
			body := fmt.Sprintf("Waktunya mengerjakan tugas '%s'", taskTitle)
			if aiNote != "" {
				body = aiNote
			}
			fcmPayload := &fcm.NotificationPayload{
				Title: fmt.Sprintf("⏰ Pengingat: %s", taskTitle),
				Body:  body,
				Data: map[string]string{
					"type":        "reminder",
					"task_id":     r.TaskID.String(),
					"reminder_id": r.ID.String(),
					"remind_at":   r.RemindAt.Format(time.RFC3339),
				},
			}
			go func(tok string, payload *fcm.NotificationPayload) {
				ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
				defer cancel()
				if _, err := w.fcmClient.Send(ctx, tok, payload); err != nil {
					log.Printf("[WARN] [REMINDER WORKER] Failed to deliver FCM notification: %v", err)
				}
			}(*fcmToken, fcmPayload)
		}

		// Mark reminder as sent so it won't be processed again
		if err := w.reminderService.MarkReminderSent(r.ID); err != nil {
			log.Printf("[ERROR] [REMINDER WORKER] Failed to mark reminder %s as sent: %v", r.ID, err)
		} else {
			log.Printf("[INFO] [REMINDER WORKER] Reminder %s marked as sent successfully", r.ID)
		}
	}
}
