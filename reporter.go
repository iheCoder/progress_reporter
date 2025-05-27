package progress_reporter

import (
	"fmt"
	"time"
)

// ProgressReporter 用于跟踪和报告代码段的执行时间。
// 可用于性能分析，识别代码中的瓶颈。
type ProgressReporter struct {
	startTime, endTime time.Time            // 记录整体开始和结束时间
	totalDuration      time.Duration        // 记录整体持续时间
	ks                 map[string]*keyStage // 存储不同关键阶段的耗时信息
}

// NewProgressReporter 创建一个新的 ProgressReporter 实例。
// 使用场景：在开始性能分析或耗时跟踪之前，首先调用此方法获取一个 ProgressReporter 对象。
// 返回一个指向 ProgressReporter 的指针。
func NewProgressReporter() *ProgressReporter {
	return &ProgressReporter{
		ks: make(map[string]*keyStage),
	}
}

// StartRecord 记录整体操作的开始时间。
// 应在需要跟踪时间的代码段的开头调用。
// 使用场景：例如，在一个复杂函数的入口处调用 StartRecord，以开始计时整个函数的执行。
func (p *ProgressReporter) StartRecord() {
	p.startTime = time.Now()
}

// EndRecord 记录整体操作的结束时间并计算总持续时间。
// 应在需要跟踪时间的代码段的末尾调用。
// 使用场景：与 StartRecord 配对使用，在对应复杂函数的出口处调用 EndRecord，以结束计时并计算总耗时。
func (p *ProgressReporter) EndRecord() {
	p.endTime = time.Now()
	p.totalDuration = p.endTime.Sub(p.startTime)
}

// StartKeyStageRecord 记录特定关键阶段的开始时间。
// 如果同名关键阶段已存在，则更新其开始��间。
// 如果是新的关键阶段，则初始化该阶段的信息。
// 使用场景：在一个长任务内部，标记某个具体子步骤的开始。例如，在数据处理流程中，标记“数据加载”阶段的开始。
// name: 关键阶段的名称。
func (p *ProgressReporter) StartKeyStageRecord(name string) {
	if _, ok := p.ks[name]; !ok {
		p.ks[name] = &keyStage{
			name:        name,
			minDuration: time.Duration(1<<63 - 1), // 初始化为最大可能持续时间
		}
	}

	p.ks[name].startTime = time.Now()
}

// EndKeyStageRecord 记录特定关键阶段的结束时间并更新其统计信息。
// 计算当前阶段的持续时间，并更新总持续时间、计数、最大和最小持续时间。
// 使用场景：与 StartKeyStageRecord 配对使用，标记某个具体子步骤的结束。例如，标记“数据加载”阶段的结束，并记录其耗时。
// name: 关键阶段的名称。
func (p *ProgressReporter) EndKeyStageRecord(name string) {
	p.ks[name].endTime = time.Now()
	currentDuration := p.ks[name].endTime.Sub(p.ks[name].startTime)
	p.ks[name].totalDuration += currentDuration
	p.ks[name].count++

	if currentDuration > p.ks[name].maxDuration {
		p.ks[name].maxDuration = currentDuration
	}
	if currentDuration < p.ks[name].minDuration {
		p.ks[name].minDuration = currentDuration
	}
}

// Report 打印整体持续时间和每个关键阶段的详细统计信息。
// 输出格式：
// Total duration: X s
//
//	Key: stage_name, Count: Y, Total duration: Z s, Max duration: A s, Min duration: B s
//
// 使用场景：在所有计时操作完成后，调用此方法将性能数据输出到控制台或日志，以便分析。
func (p *ProgressReporter) Report() {
	// print total duration
	fmt.Printf("Total duration: %d s\n", p.totalDuration/time.Second)

	// print key stage
	for k, v := range p.ks {
		fmt.Printf("\t Key: %s, Count: %d, Total duration: %d s, Max duration: %d s, Min duration: %d s\n", k, v.count, v.totalDuration/time.Second, v.maxDuration/time.Second, v.minDuration/time.Second)
	}
}

// keyStage 存储单个关键阶段的耗时统计信息。
type keyStage struct {
	name                     string        // 关键阶段的名称
	totalDuration            time.Duration // 此关键阶段的总累积持续时间
	count                    int           // 此关键阶段被记录的次数
	maxDuration, minDuration time.Duration // 此关键阶段记录到的最大和最小单次持续时间
	startTime, endTime       time.Time     // 用于计算单次关键阶段持续时间的开始和结束时间
}
