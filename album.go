package wxsg

import (
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// 话题页数据
type Album struct {
	Name     string
	AccName  string
	Count    string
	Articles []ArticleInfo2
	Node     *html.Node
}

// 话题页文章信息
type ArticleInfo2 struct {
	Index   int
	Title   string
	Url     string
	PubTime time.Time
	Image   string
}

func NewAlbum(node *html.Node) *Album {
	album := &Album{
		Node: node,
	}
	doc := goquery.NewDocumentFromNode(node)
	contentSele := doc.Find("#js_content_overlay > div:nth-child(1) > div")
	album.Name = contentSele.Find("div.album__head.js_album_head > div > div > div > div > div").Text()
	album.AccName = contentSele.Find("div.album__head-content.no-desc > div > div.album__author-info.js_profile_info.js_wx_tap_highlight.wx_tap_link > div > div").Text()
	album.Count = contentSele.Find("div.album__head-content.no-desc > div > div.album__desc.album__head_fold.js_album_desc.no-desc > div.album__desc-content.js_album_desc_content > span").Text()
	listSele := contentSele.Find("div.album__content.js_album_bd > ul")
	listSele.Children().Each(func(i int, s *goquery.Selection) {
		index, _ := strconv.Atoi(s.AttrOr("data-pos_num", "-1"))
		info := ArticleInfo2{
			Index: index,
			Title: s.AttrOr("data-title", "TITLE NOT FOUND"),
			Url:   s.AttrOr("data-link", ""),
		}
		timeStr := s.Find("div.album__item-content > div.album__item-info > span").Text()
		t, _ := strconv.ParseInt(timeStr, 10, 64)
		info.PubTime = time.Unix(t, 0)
		if style, ok := s.Find("div.album__item-img").Attr("style"); ok {
			info.Image = strings.TrimSuffix(strings.TrimPrefix(style, "background-image: url("), ");")
		}

		album.Articles = append(album.Articles, info)
	})
	return album
}

func (c Client) GetAlbumByUrl(url string, isReverse bool) (album *Album, err error) {
	if isReverse {
		url += "&is_reverse=1"
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

	album = NewAlbum(node)
	return
}

func (c Client) GetAlbumByID(id string, isReverse bool) (album *Album, err error) {
	return GetAlbumByUrl("https://mp.weixin.qq.com/mp/appmsgalbum?action=getalbum&album_id="+id, isReverse)
}
