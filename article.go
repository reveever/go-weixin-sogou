package wxsg

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

var (
	pubTimeReg = regexp.MustCompile(`var oriCreateTime = '(\d+)'`)
)

// 公众号文章数据
type Article struct {
	URL     string
	Title   string
	Author  string
	AccName string
	PubTime time.Time
	Albums  []AlbumInfo
	Node    *html.Node
}

// 公众号文章内话题信息
type AlbumInfo struct {
	ID    string
	Name  string
	Count string
	Url   string
}

func NewArticle(url string, node *html.Node) *Article {
	article := &Article{
		URL:  url,
		Node: node,
	}
	doc := goquery.NewDocumentFromNode(node)
	contentSele := doc.Find("#img-content")
	article.Title = strings.TrimSpace(contentSele.Find("#activity-name").Text())
	article.Author = strings.TrimSpace(contentSele.Find("#meta_content > span.rich_media_meta.rich_media_meta_text").Text())
	article.AccName = strings.TrimSpace(contentSele.Find("#js_name").Text())
	doc.Find("#activity-detail > script").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "oriCreateTime") {
			ret := pubTimeReg.FindStringSubmatch(text)
			if len(ret) == 2 {
				s, _ := strconv.ParseInt(ret[1], 10, 64)
				article.PubTime = time.Unix(s, 0)
			}
		}
	})

	scripts := doc.Find("#activity-detail > script")
	scripts.EachWithBreak(func(i int, s *goquery.Selection) bool {
		pattern := regexp.MustCompile(`var\s*publicTagInfo\s*=\s*\[\s*({[\s\S]*?}\s*,)\s*\]\s*;`)
		match := pattern.FindStringSubmatch(s.Text())
		if len(match) != 2 {
			return true
		}
		pattern2 := regexp.MustCompile(`{([\s\S]*?)}\s*,\s*`)
		titleReg := regexp.MustCompile(`title:\s*'(.*)'\s*,`)
		sizeReg := regexp.MustCompile(`size:\s*'(.*)'\s*\*\s*1\s*,`)
		linkReg := regexp.MustCompile(`link:\s*'(.*)'\s*,`)
		albumIDReg := regexp.MustCompile(`albumId:\s*'(.*)'\s*,`)
		for _, match := range pattern2.FindAllStringSubmatch(match[1], -1) {
			if len(match) != 2 {
				continue
			}
			title := titleReg.FindStringSubmatch(match[1])
			count := sizeReg.FindStringSubmatch(match[1])
			link := linkReg.FindStringSubmatch(match[1])
			albumIDReg := albumIDReg.FindStringSubmatch(match[1])
			if len(title) != 2 || len(count) != 2 || len(link) != 2 || len(albumIDReg) != 2 {
				continue
			}
			article.Albums = append(article.Albums, AlbumInfo{
				ID:    albumIDReg[1],
				Name:  title[1],
				Count: count[1],
				Url:   link[1],
			})
		}
		return false
	})
	return article
}

// 打印文章, 只保留文字和图片链接
func (a *Article) Content() string {
	if a == nil || a.Node == nil {
		return ""
	}
	contentSele := goquery.NewDocumentFromNode(a.Node).Find("#js_content")

	var (
		b bytes.Buffer
		f func(*html.Node)
	)
	f = func(n *html.Node) {
		switch {
		case n.Type == html.ElementNode && n.Data == "script":
			return
		case n.Type == html.ElementNode && n.Data == "p":
			defer b.WriteString("\n")
		case n.Type == html.ElementNode && n.Data == "img":
			for _, attr := range n.Attr {
				if attr.Key == "data-src" {
					b.WriteString(fmt.Sprintf("[img %s]\n", attr.Val))
				}
			}
			return
		case n.Type == html.TextNode:
			b.WriteString(n.Data)
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(contentSele.Get(0))
	return b.String()
}

func (a *Article) String() string {
	var b bytes.Buffer
	b.WriteString(a.URL + "\n")
	b.WriteString(a.Title + "\n")
	b.WriteString(a.Author + " ")
	b.WriteString(a.AccName + " ")
	b.WriteString(a.PubTime.String() + "\n")
	b.WriteString(a.Content())
	return b.String()
}

func (c Client) GetArticleByUrl(url string) (article *Article, err error) {
	if strings.Contains(url, "weixin.sogou.com/link") {
		url, err = GetArticleRealUrl(url)
		if err != nil {
			return
		}
	} else if !strings.Contains(url, "mp.weixin.qq.com/s") {
		err = fmt.Errorf("invalid url: %s", url)
		return
	}

	resp, err := c.Do(c.buildGetReq(url))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	node, err := html.Parse(resp.Body)
	if err != nil {
		return
	}

	article = NewArticle(url, node)
	return
}

func (c Client) GetArticleByTitle(title, accName string) (*Article, error) {
	results, err := SearchArticle(title, 1)
	if err != nil {
		return nil, fmt.Errorf("search article: %v", err)

	}
	for _, result := range results {
		if result.Title != title {
			continue
		}
		if accName != "" && result.AccName != accName {
			continue
		}
		article, err := c.GetArticleByUrl(result.Url)
		if err != nil {
			return nil, fmt.Errorf("get article: %v", err)
		}
		return article, nil
	}
	return nil, errors.New("not found")
}

func (c Client) GetLatestArticleByAccount(accName, weixinID string) (*Article, error) {
	results, err := SearchAccount(accName, 1)
	if err != nil {
		return nil, fmt.Errorf("search account: %v", err)
	}

	for _, result := range results {
		if result.Name != accName {
			continue
		}
		if weixinID != "" && result.WeixinID != weixinID {
			continue
		}
		article, err := c.GetArticleByUrl(result.LatestArticleUrl)
		if err != nil {
			return nil, fmt.Errorf("get article: %v", err)
		}
		return article, nil
	}
	return nil, errors.New("not found")
}
