# 领域层

领域层（Domain Layer）是应用程序的核心，负责处理业务逻辑和规则。它独立于其他层，专注于表达领域概念和实现业务规则。

## 职责
- 定义核心业务逻辑和规则。
- 表达领域模型（实体、值对象、聚合等）。
- 提供接口供应用层调用。
- 确保业务逻辑的完整性和一致性。

## 组成
领域层通常包含以下组件：
1. **实体（Entity）**  
   表示具有唯一标识的业务对象，封装与其相关的业务逻辑。  
   **示例**：`Task` 是一个待办事项的实体，包含任务的标题、描述、状态等。

2. **值对象（Value Object）**  
   表示不可变的业务对象，通常用于描述属性或特性。  
   **示例**：`Priority` 是一个值对象，表示任务的优先级（如高、中、低）。

3. **聚合（Aggregate）**  
   一组相关的实体和值对象的集合，具有一个聚合根（Aggregate Root）作为入口。  
   **示例**：`TodoList` 是一个聚合，包含多个 `Task` 实体。

4. **领域服务（Domain Service）**  
   表示不属于任何实体或值对象的业务逻辑，通常是跨多个聚合的操作。  
   **示例**：`TaskScheduler` 是一个领域服务，用于根据优先级和截止日期安排任务。

5. **领域事件（Domain Event）**  
   用于表示领域中发生的重要业务事件。  
   **示例**：`TaskCompletedEvent` 表示任务完成时触发的事件。

6. **仓储接口（Repository Interface）**  
   定义持久化操作的接口，具体实现由基础设施层提供。  
   **示例**：`TaskRepository` 定义了保存和查询任务的方法。

## 设计原则
- **高内聚低耦合**：领域层应尽量减少对外部依赖，保持独立性。
- **领域驱动设计（DDD）**：遵循 DDD 的原则，确保领域模型与业务需求一致。
- **不可变性**：值对象应设计为不可变，确保数据一致性。
- **单一职责原则**：每个组件只负责一个明确的领域概念。

## 示例
以下是领域层的一个简单示例，展示了如何设计一个待办事项的领域模型：

### 实体（Task）
```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/domain/task.go
package domain

import "time"

// Task 表示一个任务的实体
type Task struct {
    ID          string
    Title       string
    Description string
    Priority    Priority
    Completed   bool
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// MarkAsCompleted 将任务标记为已完成
func (t *Task) MarkAsCompleted() {
    t.Completed = true
    t.UpdatedAt = time.Now()
}
```

### 值对象（Priority）
```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/domain/priority.go
package domain

// Priority 表示任务的优先级
type Priority struct {
    Level string // 高、中、低
}

// NewPriority 创建一个新的优先级
func NewPriority(level string) Priority {
    if level != "高" && level != "中" && level != "低" {
        panic("无效的优先级")
    }
    return Priority{Level: level}
}
```

### 聚合（TodoList）
```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/domain/todolist.go
package domain

// TodoList 表示一个任务列表的聚合
type TodoList struct {
    Tasks []Task
}

// AddTask 添加任务到任务列表
func (tl *TodoList) AddTask(task Task) {
    tl.Tasks = append(tl.Tasks, task)
}

// GetPendingTasks 获取未完成的任务
func (tl *TodoList) GetPendingTasks() []Task {
    var pending []Task
    for _, task := range tl.Tasks {
        if !task.Completed {
            pending = append(pending, task)
        }
    }
    return pending
}
```

### 领域服务（TaskScheduler）
```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/domain/taskscheduler.go
package domain

import "sort"

// TaskScheduler 表示任务调度的领域服务
type TaskScheduler struct{}

// ScheduleTasks 根据优先级和创建时间排序任务
func (ts *TaskScheduler) ScheduleTasks(tasks []Task) []Task {
    sort.Slice(tasks, func(i, j int) bool {
        if tasks[i].Priority.Level == tasks[j].Priority.Level {
            return tasks[i].CreatedAt.Before(tasks[j].CreatedAt)
        }
        return tasks[i].Priority.Level > tasks[j].Priority.Level
    })
    return tasks
}
```

## 注意事项
- 领域层不应依赖于框架或外部库，保持纯粹性。
- 避免将基础设施或应用层的逻辑混入领域层。
- 通过领域事件解耦复杂的业务逻辑。

通过清晰的领域层设计，可以确保待办事项应用程序的业务逻辑清晰、可维护，并且能够适应未来的变化。