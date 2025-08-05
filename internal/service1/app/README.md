# 应用层

应用层（Application Layer）负责协调用户请求、调用领域层处理业务逻辑，并将结果返回给用户。它是领域层和接口层之间的桥梁，主要处理应用程序的用例逻辑。

## 职责
- 调用领域层完成具体的业务操作。
- 处理用户请求并返回结果。
- 管理事务（如果需要）。
- 不包含业务逻辑，仅负责用例的编排。

## 组成
应用层通常包含以下组件：
1. **应用服务（Application Service）**  
   应用服务是应用层的核心，负责实现具体的用例逻辑。它通过调用领域层的组件（如实体、领域服务等）来完成业务操作。

2. **DTO（数据传输对象）**  
   用于在应用层和接口层之间传递数据，通常是简单的结构体或类。

3. **命令和查询（Command & Query）**  
   应用层可以根据 CQRS（命令查询职责分离）模式，将写操作（命令）和读操作（查询）分开。

## 设计原则
- **无状态性**：应用服务应设计为无状态的，避免保存任何状态信息。
- **单一职责原则**：每个应用服务只负责一个具体的用例。
- **依赖领域层**：应用层应依赖领域层，而不是基础设施层。
- **事务管理**：如果需要事务，应用层负责开启和提交事务。

## 示例
以下是一个简单的待办事项（Todo）应用层示例：

### 应用服务（TaskService）
```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/application/task_service.go
package application

import (
    "time"

    "github.com/lovehang/newb/internal/todo/domain"
)

// TaskService 提供任务相关的应用服务
type TaskService struct {
    TaskRepo domain.TaskRepository
}

// CreateTask 创建一个新的任务
func (s *TaskService) CreateTask(title, description, priority string) (string, error) {
    task := domain.Task{
        ID:          generateID(), // 假设有一个生成唯一 ID 的方法
        Title:       title,
        Description: description,
        Priority:    domain.NewPriority(priority),
        Completed:   false,
        CreatedAt:   time.Now(),
        UpdatedAt:   time.Now(),
    }

    if err := s.TaskRepo.Save(task); err != nil {
        return "", err
    }
    return task.ID, nil
}

// CompleteTask 将任务标记为已完成
func (s *TaskService) CompleteTask(taskID string) error {
    task, err := s.TaskRepo.FindByID(taskID)
    if err != nil {
        return err
    }

    task.MarkAsCompleted()
    return s.TaskRepo.Save(task)
}
```

### 数据传输对象（DTO）
```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/application/dto.go
package application

// TaskDTO 表示任务的数据传输对象
type TaskDTO struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Priority    string `json:"priority"`
    Completed   bool   `json:"completed"`
    CreatedAt   string `json:"created_at"`
    UpdatedAt   string `json:"updated_at"`
}
```

### 命令和查询
```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/application/commands.go
package application

// CreateTaskCommand 表示创建任务的命令
type CreateTaskCommand struct {
    Title       string
    Description string
    Priority    string
}

// CompleteTaskCommand 表示完成任务的命令
type CompleteTaskCommand struct {
    TaskID string
}
```

```go
// filepath: /Users/xiancheng/github/lovehang/newb/newb/internal/todo/application/queries.go
package application

// GetTaskQuery 表示获取任务的查询
type GetTaskQuery struct {
    TaskID string
}
```

## 注意事项
- 应用层不应包含业务逻辑，所有业务逻辑应放在领域层。
- 应用层可以依赖领域层和基础设施层，但不应直接操作基础设施层的实现细节。
- 应用服务应尽量保持简单，避免复杂的逻辑。

通过清晰的应用层设计，可以确保应用程序的用例逻辑清晰、易于维护，并且能够适应未来的扩展需求。