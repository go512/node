package ai

import (
	"context"
	"fmt"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
)

const (
	Api   = "https://ark.cn-beijing.volces.com/api/v3"
	Token = "a0102772-f786-4f9f-ab27-df922a280868"
	Model = "deepseek-v3-2-251201"
)

func AiFIBACountryV2(ctx context.Context, country string) (string, error) {
	prompt := fmt.Sprintf("以下是足球球队名称，只返回该球队所在的国家/地区名称（英文）. The country name is %s", country)

	client := openai.NewClient(
		option.WithBaseURL(Api),
		option.WithAPIKey(Token),
	)

	chatCompletion, err := client.Chat.Completions.New(ctx,
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage("You are a sports data expert."),
				openai.UserMessage(prompt),
			},
			Model:     Model,
			MaxTokens: param.NewOpt(int64(100)),
		})
	if err != nil {
		return "", err
	}
	return chatCompletion.Choices[0].Message.Content, err
}

func Chat(ctx context.Context, prompt string) (string, error) {

	client := openai.NewClient(
		option.WithBaseURL(Api),
		option.WithAPIKey(Token),
	)

	chatCompletion, err := client.Chat.Completions.New(ctx,
		openai.ChatCompletionNewParams{
			Messages: []openai.ChatCompletionMessageParamUnion{
				openai.SystemMessage("You are a sports data expert."),
				openai.UserMessage(prompt),
			},
			Model:     Model,
			MaxTokens: param.NewOpt(int64(500)),
		})
	if err != nil {
		return "", err
	}
	return chatCompletion.Choices[0].Message.Content, err
}

func CountryList(ctx context.Context) (string, error) {
	prompt := "根据维基百科 返回大洋洲国家英文名称 以及缩写以json格式返回"

	return Chat(ctx, prompt)
}
