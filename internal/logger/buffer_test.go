package logger

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestNewLogBuffer(t *testing.T) {
	buffer := NewLogBuffer(100)

	if buffer == nil {
		t.Fatal("Expected non-nil buffer")
	}

	if buffer.Size() != 0 {
		t.Errorf("Expected size 0, got %d", buffer.Size())
	}
}

func TestLogBuffer_Add(t *testing.T) {
	buffer := NewLogBuffer(5)

	entry := LogEntry{
		Time:    time.Now(),
		Level:   "info",
		Message: "test message",
		Attributes: map[string]string{
			"key":       "value",
			"user":      "testuser",
			"operation": "test_op",
			"tool":      "test_tool",
		},
	}

	buffer.Add(entry)

	if buffer.Size() != 1 {
		t.Errorf("Expected size 1, got %d", buffer.Size())
	}
}

func TestLogBuffer_CircularOverwrite(t *testing.T) {
	buffer := NewLogBuffer(3)

	// Add 5 entries to a buffer of size 3
	for i := range 5 {
		buffer.Add(LogEntry{
			Time:    time.Now(),
			Level:   "info",
			Message: "message " + string(rune('0'+i)),
		})
	}

	// Should only have last 3 entries
	if buffer.Size() != 3 {
		t.Errorf("Expected size 3, got %d", buffer.Size())
	}

	// Query all to verify circular buffer behavior
	// After adding 5 to buffer of 3: [message 3, message 4, message 2]
	// After reverse: [message 2, message 4, message 3]
	entries := buffer.Query(LogFilter{})
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// Verify we have the expected messages (circular overwrite + reverse)
	messages := make(map[string]bool)
	for _, e := range entries {
		messages[e.Message] = true
	}

	expectedMessages := []string{"message 2", "message 3", "message 4"}
	for _, expected := range expectedMessages {
		if !messages[expected] {
			t.Errorf("Expected to find '%s' in results", expected)
		}
	}

	// Oldest entry (message 0, 1) should not be present
	if messages["message 0"] || messages["message 1"] {
		t.Error("Old messages should have been overwritten")
	}
}

func TestLogBuffer_Query_All(t *testing.T) {
	buffer := NewLogBuffer(10)

	// Add multiple entries
	for range 5 {
		buffer.Add(LogEntry{
			Time:    time.Now(),
			Level:   "info",
			Message: "message",
		})
	}

	entries := buffer.Query(LogFilter{})

	if len(entries) != 5 {
		t.Errorf("Expected 5 entries, got %d", len(entries))
	}
}

func TestLogBuffer_Query_LevelFilter(t *testing.T) {
	buffer := NewLogBuffer(10)

	buffer.Add(LogEntry{Level: "debug", Message: "debug msg"})
	buffer.Add(LogEntry{Level: "info", Message: "info msg"})
	buffer.Add(LogEntry{Level: "warn", Message: "warn msg"})
	buffer.Add(LogEntry{Level: "error", Message: "error msg"})

	// Filter by level
	entries := buffer.Query(LogFilter{Level: "error"})

	if len(entries) != 1 {
		t.Errorf("Expected 1 error entry, got %d", len(entries))
	}

	if entries[0].Level != "error" {
		t.Errorf("Expected level 'error', got '%s'", entries[0].Level)
	}
}

func TestLogBuffer_Query_DateFilter(t *testing.T) {
	buffer := NewLogBuffer(10)

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)
	tomorrow := now.Add(24 * time.Hour)

	buffer.Add(LogEntry{
		Time:    yesterday,
		Message: "old message",
	})
	buffer.Add(LogEntry{
		Time:    now,
		Message: "recent message",
	})

	// Filter: after yesterday
	entries := buffer.Query(LogFilter{
		DateFrom: yesterday.Add(1 * time.Hour),
	})

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry after yesterday, got %d", len(entries))
	}

	// Filter: before tomorrow
	entries = buffer.Query(LogFilter{
		DateTo: tomorrow,
	})

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries before tomorrow, got %d", len(entries))
	}
}

