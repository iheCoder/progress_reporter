package progress_reporter

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewProgressBar(t *testing.T) {
	desc := "Test Description"
	total := 100
	barLength := 50
	pb := NewProgressBar(desc, total, barLength)

	if pb.description != desc {
		t.Errorf("Expected description %s, got %s", desc, pb.description)
	}
	if pb.total != total {
		t.Errorf("Expected total %d, got %d", total, pb.total)
	}
	if pb.barLength != barLength {
		t.Errorf("Expected barLength %d, got %d", barLength, pb.barLength)
	}
	if pb.current != 0 {
		t.Errorf("Expected current 0, got %d", pb.current)
	}
	if pb.currentStageName != "" {
		t.Errorf("Expected currentStageName empty, got %s", pb.currentStageName)
	}
}

func TestIncrementAndIncrementBy(t *testing.T) {
	pb := NewProgressBar("Test Increment", 10, 10)

	pb.Increment()
	if pb.current != 1 {
		t.Errorf("Expected current 1 after Increment, got %d", pb.current)
	}

	pb.IncrementBy(4)
	if pb.current != 5 {
		t.Errorf("Expected current 5 after IncrementBy(4), got %d", pb.current)
	}

	// Test incrementing beyond total
	pb.IncrementBy(10) // current is 5, total is 10. 5 + 10 = 15, should be capped at 10
	if pb.current != 10 {
		t.Errorf("Expected current to be capped at total %d, got %d", pb.total, pb.current)
	}

	// Test negative increment (should be ignored by current logic, but good to be aware)
	// Current IncrementBy prints an error and returns, so current should not change.
	oldCurrent := pb.current
	pb.IncrementBy(-1)
	if pb.current != oldCurrent {
		t.Errorf("Expected current %d to remain unchanged after negative IncrementBy, got %d", oldCurrent, pb.current)
	}
}

func TestAddTotal(t *testing.T) {
	pb := NewProgressBar("Test AddTotal", 10, 10)

	pb.AddTotal(5)
	if pb.total != 15 {
		t.Errorf("Expected total 15 after AddTotal(5), got %d", pb.total)
	}

	pb.IncrementBy(12)
	if pb.current != 12 {
		t.Errorf("Expected current 12, got %d", pb.current)
	}

	// Test reducing total below current
	pb.AddTotal(-8) // total becomes 15 - 8 = 7. current is 12, should be adjusted to 7
	if pb.total != 7 {
		t.Errorf("Expected total 7 after AddTotal(-8), got %d", pb.total)
	}
	if pb.current != 7 {
		t.Errorf("Expected current to be adjusted to 7 when total reduced, got %d", pb.current)
	}

	// Test reducing total to negative (should be 0)
	pb.AddTotal(-100)
	if pb.total != 0 {
		t.Errorf("Expected total 0 when reduced below zero, got %d", pb.total)
	}
	if pb.current != 0 {
		t.Errorf("Expected current 0 when total is 0, got %d", pb.current)
	}
}

func TestSetCurrentStage(t *testing.T) {
	pb := NewProgressBar("Test Stage", 10, 10)
	stageName := "Processing Data"
	pb.SetCurrentStage(stageName)
	if pb.currentStageName != stageName {
		t.Errorf("Expected currentStageName '%s', got '%s'", stageName, pb.currentStageName)
	}
}

func TestFinish(t *testing.T) {
	pb := NewProgressBar("Test Finish", 10, 10)
	pb.IncrementBy(5)
	pb.Finish()

	if pb.current != pb.total {
		t.Errorf("Expected current to be equal to total on Finish, got current %d, total %d", pb.current, pb.total)
	}
	if pb.currentStageName != "完成" {
		t.Errorf("Expected currentStageName '完成' on Finish, got '%s'", pb.currentStageName)
	}
}

func TestDisplayOutput(t *testing.T) {
	pb := NewProgressBar("Display Test", 100, 20)
	pb.IncrementBy(25)
	pb.SetCurrentStage("Testing Stage")

	// Capture output
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	pb.Display()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Basic checks for content - exact format is tricky due to \r and timing
	if !strings.Contains(output, "Display Test") {
		t.Errorf("Display output does not contain description: %s", output)
	}
	if !strings.Contains(output, "25%") {
		t.Errorf("Display output does not contain correct percentage: %s", output)
	}
	if !strings.Contains(output, "(25/100)") {
		t.Errorf("Display output does not contain correct count: %s", output)
	}
	if !strings.Contains(output, "Testing Stage") {
		t.Errorf("Display output does not contain stage name: %s", output)
	}
	if !strings.Contains(output, "ETA") {
		t.Errorf("Display output does not contain ETA: %s", output)
	}
	if !strings.Contains(output, "Avg") {
		t.Errorf("Display output does not contain Avg: %s", output)
	}
}

func TestProgressBarConcurrency(t *testing.T) {
	pb := NewProgressBar("Concurrent Test", 1000, 50)
	numGoroutines := 100
	incrementsPerGoroutine := 10

	var wg sync.WaitGroup
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				pb.Increment()
				time.Sleep(time.Millisecond) // Small sleep to increase chance of race if not protected
			}
		}()
	}

	wg.Wait()

	expectedCurrent := numGoroutines * incrementsPerGoroutine
	if pb.current != expectedCurrent {
		t.Errorf("Expected current %d after concurrent increments, got %d", expectedCurrent, pb.current)
	}

	// 测试并发的 AddTotal 和 SetCurrentStage
	pb.AddTotal(200) // 总数变为 1200
	pb.SetCurrentStage("并发阶段变更")

	for i := 0; i < numGoroutines/2; i++ {
		wg.Add(1)
		go func(k int) {
			defer wg.Done()
			pb.AddTotal(1)                              // 每个 goroutine 给总数加1
			pb.SetCurrentStage(fmt.Sprintf("阶段 %d", k)) // 并发设置阶段名称
			pb.IncrementBy(1)                           // 每个 goroutine 给当前进度加1
			time.Sleep(time.Millisecond)                // 增加竞争条件的可能性
		}(i)
	}
	wg.Wait()

	expectedTotal := 1000 + 200 + (numGoroutines / 2)
	if pb.total != expectedTotal {
		t.Errorf("并发 AddTotal 后，期望总数为 %d，得到 %d", expectedTotal, pb.total)
	}

	expectedCurrentAfterMoreIncrements := expectedCurrent + (numGoroutines / 2)
	if pb.current != expectedCurrentAfterMoreIncrements {
		t.Errorf("并发 AddTotal 和 IncrementBy 后，期望当前进度为 %d，得到 %d",
			expectedCurrentAfterMoreIncrements, pb.current)
	}

	// 阶段名称会是并发设置的其中一个，很难预测是哪一个
	// 我们只需要检查它不是在并发 SetCurrentStage 调用之前的那个值
	if pb.currentStageName == "并发阶段变更" && numGoroutines/2 > 0 {
		t.Errorf("期望阶段名称被并发调用更新，但仍为 '%s'", pb.currentStageName)
	}

	// 测试完成状态
	pb.Finish()
	if pb.current != pb.total {
		t.Errorf("调用 Finish 后，期望当前进度 %d 等于总数 %d，得到 %d", pb.total, pb.total, pb.current)
	}
	if pb.currentStageName != "完成" {
		t.Errorf("调用 Finish 后，期望阶段名称为 '完成'，得到 '%s'", pb.currentStageName)
	}
}
