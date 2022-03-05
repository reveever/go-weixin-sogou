package wxsg

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	WeixinSogouUrl  = "https://weixin.sogou.com"
	SearchUrlFormat = WeixinSogouUrl + "/weixin?ie=utf8&type=%d&page=%d&query=%s"
)

type AccountInfo struct {
	Name     string
	Url      string
	Avatar   string
	QRCode   string
	WeixinID string
	// 可从 /websearch/weixin/pc/anti_account.jsp 可获得月发文数
	// Activity          string
	Introduction         string
	Identify             string
	LatestArticleTitle   string
	LatestArticleUrl     string
	LatestArticlePubTime time.Time
}

type ArticleInfo struct {
	Title   string
	Url     string
	Preview string
	AccName string
	AccUrl  string
	PubTime time.Time
	Image   string
}

func (c Client) SearchAccount(query string, page int) (results []AccountInfo, err error) {
	if page < 1 {
		page = 1
	}
	query = strings.ReplaceAll(query, " ", "+")
	req := c.buildGetReq(SearchUrlFormat, 1, page, query)
	resp, err := c.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	resultsSele := doc.Find("#main > div.news-box > ul").Children()
	if resultsSele.Length() == 0 {
		err = errors.New("no results")
		return
	}

	resultsSele.Each(func(i int, s *goquery.Selection) {
		result := AccountInfo{}
		s.Children().Each(func(i int, s *goquery.Selection) {
			key := s.Find("dt").Text()
			switch {
			case key == "":
				nameSele := s.Find("div.txt-box > p.tit > a")
				result.Name = nameSele.Text()
				if href, ok := nameSele.Attr("href"); ok {
					result.Url = WeixinSogouUrl + href
				}
				if src, ok := s.Find("div.img-box > a > img").Attr("src"); ok {
					result.Avatar = "https:" + src
				}
				if src, ok := s.Find("div.ew-pop > span > img:nth-child(3)").Attr("src"); ok {
					result.QRCode = src
				}
				result.WeixinID = s.Find("div.txt-box > p.info > label").Text()
			case strings.Contains(key, "功能介绍"):
				result.Introduction = s.Find("dd").Text()
			case strings.Contains(key, "认证"):
				result.Identify = strings.TrimSpace(s.Find("dd").Text())
			case strings.Contains(key, "最近文章"):
				articleSele := s.Find("dd > a")
				result.LatestArticleTitle = articleSele.Text()
				if href, ok := articleSele.Attr("href"); ok {
					result.LatestArticleUrl = WeixinSogouUrl + href
				}
				if ret := pubTimeScript.FindStringSubmatch(s.Find("dd > span > script").Text()); len(ret) == 2 {
					t, _ := strconv.ParseInt(ret[1], 10, 64)
					result.LatestArticlePubTime = time.Unix(t, 0)
				}
			}
		})

		results = append(results, result)
	})
	return
}

func (c Client) SearchArticle(query string, page int) (results []ArticleInfo, err error) {
	if page < 1 {
		page = 1
	}
	query = strings.ReplaceAll(query, " ", "+") // + "-" 加个无用字符能似乎提升准确率？

	req := c.buildGetReq(SearchUrlFormat, 2, page, query)
	resp, err := c.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return
	}

	resultsSele := doc.Find("#main > div.news-box > ul").Children()
	if resultsSele.Length() == 0 {
		err = errors.New("no results")
		return
	}

	resultsSele.Each(func(i int, s *goquery.Selection) {
		titleSele := s.Find("div.txt-box > h3 > a")
		accountSele := s.Find("div.txt-box > div > a")
		result := ArticleInfo{
			Title:   titleSele.Text(),
			Preview: s.Find("div.txt-box > p").Text(),
			AccName: accountSele.Text(),
		}

		if href, ok := titleSele.Attr("href"); ok {
			result.Url = WeixinSogouUrl + href
		}
		if href, ok := accountSele.Attr("href"); ok {
			result.AccUrl = WeixinSogouUrl + href
		}
		if ret := pubTimeScript.FindStringSubmatch(s.Find("div.txt-box > div > span > script").Text()); len(ret) == 2 {
			t, _ := strconv.ParseInt(ret[1], 10, 64)
			result.PubTime = time.Unix(t, 0)
		}
		if src, ok := s.Find("div.img-box > a > img").Attr("src"); ok {
			result.Image = "https:" + src
		}

		results = append(results, result)
	})
	return
}

// 搜索获得的文章地址往往地址指向了一个中间页面, 需要拼接字符串获得真实地址
func (c Client) GetArticleRealUrl(url string) (url2 string, err error) {
	resp, err := c.Do(c.buildGetReq(url))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	if !bytes.Contains(buf, []byte(`url.replace("@", "");`)) {
		err = fmt.Errorf("unexpect page: %s", string(buf))
		return
	}

	ret := urlMergeCmdReg.FindAllSubmatch(buf, -1)
	for _, result := range ret {
		if len(result) != 2 {
			err = fmt.Errorf("unexpect page: %s", string(buf))
			return
		}
		url2 += string(result[1])
	}
	url2 = strings.ReplaceAll(url2, "@", "")
	return
}
