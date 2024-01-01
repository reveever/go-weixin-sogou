package wxsg

import (
	"fmt"
)

func ExampleSearchAccount() {
	resutls, err := SearchAccount("睡前消息", 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i, result := range resutls {
		fmt.Printf("[%d] %s (%s) %s\n", i+1, result.Name, result.WeixinID, result.Introduction)
		if result.Identify != "" {
			fmt.Println(result.Identify)
		}
		if result.LatestArticleTitle != "" {
			fmt.Printf("%s %s\n", result.LatestArticleTitle, result.LatestArticlePubTime.String())
		}
		fmt.Println()
	}
}

func ExampleSearchArticle() {
	resutls, err := SearchArticle("睡前消息【2021-12-31】", 1)
	if err != nil {
		fmt.Println(err)
		return
	}
	for i, result := range resutls {
		fmt.Printf("[%d] %s (%s) %s\n", i+1, result.Title, result.AccName, result.PubTime.String())
		fmt.Println(result.Preview)
		fmt.Println()
	}
}

func ExampleGetArticleByTitle() {
	article, err := GetArticleByTitle("睡前消息【2021-12-31】政府给“剩女”出“嫁妆”", "睡前消息编辑部")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(article.String())
}

func ExampleGetLatestArticleByAccount() {
	article, err := GetLatestArticleByAccount("睡前消息编辑部", "MQZstudio")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(article.String())
}

func ExampleGetArticleByUrl() {
	article, err := GetArticleByUrl("https://mp.weixin.qq.com/s/qgr3OR5Xha8MWMMv0mV7_A")
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(article.Albums)
	fmt.Println(article.String())
}

func ExampleGetAlbumByUrl() {
	album, err := GetAlbumByID("2036709839434842113", false)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(album.Name, album.AccName, album.Count)
	for _, info := range album.Articles {
		fmt.Printf("%d. %s\n%s\n\n", info.Index, info.Title, info.PubTime.String())
	}
}