func TestLogBuffer_Query_KeywordFilter(t *testing.T) {
	buffer := NewLogBuffer(10)

	buffer.Add(LogEntry{Message: "this is a test message"})
	buffer.Add(LogEntry{Message: "another different message"})
	buffer.Add(LogEntry{Message: "test again"})

	entries := buffer.Query(LogFilter{Keyword: "test"})

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries with 'test', got %d", len(entries))
	}
}

func TestLogBuffer_Query_UserFilter(t *testing.T) {
	buffer := NewLogBuffer(10)

	buffer.Add(LogEntry{Message: "alice's message", Attributes: map[string]string{"user": "alice"}})
	buffer.Add(LogEntry{Message: "bob's message", Attributes: map[string]string{"user": "bob"}})
	buffer.Add(LogEntry{Message: "alice again", Attributes: map[string]string{"user": "alice"}})

	entries := buffer.Query(LogFilter{User: "alice"})

	if len(entries) != 2 {
		t.Errorf("Expected 2 entries from alice, got %d", len(entries))
	}
}

func TestLogBuffer_Query_OperationFilter(t *testing.T) {
	buffer := NewLogBuffer(10)

	buffer.Add(LogEntry{Message: "msg1", Attributes: map[string]string{"operation": "create_persona"}})
	buffer.Add(LogEntry{Message: "msg2", Attributes: map[string]string{"operation": "update_element"}})
	buffer.Add(LogEntry{Message: "msg3", Attributes: map[string]string{"operation": "create_persona"}})

	entries := buffer.Query(LogFilter{Operation: "create_persona"})

	if len(entries) != 2 {
		t.Errorf("Expected 2 create_persona entries, got %d", len(entries))
	}
}

func TestLogBuffer_Query_ToolFilter(t *testing.T) {
	buffer := NewLogBuffer(10)

	buffer.Add(LogEntry{Message: "msg1", Attributes: map[string]string{"tool": "backup_portfolio"}})
	buffer.Add(LogEntry{Message: "msg2", Attributes: map[string]string{"tool": "restore_portfolio"}})
	buffer.Add(LogEntry{Message: "msg3", Attributes: map[string]string{"tool": "backup_portfolio"}})

	entries := buffer.Query(LogFilter{Tool: "backup_portfolio"})

	if len(entries) != 2 {
		t.Errorf("Expected 2 backup_portfolio entries, got %d", len(entries))
	}
}

func TestLogBuffer_Query_Limit(t *testing.T) {
	buffer := NewLogBuffer(10)

	for range 8 {
		buffer.Add(LogEntry{Message: "message"})
	}

	entries := buffer.Query(LogFilter{Limit: 3})

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries (limit), got %d", len(entries))
	}
}

func TestLogBuffer_Query_CombinedFilters(t *testing.T) {
	buffer := NewLogBuffer(20)

	now := time.Now()

	buffer.Add(LogEntry{
		Time:       now,
		Level:      "error",
		Message:    "error occurred",
		Attributes: map[string]string{"user": "alice"},
	})
	buffer.Add(LogEntry{
		Time:       now,
		Level:      "info",
		Message:    "info message",
		Attributes: map[string]string{"user": "alice"},
	})
	buffer.Add(LogEntry{
		Time:       now.Add(-1 * time.Hour),
		Level:      "error",
		Message:    "old error",
		Attributes: map[string]string{"user": "bob"},
	})

	// Filter: level=error AND user=alice
	entries := buffer.Query(LogFilter{
		Level: "error",
		User:  "alice",
	})

	if len(entries) != 1 {
		t.Errorf("Expected 1 entry (error + alice), got %d", len(entries))
	}

	if entries[0].Message != "error occurred" {
		t.Errorf("Expected 'error occurred', got '%s'", entries[0].Message)
	}
}

func TestLogBuffer_Clear(t *testing.T) {
	buffer := NewLogBuffer(10)

	for range 5 {
		buffer.Add(LogEntry{Message: "message"})
	}

	if buffer.Size() != 5 {
		t.Errorf("Expected size 5 before clear, got %d", buffer.Size())
	}

	buffer.Clear()

	if buffer.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", buffer.Size())
	}

	entries := buffer.Query(LogFilter{})
	if len(entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(entries))
	}
}

