# go-weixin-sogou
[![Go Reference](https://pkg.go.dev/badge/github.com/reveever/go-weixin-sogou.svg)](https://pkg.go.dev/github.com/reveever/go-weixin-sogou)

基于搜狗微信入口，获取公众号文章和话题

通过 goquery 实现，支持文章搜索，公众号搜索，文章获取，话题列表获取

## Example

引用：`go get -u github.com/reveever/go-weixin-sogou`.

搜索文章
```go
package main

import (
	"fmt"
  
	wxsg "github.com/reveever/go-weixin-sogou"
)

func main() {
	results, err := wxsg.SearchArticle("title", 1)
	if err != nil {
		panic(err)
	}

	for i, v := range results {
		fmt.Printf("%d: %s (%s)\n", i, v.Title, v.AccName)
	}
}
```

获取公众号最新文章
```go
package main

import (
	"fmt"

	wxsg "github.com/reveever/go-weixin-sogou"
)

func main() {
	article, err := wxsg.GetLatestArticleByAccount("睡前消息编辑部", "MQZstudio")
	if err != nil {
		panic(err)
	}

	fmt.Println(article)
}
```

[Doc & Examples](https://pkg.go.dev/github.com/reveever/go-weixin-sogou)
