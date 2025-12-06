package yingyong

import (
	"context"
)
import "fmt"
import "time"

/**
 * context包 提供了一种在goroutine之间传递请求范围的元数据，取消信号和超时控制
 * type Context interface {
 *		Deadline() (deadline time.Time, ok bool) // 返回上下文截止时间
 *		Done() <-chan struct{} // 获取一个channel，当上下文取消时，会关闭该channel
 *		Err() error // 返回取消的原因 (如超时，手动取消)
 * 		Value(key any) any  //获取上下文存储的键值对
 *	}
 * 2、上下文的类型
 *  空上下文： context.Background() 用于主函数/初始化，context.TODO (不确定上下文类型使用)
 *  取消上下文： context.WithCancel(parent) 手动取消（can cel）
 *  超时上下文： context.WithTimeout(parent, timeout) 超时取消（time out）
 */

//1、手动取消上下文

func worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("worker done")
			return
		default:
			fmt.Println("worker working")
			time.Sleep(time.Second)
		}
	}
}

func fetchData(ctx context.Context) error {
	//模拟耗时操作
	select {
	case <-time.After(3 * time.Second):
		fmt.Println("fetchData done")
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func main() {
	//创建可取消上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() //确保最终调用取消，释放资源

	go worker(ctx)

	//5秒后取消上下文
	time.Sleep(2 * time.Second)

	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	err := fetchData(ctx)
	if err != nil {
		fmt.Println("fetchData error:", err)
	}
}

/**
 * 通用方法，全局字符串,这种类型断言会在运行时引发panic，所以尽量避免使用
 * 字符串建可能会在软件包之间发生冲突
 */

const (
	KeyUserID = "user_id"
)

func storeUserInfo(ctx context.Context, userID int64) context.Context {
	ctx = context.WithValue(ctx, KeyUserID, userID)
	return ctx
}

func getUserInfo(ctx context.Context) string {
	userID := ctx.Value(KeyUserID).(string)
	return userID
}

//以上弊端，如上所写，更好的方法：结构化上下文值

type contextKey struct{}
type ContextValue struct {
	UserID int64
	Name   string
	Email  string
}

func NewContext(ctx context.Context, value *ContextValue) context.Context {
	return context.WithValue(ctx, contextKey{}, value)
}

// 获取结构化上下文值
func FromContext(ctx context.Context) *ContextValue {
	v, ok := ctx.Value(contextKey{}).(*ContextValue)
	if !ok {
		return nil
	}
	return v
}

func testxx() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		//模拟一些工作
		time.Sleep(time.Second)
		cancel()
	}()

	//等待取消信号
	select {
	case <-ctx.Done():
		fmt.Println("上下文被取消了", ctx.Err())
	case <-time.After(5 * time.Second):
		fmt.Println("超时了")
	}
}
