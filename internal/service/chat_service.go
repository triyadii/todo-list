package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"todo-backend/internal/models"
	"todo-backend/internal/repository"
	"todo-backend/pkg/groq"

	"github.com/google/uuid"
)

type ActionLog struct {
	ToolName string      `json:"tool_name"`
	Status   string      `json:"status"` // "success" or "failed"
	Message  string      `json:"message"`
	Data     interface{} `json:"data,omitempty"`
}

type ChatResponse struct {
	UserMessage      *models.ChatMessage `json:"user_message"`
	AssistantMessage *models.ChatMessage `json:"assistant_message"`
	Actions          []ActionLog         `json:"actions,omitempty"`
}

type ChatService interface {
	SendMessage(userID uuid.UUID, req *SendMessageRequest) (*ChatResponse, error)
	GetHistory(userID uuid.UUID, limit int) ([]models.ChatMessage, error)
	ClearHistory(userID uuid.UUID) error
}

type chatService struct {
	chatRepo        repository.ChatRepository
	taskService     TaskService
	reminderService ReminderService
	groqClient      *groq.Client
	timezone        string
}

func NewChatService(
	chatRepo repository.ChatRepository,
	taskService TaskService,
	reminderService ReminderService,
	groqClient *groq.Client,
	timezone string,
) ChatService {
	if timezone == "" {
		timezone = "Asia/Jakarta"
	}
	return &chatService{
		chatRepo:        chatRepo,
		taskService:     taskService,
		reminderService: reminderService,
		groqClient:      groqClient,
		timezone:        timezone,
	}
}

type SendMessageRequest struct {
	Role    models.ChatRole `json:"role"`
	Content string          `json:"content"`
}

