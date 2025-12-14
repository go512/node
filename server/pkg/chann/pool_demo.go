package chann

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type DemoTask struct {
	TaskID   int
	Content  string
	ExecTime time.Duration //任务实际执行时长

	//自定义回调
	customOnComplete func(result any)
	customOnError    func(err error)
}

// Excute 实现task接口
func (t *DemoTask) Execute(ctx context.Context) (any, error) {
	//模拟任务执行，检查上下文是否超时
	select {
	case <-ctx.Done():
		return nil, ctx.Err() //已超时，直接返回
	default:
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		elapsed := 0 * time.Millisecond
		for elapsed < t.ExecTime {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-ticker.C:
				elapsed += 100 * time.Millisecond
			}
		}

		//模拟业务错误（taskId 为偶数时失败）
		if t.TaskID%2 == 0 {
			return nil, errors.New(fmt.Sprintf("任务%d业务执行失败（ID为偶数）", t.TaskID))
		}

		return fmt.Sprintf("任务%d执行结果：%s", t.TaskID, t.Content), nil
	}
}

func (t *DemoTask) OnComplete(result any) {
	if t.customOnComplete != nil {
		t.customOnComplete(result)
	} else {
		fmt.Printf("【回调-成功】任务%d：%v\n", t.TaskID, result)
	}
}

// OnError 失败回调
func (t *DemoTask) OnError(err error) {
	if t.customOnError != nil {
		t.customOnError(err)
	} else {
		fmt.Printf("【回调-失败】任务%d：%v\n", t.TaskID, err)
	}
}

func (t *DemoTask) SetOnComplete(fn func(result any)) {
	t.customOnComplete = fn
}

func (t *DemoTask) SetOnError(fn func(err error)) {
	t.customOnError = fn
}