func TestBufferedHandler_Handle(t *testing.T) {
	buffer := NewLogBuffer(10)
	var buf bytes.Buffer
	handler := NewBufferedHandler(slog.NewJSONHandler(&buf, nil), buffer)

	ctx := context.Background()

	record := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	record.AddAttrs(
		slog.String("key", "value"),
		slog.String("user", "testuser"),
		slog.String("operation", "test_operation"),
	)

	err := handler.Handle(ctx, record)
	if err != nil {
		t.Fatalf("Handle failed: %v", err)
	}

	if buffer.Size() != 1 {
		t.Errorf("Expected 1 entry in buffer, got %d", buffer.Size())
	}

	entries := buffer.Query(LogFilter{})
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(entries))
	}

	entry := entries[0]
	if entry.Level != "INFO" {
		t.Errorf("Expected level INFO, got %s", entry.Level)
	}
	if entry.Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", entry.Message)
	}

	// Check attributes
	if entry.Attributes["key"] != "value" {
		t.Errorf("Expected attribute key='value', got %v", entry.Attributes["key"])
	}
	if entry.Attributes["user"] != "testuser" {
		t.Errorf("Expected attribute user='testuser', got %v", entry.Attributes["user"])
	}
	if entry.Attributes["operation"] != "test_operation" {
		t.Errorf("Expected attribute operation='test_operation', got %v", entry.Attributes["operation"])
	}
}

func TestBufferedHandler_Enabled(t *testing.T) {
	buffer := NewLogBuffer(10)
	var buf bytes.Buffer
	baseHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})
	handler := NewBufferedHandler(baseHandler, buffer)

	ctx := context.Background()

	// Should respect base handler's level
	if !handler.Enabled(ctx, slog.LevelError) {
		t.Error("Expected ERROR level to be enabled")
	}

	if handler.Enabled(ctx, slog.LevelDebug) {
		t.Error("Expected DEBUG level to be disabled (base handler is WARN)")
	}
}

func TestBufferedHandler_WithAttrs(t *testing.T) {
	buffer := NewLogBuffer(10)
	var buf bytes.Buffer
	handler := NewBufferedHandler(slog.NewJSONHandler(&buf, nil), buffer)

	newHandler := handler.WithAttrs([]slog.Attr{
		slog.String("component", "test"),
	})

	if newHandler == nil {
		t.Error("Expected non-nil handler from WithAttrs")
	}

	// Should be a buffered handler
	_, ok := newHandler.(*BufferedHandler)
	if !ok {
		t.Error("Expected WithAttrs to return *BufferedHandler")
	}
}

func TestBufferedHandler_WithGroup(t *testing.T) {
	buffer := NewLogBuffer(10)
	var buf bytes.Buffer
	handler := NewBufferedHandler(slog.NewJSONHandler(&buf, nil), buffer)

	newHandler := handler.WithGroup("testgroup")

	if newHandler == nil {
		t.Error("Expected non-nil handler from WithGroup")
	}

	// Should be a buffered handler
	_, ok := newHandler.(*BufferedHandler)
	if !ok {
		t.Error("Expected WithGroup to return *BufferedHandler")
	}
}

func TestInitWithBuffer(t *testing.T) {
	cfg := &Config{
		Level:     slog.LevelInfo,
		Format:    "json",
		AddSource: false,
	}

	InitWithBuffer(cfg, 50)

	logger := Get()
	if logger == nil {
		t.Fatal("Expected non-nil logger after InitWithBuffer")
	}

	buffer := GetLogBuffer()
	if buffer == nil {
		t.Fatal("Expected non-nil log buffer after InitWithBuffer")
	}

	// Test that logging actually populates the buffer
	logger.Info("test message", "key", "value")

	if buffer.Size() != 1 {
		t.Errorf("Expected buffer size 1, got %d", buffer.Size())
	}

	entries := buffer.Query(LogFilter{})
	if len(entries) != 1 {
		t.Fatalf("Expected 1 entry in buffer, got %d", len(entries))
	}

	if entries[0].Message != "test message" {
		t.Errorf("Expected message 'test message', got '%s'", entries[0].Message)
	}
}

