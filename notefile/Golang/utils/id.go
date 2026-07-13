package utils

import (
	"encoding/base32"
	"fmt"
	"strings"
	"sync"
)

const salt = "6rx7aXuTnVtM1ZMXOgJNHX9cmibQs3vAOApvT9KPQe"

var b32encoder = base32.NewEncoding("abcdefghijklmnopqrstuvwxyz234567").WithPadding(base32.StdPadding)

func Encode(s string) string {
	if s == "" {
		return ""
	}
	return b32encoder.EncodeToString(mangle([]byte(toAbbr(s))))
}

func Decode(s string) string {
	if s == "" {
		return ""
	}
	b, err := b32encoder.DecodeString(s)
	if err != nil {
		return ""
	}
	return fromAbbr(string(unmangle(b)))

}

func mangle(s []byte) []byte {
	n := len(s)            // 获取输入字节总长度
	b := make([]byte, n)   // 创建等长空字节数组，存放结果
	lastPos := n - 1       // 最后一个字节的索引
	lastByte := s[lastPos] // 取出原始最后一字节，全程不变参与与运算

	// 循环：只处理 0 ～ n-2 所有前置字节（跳过最后一位）
	for i := 0; i < lastPos; i++ {
		b[i] = s[i] ^ lastByte ^ salt[i]
	}

	b[lastPos] = s[lastPos] // 最后一个字节直接复制，不做任何修改
	return b
}

func unmangle(s []byte) []byte {
	return mangle(s)
}

func simpleCipher(s []byte) []byte {
	n := len(s)
	b := make([]byte, n)

	for i := 0; i < n; i++ {
		b[i] = s[i] ^ 0xff
	}
	return b
}

var WelkAbbr = map[string]string{
	"sr:match:":  "{1}",
	"sr:season:": "{2}",
	"sr:player:": "{3}",
	"ts:match:":  "[1]",
	"ao:match:":  "<1>",
}

var (
	toAbbrReplacer   *strings.Replacer
	fromAbbrReplacer *strings.Replacer
	once             sync.Once
)

func initReplaycers() {
	once.Do(func() {
		toAbbrValues := make([]string, 0, len(WelkAbbr)*2)
		fromabbrValues := make([]string, 0, len(WelkAbbr)*2)

		for k, v := range WelkAbbr {
			toAbbrValues = append(toAbbrValues, k, v)
			fromabbrValues = append(fromabbrValues, v, k)
		}

		fmt.Println(toAbbrValues)
		fmt.Println(fromabbrValues)

		toAbbrReplacer = strings.NewReplacer(toAbbrValues...)
		fromAbbrReplacer = strings.NewReplacer(fromabbrValues...)
	})
}

func toAbbr(s string) string {
	initReplaycers()
	return toAbbrReplacer.Replace(s)
}

func fromAbbr(s string) string {
	initReplaycers()
	return fromAbbrReplacer.Replace(s)
}
