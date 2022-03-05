package wxsg

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"regexp"

	"golang.org/x/net/publicsuffix"
)

const DefaultUA = "User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36 Edg/98.0.1108.62"

var (
	DefaultClient  *Client
	pubTimeScript  = regexp.MustCompile(`document\.write\(timeConvert\('(\d+)'\)\)`)
	urlMergeCmdReg = regexp.MustCompile(`url \+= '(.*)';`)
)

// 可自定义 Client, 注意处理 cookie, 使用统一 ua, 切勿高频调用, 否则容易出现图形验证码, 对此暂未处理
type Client struct {
	*http.Client
	UseAgent string
}

func (c *Client) buildGetReq(urlFormat string, a ...interface{}) *http.Request {
	url := fmt.Sprintf(urlFormat, a...)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("User-Agent", c.UseAgent)
	return req
}

func init() {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		panic(err)
	}
	DefaultClient = &Client{
		Client: &http.Client{
			Jar: jar,
		},
		UseAgent: DefaultUA,
	}
}

// 搜狗微信 - 搜公众号
func SearchAccount(query string, page int) ([]AccountInfo, error) {
	return DefaultClient.SearchAccount(query, page)
}

// 搜狗微信 - 搜文章
func SearchArticle(query string, page int) ([]ArticleInfo, error) {
	return DefaultClient.SearchArticle(query, page)
}

// 将搜索获得的中间链接转换为文章真实链接
func GetArticleRealUrl(url string) (string, error) {
	return DefaultClient.GetArticleRealUrl(url)
}

// 通过链接访问文章, 支持文章直链或跳转链接
func GetArticleByUrl(url string) (*Article, error) {
	return DefaultClient.GetArticleByUrl(url)
}

// 通过标题与公众号名访问文章, 公众号名可为空, 默认获取搜索结果中第一个完全匹配的文章
func GetArticleByTitle(title, accName string) (*Article, error) {
	return DefaultClient.GetArticleByTitle(title, accName)
}

// 获取公众号最新文章信息, 公众号 ID 可为空, 默认读取搜索结果中第一个完全匹配的公众号
func GetLatestArticleByAccount(accName, weixinID string) (*Article, error) {
	return DefaultClient.GetLatestArticleByAccount(accName, weixinID)
}

// 通过链接访问话题页, 支持正向或反向排序
func GetAlbumByUrl(url string, isReverse bool) (*Album, error) {
	return DefaultClient.GetAlbumByUrl(url, isReverse)
}

// 通过话题 ID 访问话题页, 支持正向或反向排序
func GetAlbumByID(id string, isReverse bool) (*Album, error) {
	return DefaultClient.GetAlbumByID(id, isReverse)
}
