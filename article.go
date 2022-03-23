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
	pubTimeReg = regexp.MustCompile(`"(1[6-9]\d{8})"`)
)

// 公众号文章数据
type Article struct {
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

func NewArticle(node *html.Node) *Article {
	article := &Article{
		Node: node,
	}
	doc := goquery.NewDocumentFromNode(node)
	contentSele := doc.Find("#img-content")
	article.Title = strings.TrimSpace(contentSele.Find("#activity-name").Text())
	article.Author = strings.TrimSpace(contentSele.Find("#meta_content > span.rich_media_meta.rich_media_meta_text").Text())
	article.AccName = strings.TrimSpace(contentSele.Find("#js_name").Text())

	doc.Find("#activity-detail > script").Each(func(i int, s *goquery.Selection) {
		text := s.Text()
		if strings.Contains(text, "publish_time") {
			ret := pubTimeReg.FindStringSubmatch(text)
			if len(ret) == 2 {
				s, _ := strconv.ParseInt(ret[1], 10, 64)
				article.PubTime = time.Unix(s, 0)
			}
		}
	})

	tagsSele := contentSele.Find("#js_tags > div.article-tags")
	tagsSele.Children().Each(func(i int, s *goquery.Selection) {
		c := s.Children()
		article.Albums = append(article.Albums, AlbumInfo{
			ID:    strings.TrimSpace(s.AttrOr("data-album_id", "")),
			Name:  strings.TrimSpace(c.First().Text()),
			Count: strings.TrimSpace(c.Children().First().Text()),
			Url:   strings.TrimSpace(s.AttrOr("data-url", "")),
		})
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

	article = NewArticle(node)
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