func TestGetLogBuffer_NotInitialized(t *testing.T) {
	// Reset global buffer
	globalLogBuffer = nil

	buffer := GetLogBuffer()
	if buffer != nil {
		t.Error("Expected nil buffer when not initialized with InitWithBuffer")
	}
}

func TestParseLevelValue(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"DEBUG", 0},
		{"debug", 0},
		{"INFO", 1},
		{"info", 1},
		{"WARN", 2},
		{"warn", 2},
		{"ERROR", 3},
		{"error", 3},
		{"unknown", 1}, // Default to INFO
	}

	for _, tt := range tests {
		result := parseLevelValue(tt.input)
		if result != tt.expected {
			t.Errorf("parseLevelValue(%q) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		str      string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "WORLD", true}, // Case insensitive
		{"hello world", "foo", false},
		{"", "", true},
		{"test", "", true},
	}

	for _, tt := range tests {
		result := contains(tt.str, tt.substr)
		if result != tt.expected {
			t.Errorf("contains(%q, %q) = %v, expected %v", tt.str, tt.substr, result, tt.expected)
		}
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"HELLO", "hello"},
		{"MixedCase", "mixedcase"},
		{"lowercase", "lowercase"},
		{"", ""},
		{"123ABC", "123abc"},
	}

	for _, tt := range tests {
		result := toLower(tt.input)
		if result != tt.expected {
			t.Errorf("toLower(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestLogBuffer_Query_InvalidDates(t *testing.T) {
	buffer := NewLogBuffer(10)

	buffer.Add(LogEntry{
		Time:    time.Now(),
		Message: "test",
	})

	// Zero date should match all entries
	entries := buffer.Query(LogFilter{
		DateFrom: time.Time{}, // Zero time
	})

	// Should return all entries (filter ignored)
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry with invalid date filter, got %d", len(entries))
	}
}

func TestLogBuffer_Concurrency(t *testing.T) {
	buffer := NewLogBuffer(100)

	// Test concurrent writes
	done := make(chan bool)
	for i := range 10 {
		go func(id int) {
			for range 10 {
				buffer.Add(LogEntry{
					Message:    "concurrent message",
					Attributes: map[string]string{"user": "user" + string(rune('0'+id))},
				})
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}

	// Should have 100 entries (or maxSize if exceeded)
	size := buffer.Size()
	if size != 100 {
		t.Errorf("Expected size 100, got %d", size)
	}

	// Test concurrent reads
	for range 5 {
		go func() {
			entries := buffer.Query(LogFilter{})
			if len(entries) == 0 {
				t.Error("Expected non-zero entries in concurrent read")
			}
			done <- true
		}()
	}

	for range 5 {
		<-done
	}
}

func TestLogBuffer_SortOrder(t *testing.T) {
	buffer := NewLogBuffer(10)

	// Add entries in insertion order
	now := time.Now()
	buffer.Add(LogEntry{Time: now.Add(-3 * time.Second), Message: "first added"})
	buffer.Add(LogEntry{Time: now, Message: "second added"})
	buffer.Add(LogEntry{Time: now.Add(-1 * time.Second), Message: "third added"})

	entries := buffer.Query(LogFilter{})

	// Query reverses insertion order, so newest insertion is first
	if len(entries) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(entries))
	}

	// Results are in reverse insertion order (newest insertion first)
	if entries[0].Message != "third added" {
		t.Errorf("Expected first result to be 'third added', got '%s'", entries[0].Message)
	}
	if entries[1].Message != "second added" {
		t.Errorf("Expected second result to be 'second added', got '%s'", entries[1].Message)
	}
	if entries[2].Message != "first added" {
		t.Errorf("Expected third result to be 'first added', got '%s'", entries[2].Message)
	}
}
