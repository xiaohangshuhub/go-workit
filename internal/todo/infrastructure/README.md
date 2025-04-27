# 基础设施层

基础设施层（Infrastructure Layer）负责为应用程序提供技术支持，包括数据持久化、外部服务的集成、消息队列、日志记录等。它是应用程序与外部世界交互的桥梁。

## 职责
- 提供对数据库、文件系统、缓存等的访问。
- 集成外部服务（如第三方 API、消息队列等）。
- 实现领域层定义的仓储接口（Repository Interface）。
- 提供跨领域的技术能力（如日志、配置管理等）。

## 组成
基础设施层通常包含以下组件：
1. **持久化（Persistence）**  
   实现领域层定义的仓储接口，负责数据的存储和读取。

2. **外部服务集成（External Services Integration）**  
   处理与第三方服务的交互，例如调用外部 API。

3. **消息队列（Message Queue）**  
   处理异步消息的发送和接收。

4. **通用技术支持（Utilities）**  
   提供日志记录、配置管理等通用功能。

## 设计原则
- **依赖倒置原则**：基础设施层应依赖于领域层，而不是反过来。
- **技术细节隔离**：将具体的技术实现细节封装在基础设施层，避免泄漏到其他层。
- **可替换性**：通过接口和依赖注入，使基础设施层的实现可以轻松替换。

## 示例
以下是一个简单的待办事项（Todo）基础设施层示例：

### 仓储实现（TaskRepository）
```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/infrastructure/task_repository.go
package infrastructure

import (
    "errors"
    "sync"

    "github.com/lovehang/newb/internal/todo/domain"
)

// InMemoryTaskRepository 是一个基于内存的任务仓储实现
type InMemoryTaskRepository struct {
    tasks map[string]domain.Task
    mu    sync.RWMutex
}

// NewInMemoryTaskRepository 创建一个新的内存任务仓储
func NewInMemoryTaskRepository() *InMemoryTaskRepository {
    return &InMemoryTaskRepository{
        tasks: make(map[string]domain.Task),
    }
}

// Save 保存任务
func (r *InMemoryTaskRepository) Save(task domain.Task) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.tasks[task.ID] = task
    return nil
}

// FindByID 根据 ID 查找任务
func (r *InMemoryTaskRepository) FindByID(id string) (domain.Task, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    task, exists := r.tasks[id]
    if !exists {
        return domain.Task{}, errors.New("任务未找到")
    }
    return task, nil
}
```

### 外部服务集成（示例：发送通知）
```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/infrastructure/notification_service.go
package infrastructure

import "fmt"

// NotificationService 提供通知功能
type NotificationService struct{}

// SendNotification 发送通知
func (s *NotificationService) SendNotification(taskID, message string) error {
    // 假设这里调用了外部通知服务
    fmt.Printf("通知发送成功：任务ID=%s, 消息=%s\n", taskID, message)
    return nil
}
```

### 配置管理
```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/infrastructure/config.go
package infrastructure

import "os"

// Config 提供配置管理功能
type Config struct{}

// Get 获取配置值
func (c *Config) Get(key string) string {
    return os.Getenv(key)
}
```

### 日志记录
```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/infrastructure/logger.go
package infrastructure

import "log"

// Logger 提供日志记录功能
type Logger struct{}

// Info 记录信息日志
func (l *Logger) Info(message string) {
    log.Printf("[INFO] %s\n", message)
}

// Error 记录错误日志
func (l *Logger) Error(message string) {
    log.Printf("[ERROR] %s\n", message)
}
```

## 注意事项
- 基础设施层的实现应尽量保持简单，避免复杂的业务逻辑。
- 使用接口隔离领域层与基础设施层，确保领域层的独立性。
- 基础设施层的组件应支持可替换性，以便在不同环境下使用不同的实现（如测试环境使用内存数据库，生产环境使用关系型数据库）。

通过清晰的基础设施层设计，可以确保应用程序与外部世界的交互清晰、可维护，并且能够适应未来的技术变化。