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
	pb.Display()
}

// SetCurrentStage 设置当前正在进行的阶段名称。
// name: 阶段的名称。
func (pb *ProgressBar) SetCurrentStage(name string) {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.currentStageName = name
	pb.Display() // 更新阶段名称后也刷新显示
}

// Display 在控制台中打印当前的进度条状态。
// 输出格式示例:
// My Task: [=====>--------------------] 25% (5/20) | Stage: Processing | Elapsed: 5s
func (pb *ProgressBar) Display() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	if pb.total == 0 { // 防止除以零
		fmt.Printf("\r%s: [ %s ] %d%% (%d/%d) | Stage: %s | Elapsed: %s",
			pb.description,
			strings.Repeat("-", pb.barLength),
			0,
			pb.current,
			pb.total,
			pb.currentStageName,
			time.Since(pb.startTime).Round(time.Second))
		return
	}

	percent := float64(pb.current) / float64(pb.total)
	filledLength := int(float64(pb.barLength) * percent)
	bar := strings.Repeat("=", filledLength) + strings.Repeat("-", pb.barLength-filledLength)

	// 使用 \r 回车符将光标移到行首，实现动态更新效果
	fmt.Printf("\r%s: [%s] %3.0f%% (%d/%d) | Stage: %s | Elapsed: %s",
		pb.description,
		bar,
		percent*100,
		pb.current,
		pb.total,
		pb.currentStageName,
		time.Since(pb.startTime).Round(time.Second))

	if pb.current == pb.total {
		fmt.Println() // 完成后换行
	}
}

// Finish 标记进度条完成，并打印最终状态。
func (pb *ProgressBar) Finish() {
	pb.mu.Lock()
	defer pb.mu.Unlock()

	pb.current = pb.total // 确保进度为100%
	pb.currentStageName = "完成"
	pb.Display()
	fmt.Println() // 确保在完成后换行
}