func (s *chatService) SendMessage(userID uuid.UUID, req *SendMessageRequest) (*ChatResponse, error) {
	if req.Content == "" {
		return nil, errors.New("content is required")
	}

	role := req.Role
	if role == "" {
		role = models.ChatRoleUser
	}
	if role != models.ChatRoleUser && role != models.ChatRoleAssistant && role != models.ChatRoleSystem {
		return nil, errors.New("role must be 'user', 'assistant', or 'system'")
	}

	// 1. Save user's message
	userMsg := &models.ChatMessage{
		UserID:  userID,
		Role:    role,
		Content: req.Content,
	}
	if err := s.chatRepo.Create(userMsg); err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// If role is assistant/system or groqClient is not initialized, return without invoking AI
	if role != models.ChatRoleUser || s.groqClient == nil {
		return &ChatResponse{
			UserMessage: userMsg,
		}, nil
	}

	// 2. Fetch recent conversation history for memory context
	history, err := s.chatRepo.FindByUserID(userID, 15)
	if err != nil {
		log.Printf("[WARN] Failed to fetch chat history for user %s: %v", userID, err)
	}

	// 3. Assemble Groq messages sequence
	systemPrompt := groq.BuildSystemPrompt(time.Now(), s.timezone)
	groqMessages := []groq.Message{
		{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	for _, h := range history {
		if h.ID == userMsg.ID {
			continue // skip current message, we'll append it
		}
		groqMessages = append(groqMessages, groq.Message{
			Role:    string(h.Role),
			Content: h.Content,
		})
	}
	groqMessages = append(groqMessages, groq.Message{
		Role:    "user",
		Content: req.Content,
	})

	// 4. Send request to Groq API with function definitions
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
	defer cancel()

	initialCompletion, err := s.groqClient.ChatCompletion(ctx, &groq.ChatCompletionRequest{
		Messages:    groqMessages,
		Tools:       groq.GetDefaultTools(),
		ToolChoice:  "auto",
		Temperature: 0.2,
	})
	if err != nil {
		log.Printf("[ERROR] Groq API call failed: %v", err)
		// Fallback assistant response
		assistantMsg := &models.ChatMessage{
			UserID:  userID,
			Role:    models.ChatRoleAssistant,
			Content: "Maaf, saat ini asisten AI sedang mengalami kendala koneksi ke server Groq. Silakan coba lagi beberapa saat lagi.",
		}
		_ = s.chatRepo.Create(assistantMsg)
		return &ChatResponse{
			UserMessage:      userMsg,
			AssistantMessage: assistantMsg,
		}, nil
	}

	if len(initialCompletion.Choices) == 0 {
		return &ChatResponse{UserMessage: userMsg}, nil
	}

	choice := initialCompletion.Choices[0]
	var actionsExecuted []ActionLog
	assistantReplyText := choice.Message.Content

	// 5. Handle Tool Calls if triggered by AI
	if len(choice.Message.ToolCalls) > 0 {
		// Append assistant's intent with tool calls to context
		groqMessages = append(groqMessages, choice.Message)

		for _, toolCall := range choice.Message.ToolCalls {
			actionLog, toolResultJSON := s.executeToolCall(userID, toolCall)
			actionsExecuted = append(actionsExecuted, actionLog)

			// Append tool execution response
			groqMessages = append(groqMessages, groq.Message{
				Role:       "tool",
				ToolCallID: toolCall.ID,
				Name:       toolCall.Function.Name,
				Content:    toolResultJSON,
			})
		}

		// 6. Request final conversational response from Groq based on tool results
		secondCompletion, err := s.groqClient.ChatCompletion(ctx, &groq.ChatCompletionRequest{
			Messages:    groqMessages,
			Tools:       groq.GetDefaultTools(),
			Temperature: 0.3,
		})
		if err != nil {
			log.Printf("[WARN] Groq second completion failed: %v", err)
			assistantReplyText = "Permintaan Anda telah berhasil diproses oleh sistem."
		} else if len(secondCompletion.Choices) > 0 {
			assistantReplyText = secondCompletion.Choices[0].Message.Content
		}
	}

	// 7. Sanitize response (plain text, no bold symbols, no tables) and save to Database
	cleanedContent := cleanChatText(assistantReplyText)
	assistantMsg := &models.ChatMessage{
		UserID:  userID,
		Role:    models.ChatRoleAssistant,
		Content: cleanedContent,
	}
	if err := s.chatRepo.Create(assistantMsg); err != nil {
		log.Printf("[ERROR] Failed to save assistant message: %v", err)
	}

	return &ChatResponse{
		UserMessage:      userMsg,
		AssistantMessage: assistantMsg,
		Actions:          actionsExecuted,
	}, nil
}

func (s *chatService) executeToolCall(userID uuid.UUID, toolCall groq.ToolCall) (ActionLog, string) {
	name := toolCall.Function.Name
	argsJSON := toolCall.Function.Arguments

	log.Printf("[INFO] Executing tool call '%s' with args: %s for user %s", name, argsJSON, userID)

	switch name {
	case "create_task":
		var args struct {
			Title       string  `json:"title"`
			Description *string `json:"description"`
			DueDate     string  `json:"due_date"`
			StartTime   *string `json:"start_time"`
			EndTime     *string `json:"end_time"`
			Priority    string  `json:"priority"`
		}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ActionLog{ToolName: name, Status: "failed", Message: err.Error()}, `{"status":"error","message":"invalid arguments"}`
		}

		// If start_time is provided but end_time is empty, default end_time to +1 hour for calendar/schedule view
		if args.StartTime != nil && *args.StartTime != "" && (args.EndTime == nil || *args.EndTime == "") {
			calcEndTime := s.calculateEndTime(*args.StartTime)
			args.EndTime = &calcEndTime
		}

		task, err := s.taskService.CreateTask(userID, &CreateTaskRequest{
			Title:       args.Title,
			Description: args.Description,
			DueDate:     args.DueDate,
			StartTime:   args.StartTime,
			EndTime:     args.EndTime,
			Priority:    models.TaskPriority(args.Priority),
		})
		if err != nil {
			return ActionLog{ToolName: name, Status: "failed", Message: err.Error()}, fmt.Sprintf(`{"status":"error","message":"%s"}`, err.Error())
		}

		// Auto-schedule a reminder if task has a start_time
		if task.StartTime != nil && *task.StartTime != "" {
			remindTime, err := s.calculateReminderTime(task.DueDate, *task.StartTime)
			if err == nil {
				reminderNote := fmt.Sprintf("Pengingat: Tugas '%s' dijadwalkan pada pukul %s", task.Title, *task.StartTime)
				_, remErr := s.reminderService.CreateReminder(userID, &CreateReminderRequest{
					TaskID:    task.ID,
					RemindAt:  remindTime,
					AIMessage: &reminderNote,
				})
				if remErr == nil {
					log.Printf("[INFO] Automatically created reminder for task '%s' at %s", task.Title, remindTime.Format("2006-01-02 15:04:05"))
				}
			}
		}

		resBytes, _ := json.Marshal(task)
		return ActionLog{
			ToolName: name,
			Status:   "success",
			Message:  fmt.Sprintf("Berhasil membuat tugas '%s'", task.Title),
			Data:     task,
		}, string(resBytes)

	case "get_tasks":
		var args struct {
			Status   string `json:"status"`
			Priority string `json:"priority"`
			DueDate  string `json:"due_date"`
		}
		_ = json.Unmarshal([]byte(argsJSON), &args)

		tasks, err := s.taskService.GetTasks(userID, args.Status, args.Priority, args.DueDate)
		if err != nil {
			return ActionLog{ToolName: name, Status: "failed", Message: err.Error()}, fmt.Sprintf(`{"status":"error","message":"%s"}`, err.Error())
		}

		resBytes, _ := json.Marshal(tasks)
		return ActionLog{
			ToolName: name,
			Status:   "success",
			Message:  fmt.Sprintf("Ditemukan %d tugas", len(tasks)),
			Data:     tasks,
		}, string(resBytes)

	case "update_task":
		var args struct {
			TaskID      string  `json:"task_id"`
			SearchTitle string  `json:"search_title"`
			Title       *string `json:"title"`
			Description *string `json:"description"`
			DueDate     *string `json:"due_date"`
			StartTime   *string `json:"start_time"`
			EndTime     *string `json:"end_time"`
			Status      *string `json:"status"`
			Priority    *string `json:"priority"`
		}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ActionLog{ToolName: name, Status: "failed", Message: err.Error()}, `{"status":"error","message":"invalid arguments"}`
		}

		targetTask, err := s.taskService.FindTaskByKeywordOrID(userID, args.TaskID, args.SearchTitle)
		if err != nil || targetTask == nil {
			return ActionLog{ToolName: name, Status: "failed", Message: "Tugas tidak ditemukan"}, `{"status":"error","message":"task not found"}`
		}

		var statusVal *models.TaskStatus
		if args.Status != nil {
			sVal := models.TaskStatus(*args.Status)
			statusVal = &sVal
		}
		var priorityVal *models.TaskPriority
		if args.Priority != nil {
			pVal := models.TaskPriority(*args.Priority)
			priorityVal = &pVal
		}

		updatedTask, err := s.taskService.UpdateTask(userID, targetTask.ID, &UpdateTaskRequest{
			Title:       args.Title,
			Description: args.Description,
			DueDate:     args.DueDate,
			StartTime:   args.StartTime,
			EndTime:     args.EndTime,
			Status:      statusVal,
			Priority:    priorityVal,
		})
		if err != nil {
			return ActionLog{ToolName: name, Status: "failed", Message: err.Error()}, fmt.Sprintf(`{"status":"error","message":"%s"}`, err.Error())
		}

		resBytes, _ := json.Marshal(updatedTask)
		return ActionLog{
			ToolName: name,
			Status:   "success",
			Message:  fmt.Sprintf("Berhasil memperbarui tugas '%s'", updatedTask.Title),
			Data:     updatedTask,
		}, string(resBytes)

	case "delete_task":
		var args struct {
			TaskID      string `json:"task_id"`
			SearchTitle string `json:"search_title"`
		}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ActionLog{ToolName: name, Status: "failed", Message: err.Error()}, `{"status":"error","message":"invalid arguments"}`
		}

		targetTask, err := s.taskService.FindTaskByKeywordOrID(userID, args.TaskID, args.SearchTitle)
		if err != nil || targetTask == nil {
			return ActionLog{ToolName: name, Status: "failed", Message: "Tugas tidak ditemukan"}, `{"status":"error","message":"task not found"}`
		}

		if err := s.taskService.DeleteTask(userID, targetTask.ID); err != nil {
			return ActionLog{ToolName: name, Status: "failed", Message: err.Error()}, fmt.Sprintf(`{"status":"error","message":"%s"}`, err.Error())
		}

		return ActionLog{
			ToolName: name,
			Status:   "success",
			Message:  fmt.Sprintf("Berhasil menghapus tugas '%s'", targetTask.Title),
			Data:     map[string]interface{}{"deleted_id": targetTask.ID, "title": targetTask.Title},
		}, fmt.Sprintf(`{"status":"success","deleted_task_id":"%s"}`, targetTask.ID)

	case "create_reminder":
		var args struct {
			TaskID      string  `json:"task_id"`
			SearchTitle string  `json:"search_title"`
			RemindAt    string  `json:"remind_at"`
			AIMessage   *string `json:"ai_message"`
		}
		if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
			return ActionLog{ToolName: name, Status: "failed", Message: err.Error()}, `{"status":"error","message":"invalid arguments"}`
		}

		targetTask, err := s.taskService.FindTaskByKeywordOrID(userID, args.TaskID, args.SearchTitle)
		if err != nil || targetTask == nil {
			return ActionLog{ToolName: name, Status: "failed", Message: "Tugas yang ingin diingatkan tidak ditemukan"}, `{"status":"error","message":"task not found"}`
		}

		remindTime, err := s.parseRemindAt(args.RemindAt)
		if err != nil {
			return ActionLog{ToolName: name, Status: "failed", Message: err.Error()}, fmt.Sprintf(`{"status":"error","message":"%s"}`, err.Error())
		}

		reminder, err := s.reminderService.CreateReminder(userID, &CreateReminderRequest{
			TaskID:    targetTask.ID,
			RemindAt:  remindTime,
			AIMessage: args.AIMessage,
		})
		if err != nil {
			return ActionLog{ToolName: name, Status: "failed", Message: err.Error()}, fmt.Sprintf(`{"status":"error","message":"%s"}`, err.Error())
		}

		resBytes, _ := json.Marshal(reminder)
		return ActionLog{
			ToolName: name,
			Status:   "success",
			Message:  fmt.Sprintf("Berhasil menyetel pengingat untuk tugas '%s' pada %s", targetTask.Title, remindTime.Format("02 Jan 2006 15:04")),
			Data:     reminder,
		}, string(resBytes)

	default:
		return ActionLog{ToolName: name, Status: "failed", Message: "Unknown tool"}, `{"status":"error","message":"unknown tool"}`
	}
}

