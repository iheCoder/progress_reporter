package progress_reporter

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ProgressBar 用于跟踪和显示任务的完成进度。
type ProgressBar struct {
	total            int        // 总工作单元数
	current          int        // 当前已完成的工作单元数
	barLength        int        // 进度条在控制台中显示的长度
	startTime        time.Time  // 进度条开始的时间
	description      string     // 进度条的描述文字
	currentStageName string     // 当前阶段名称，用于更细致的进度展示
	mu               sync.Mutex // 用于保护并发访问
}

// NewProgressBar 创建一个新的 ProgressBar 实例。
// description: 进度条的描述。
// total: 总工作单元数。
// barLength: 进度条在控制台显示的字符长度。
func NewProgressBar(description string, total int, barLength int) *ProgressBar {
	return &ProgressBar{
		total:            total,
		current:          0,
		barLength:        barLength,
		startTime:        time.Now(),
		description:      description,
		currentStageName: "",
		mu:               sync.Mutex{},
	}
}

// Increment 使已完成的工作单元数增加1，并刷新进度条显示。
func (pb *ProgressBar) Increment() {
	pb.IncrementBy(1)
}

// IncrementBy 使已完成的工作单元数增加指定数量，并刷新进度条显示。
// n: 增加的数量。
func (pb *ProgressBar) IncrementBy(n int) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if n < 0 {
		fmt.Println("Error: Increment value cannot be negative.")
		return
	}
	pb.current += n
	if pb.current > pb.total {
		pb.current = pb.total // 防止当前进度超过总数
	}

	pb.displayWithoutLock(pb.current, pb.total, pb.description, pb.barLength,
		pb.startTime, pb.currentStageName) // 使用不带锁的显示方法
}

// AddTotal 动态增加总工作单元数。
// n: 要增加到总数上的值。如果为负数，则总数会减少。
func (pb *ProgressBar) AddTotal(n int) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.total += n
	if pb.total < 0 {
		pb.total = 0 // 总数不能为负
	}

	// 如果当前进度超过了新的总数（例如，总数被减少了），则调整当前进度
	if pb.current > pb.total {
		pb.current = pb.total
	}

	pb.displayWithoutLock(pb.current, pb.total, pb.description, pb.barLength,
		pb.startTime, pb.currentStageName) // 使用不带锁的显示方法
}

// SetCurrentStage 设置当前正在进行的阶段名称。
// name: 阶段的名称。
func (pb *ProgressBar) SetCurrentStage(name string) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.currentStageName = name
	pb.displayWithoutLock(pb.current, pb.total, pb.description, pb.barLength,
		pb.startTime, pb.currentStageName) // 使用不带锁的显示方法
}

// Display 在控制台中打印当前的进度条状态。
// 输出格式示例:
// My Task: [=====>--------------------] 25% (5/20) | Stage: Processing | Elapsed: 5s | Avg: 0.5 items/s | ETA: 30s
func (pb *ProgressBar) Display() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	// 使用公共的格式化函数
	output := pb.formatProgressBar(pb.current, pb.total, pb.description, pb.barLength,
		pb.startTime, pb.currentStageName)
	fmt.Print(output)

	if pb.current == pb.total {
		fmt.Println() // 完成后换行
	}
}

// displayWithoutLock 在控制台中打印当前的进度条状态（不带锁）。
func (pb *ProgressBar) displayWithoutLock(current, total int, description string, barLength int, startTime time.Time, currentStageName string) {
	// 使用公共的格式化函数
	output := pb.formatProgressBar(current, total, description, barLength, startTime, currentStageName)
	fmt.Print(output)

	if current == total {
		fmt.Println() // 完成后换行
	}
}

// formatProgressBar 根据给定的参数格式化进度条字符串
// 返回一个已格式化的字符串，包含进度条、百分比、计数等信息
func (pb *ProgressBar) formatProgressBar(current, total int, description string, barLength int,
	startTime time.Time, currentStageName string) string {
	elapsedTime := time.Since(startTime)
	elapsedSeconds := elapsedTime.Seconds()

	avgSpeedString := "0.0 items/s"
	var avgSpeed float64
	if elapsedSeconds > 0 && current > 0 {
		avgSpeed = float64(current) / elapsedSeconds
		avgSpeedString = fmt.Sprintf("%.1f items/s", avgSpeed)
	} else if current == 0 && elapsedSeconds > 0 {
		avgSpeedString = "0.0 items/s"
	}

	etaString := "N/A"
	if current == total {
		etaString = "Done"
	} else if avgSpeed > 0 {
		remainingItems := total - current
		etaSeconds := float64(remainingItems) / avgSpeed
		etaString = (time.Duration(etaSeconds*1000) * time.Millisecond).Round(time.Second).String()
	} else if current == 0 && total > 0 {
		etaString = "Estimating..."
	}

	if total == 0 { // 防止除以零
		return fmt.Sprintf("\r%s: [ %s ] %d%% (%d/%d) | Stage: %s | Elapsed: %s | Avg: %s | ETA: %s",
			description,
			strings.Repeat("-", barLength),
			0,
			current,
			total,
			currentStageName,
			elapsedTime.Round(time.Second).String(),
			avgSpeedString,
			etaString)
	}

	percent := float64(current) / float64(total)
	filledLength := int(float64(barLength) * percent)
	bar := strings.Repeat("=", filledLength) + strings.Repeat("-", barLength-filledLength)

	// 使用 \r 回车符将光标移到行首，实现动态更新效果
	return fmt.Sprintf("\r%s: [%s] %3.0f%% (%d/%d) | Stage: %s | Elapsed: %s | Avg: %s | ETA: %s",
		description,
		bar,
		percent*100,
		current,
		total,
		currentStageName,
		elapsedTime.Round(time.Second).String(),
		avgSpeedString,
		etaString)
}

// Finish 标记进度条完成，并打印最终状态。
func (pb *ProgressBar) Finish() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current = pb.total // 确保进度为100%
	pb.currentStageName = "完成"
	pb.displayWithoutLock(pb.current, pb.total, pb.description, pb.barLength,
		pb.startTime, pb.currentStageName) // 使用不带锁的显示方法
	fmt.Println() // 确保在完成后换行
}