func (s *chatService) parseRemindAt(remindStr string) (time.Time, error) {
	loc, err := time.LoadLocation(s.timezone)
	if err != nil {
		loc = time.Local
	}

	layouts := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02 15:04",
		"2006-01-02T15:04",
	}

	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, remindStr, loc); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("format waktu pengingat '%s' tidak valid", remindStr)
}

func (s *chatService) GetHistory(userID uuid.UUID, limit int) ([]models.ChatMessage, error) {
	if limit <= 0 {
		limit = 50 // default to last 50 messages
	}
	return s.chatRepo.FindByUserID(userID, limit)
}

func (s *chatService) calculateEndTime(startTime string) string {
	layouts := []string{"15:04:05", "15:04", "3:04 PM", "3:04PM"}
	for _, l := range layouts {
		if t, err := time.Parse(l, startTime); err == nil {
			return t.Add(1 * time.Hour).Format("15:04:05")
		}
	}
	return startTime
}

func (s *chatService) calculateReminderTime(dueDate, startTime string) (time.Time, error) {
	loc, err := time.LoadLocation(s.timezone)
	if err != nil {
		loc = time.Local
	}

	layout := "2006-01-02 15:04:05"
	dtStr := fmt.Sprintf("%s %s", dueDate, startTime)
	eventTime, err := time.ParseInLocation(layout, dtStr, loc)
	if err != nil {
		eventTime, err = time.ParseInLocation("2006-01-02 15:04", dtStr, loc)
		if err != nil {
			return time.Time{}, err
		}
	}

	remindAt := eventTime.Add(-15 * time.Minute)
	if time.Now().After(remindAt) {
		remindAt = eventTime
	}
	return remindAt, nil
}

func (s *chatService) ClearHistory(userID uuid.UUID) error {
	return s.chatRepo.DeleteByUserID(userID)
}

func cleanChatText(text string) string {
	// Strip bold, italic, and code markdown symbols
	text = strings.ReplaceAll(text, "**", "")
	text = strings.ReplaceAll(text, "__", "")
	text = strings.ReplaceAll(text, "```", "")
	text = strings.ReplaceAll(text, "`", "")

	lines := strings.Split(text, "\n")
	var cleanedLines []string
	itemCounter := 1

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			cleanedLines = append(cleanedLines, "")
			continue
		}

		// Check if line is a markdown table row (starts and ends with |)
		if strings.HasPrefix(trimmed, "|") && strings.HasSuffix(trimmed, "|") {
			inner := strings.Trim(trimmed, "|")

			// Skip table separator line like |---|---| or |:---|:---|
			isSep := true
			for _, r := range inner {
				if r != '-' && r != ':' && r != '|' && r != ' ' {
					isSep = false
					break
				}
			}
			if isSep {
				continue
			}

			// Skip header rows like | No | Judul Tugas | Waktu | ...
			lower := strings.ToLower(inner)
			if strings.Contains(lower, "judul") || (strings.Contains(lower, "tugas") && strings.Contains(lower, "status")) || strings.Contains(lower, "prioritas") {
				continue
			}

			// Parse row columns into clean list items
			cols := strings.Split(inner, "|")
			var validCols []string
			for _, col := range cols {
				c := strings.TrimSpace(col)
				if c != "" {
					validCols = append(validCols, c)
				}
			}
			if len(validCols) > 0 {
				// If the first col is just a number (e.g. "1"), remove it to avoid double numbering
				if len(validCols) > 1 && len(validCols[0]) <= 3 {
					isNum := true
					for _, ch := range validCols[0] {
						if ch < '0' || ch > '9' {
							isNum = false
							break
						}
					}
					if isNum {
						validCols = validCols[1:]
					}
				}
				rowText := fmt.Sprintf("%d. %s", itemCounter, strings.Join(validCols, " - "))
				itemCounter++
				cleanedLines = append(cleanedLines, rowText)
				continue
			}
		}

		// Remove header hashes (# or ## or ###)
		if strings.HasPrefix(trimmed, "#") {
			trimmed = strings.TrimLeft(trimmed, "# ")
			line = trimmed
		}

		cleanedLines = append(cleanedLines, line)
	}

	result := strings.Join(cleanedLines, "\n")
	result = strings.ReplaceAll(result, "*", "")
	return strings.TrimSpace(result)
}

