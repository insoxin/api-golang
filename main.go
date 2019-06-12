package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/plugins/cors"
	"github.com/tidwall/gjson"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
)

/**
 * Author: Filmy
 * Group: Mlooc
 * Date: 2019/5/28
 * Time: 8:58
 */

type analysisController struct {
	beego.Controller
}

var pc_ua = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36"
var phone_ua = "Mozilla/5.0 (iPhone; CPU iPhone OS 11_0 like Mac OS X) AppleWebKit/604.1.38 (KHTML, like Gecko) Version/11.0 Mobile/15A372 Safari/604.1"

func main() {
	beego.BConfig.Listen.HTTPAddr = ""   // 监听所有网卡
	beego.BConfig.EnableGzip = true      // 启用Gzip压缩
	beego.BConfig.Listen.HTTPPort = 6969 // 监听端口9868

	beego.BConfig.RunMode = "dev" // 开发模式

	// 允许跨域访问
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowAllOrigins: true,
		//AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		//AllowHeaders:     []string{"Origin", "Authorization", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		//ExposeHeaders:    []string{"Content-Length", "Access-Control-Allow-Origin", "Access-Control-Allow-Headers", "Content-Type"},
		//AllowCredentials: true,
	}))

	beego.Router("", &analysisController{})
	beego.Router("/", &analysisController{})
	beego.Run()
}

func (this *analysisController) Get() {
	url := this.GetString("url")
	if strings.Index(url, "weishi.qq.com") != -1 {
		this.Data["json"] = weiShi(url)
	} else if strings.Index(url, "douyin.com") != -1 || strings.Index(url, "iesdouyin.com") != -1 {
		this.Data["json"] = douYin(url)
	} else if strings.Index(url, "pipix.com") != -1 {
		this.Data["json"] = ppx(url)
	} else if strings.Index(url, "izuiyou.com") != -1 {
		this.Data["json"] = zuiYou(url)
	} else if strings.Index(url, "huoshan.com") != -1 {
		this.Data["json"] = huoShan(url)
	} else if strings.Index(url, "kuaishou.com") != -1 || strings.Index(url, "gifshow.com") != -1 {
		this.Data["json"] = kuaiShou(url)
	} else {
		this.Data["json"] = Echo(400, "暂不支持该平台", nil)
	}
	this.ServeJSON()
}

func (this *analysisController) Post() {

}

func Echo(code int, msg string, data interface{}) map[string]interface{} {
	echoResult := make(map[string]interface{})
	echoResult["code"] = code
	echoResult["msg"] = msg
	echoResult["data"] = data
	return echoResult
}

func HttpGet(url string, ua string) string {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return string(body)
}

func HttpGetLocationUrl(url string, ua string) string {
	client := http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
	}
	req.Header.Set("User-Agent", ua)
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()
	return fmt.Sprintf("%v", resp.Request.URL)

}

func HttpPost(url, params string) string {
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, strings.NewReader(params))
	if err != nil {
		log.Println(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/74.0.3729.131 Safari/537.36")

	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}
	return string(body)
}

// 微视
func weiShi(url string) map[string]interface{} {
	feedid := regexp.MustCompile(`feed/(\w+)`).FindAllStringSubmatch(url, -1)
	if len(feedid) < 1 || len(feedid[0]) < 2 {
		return Echo(400, "参数错误", nil)
	}
	resp := HttpGet("https://h5.qzone.qq.com/webapp/json/weishi/WSH5GetPlayPage?feedid="+feedid[0][1], pc_ua)
	respJson := gjson.Parse(resp)
	if respJson.Get("data.feeds.0.video_url").String() == "" || respJson.Get("data.feeds.0.images.0.url").String() == "" {
		return Echo(400, respJson.Get("data.errmsg").String(), nil)
	}
	echoMap := make(map[string]interface{})
	echoMap["text"] = respJson.Get("data.feeds.0.feed_desc").String()
	echoMap["cover"] = respJson.Get("data.feeds.0.images.0.url").String()
	echoMap["playAddr"] = respJson.Get("data.feeds.0.video_url").String()
	return Echo(200, "", echoMap)
}

// 抖音
func douYin(url string) map[string]interface{} {
	resp := HttpGet(url, phone_ua)
	aweme_id := regexp.MustCompile(`itemId: "(.*?)",`).FindAllStringSubmatch(resp, -1)
	if len(aweme_id) < 1 || len(aweme_id[0]) < 2 {
		return Echo(400, "参数错误", nil)
	}

	resp = HttpGet("https://api-hl.amemv.com/aweme/v1/aweme/detail/?aid=1128&app_name=aweme&version_code=251&aweme_id="+aweme_id[0][1], phone_ua)
	respJson := gjson.Parse(resp)
	playAddr := respJson.Get("aweme_detail.video.play_addr.url_list.0").String()
	if playAddr == "" {
		return Echo(400, "解析错误", nil)
	}
	echoMap := make(map[string]interface{})
	echoMap["text"] = respJson.Get("aweme_detail.share_info.share_title").String()
	echoMap["cover"] = respJson.Get("aweme_detail.video.origin_cover.url_list.0").String()
	echoMap["playAddr"] = playAddr
	echoMap["music"] = respJson.Get("aweme_detail.music.play_url.url_list.0").String()
	return Echo(200, "", echoMap)
}

// 皮皮虾
func ppx(url string) map[string]interface{} {
	item_id := regexp.MustCompile(`\b[1-9]\d*`).FindAllStringSubmatch(url, -1)
	if len(item_id) < 1 || len(item_id[0]) < 1 {
		return Echo(400, "参数错误", nil)
	}
	resp := HttpGet("https://is.snssdk.com/bds/item/detail/?app_name=super&aid=1319&item_id="+item_id[0][0], pc_ua)
	respJson := gjson.Parse(resp)

	if respJson.Get("status_code").String() != "0" {
		return Echo(400, respJson.Get("prompt").String(), nil)
	}

	if respJson.Get("data.data.share.title").String() == "" || respJson.Get("data.data.video.video_fallback.url_list.0.url").String() == "" || respJson.Get("data.data.video.video_fallback.cover_image.url_list.0.url").String() == "" {
		return Echo(400, "解析失败", nil)
	}

	echoMap := make(map[string]interface{})
	echoMap["text"] = respJson.Get("data.data.share.title").String()
	echoMap["cover"] = respJson.Get("data.data.video.video_fallback.cover_image.url_list.0.url").String()
	echoMap["playAddr"] = respJson.Get("data.data.video.video_fallback.url_list.0.url").String()
	return Echo(200, "", echoMap)
}

// 最右
func zuiYou(url string) map[string]interface{} {
	pid := regexp.MustCompile(`detail/(\w+)`).FindAllStringSubmatch(url, -1)
	if len(pid) < 1 || len(pid[0]) < 2 {
		return Echo(400, "参数错误", nil)
	}
	resp := HttpPost("https://share.izuiyou.com/api/post/detail", `{"pid":`+pid[0][1]+`}`)
	respJson := gjson.Parse(resp)

	if respJson.Get("ret").String() != "1" {
		return Echo(400, respJson.Get("msg").String(), nil)
	}

	if respJson.Get("data.post.imgs.0.id").String() == "" || respJson.Get("data.post.content").String() == "" {
		return Echo(400, "解析失败", nil)
	}

	id := respJson.Get("data.post.imgs.0.id").String()
	echoMap := make(map[string]interface{})
	echoMap["text"] = respJson.Get("data.post.content").String()
	echoMap["cover"] = respJson.Get("data.post.videos." + id + ".cover_urls.0").String()
	echoMap["playAddr"] = respJson.Get("data.post.videos." + id + ".url").String()
	return Echo(200, "", echoMap)
}

// 火山
func huoShan(url string) map[string]interface{} {
	resp := HttpGet(url, pc_ua)
	json := regexp.MustCompile(`create\({d:(.*?)}\);`).FindAllStringSubmatch(resp, -1)
	if len(json) < 1 || len(json[0]) < 2 {
		return Echo(400, "参数错误", nil)
	}
	respJson := gjson.Parse(json[0][1])
	video_id := respJson.Get("video.uri").Str
	cover := respJson.Get("video.cover.url_list.0").Str
	playAddr := HttpGetLocationUrl("http://hotsoon.snssdk.com/hotsoon/item/video/_playback/?video_id="+video_id, pc_ua)
	if playAddr == "" {
		return Echo(400, "解析错误", nil)
	}
	echoMap := make(map[string]interface{})
	echoMap["cover"] = cover
	echoMap["playAddr"] = playAddr
	return Echo(200, "", echoMap)
}

// 快手
func kuaiShou(url string) map[string]interface{} {
	resp := HttpGet(url, pc_ua)
	photoId := regexp.MustCompile(`"photoId":"(.*?)",`).FindAllStringSubmatch(resp, -1)
	if len(photoId) < 1 || len(photoId[0]) < 2 {
		return Echo(400, "参数错误", nil)
	}
	resp = HttpGet("https://api.kmovie.gifshow.com/rest/n/kmovie/app/photo/getPhotoById?WS&jjh_yqc&ws&photoId="+photoId[0][1], pc_ua)
	respJson := gjson.Parse(resp)
	text := respJson.Get("photo.caption").Str
	cover := respJson.Get("photo.coverUrl").Str
	playAddr := respJson.Get("photo.mainUrl").Str
	if playAddr == "" {
		return Echo(400, "解析错误", nil)
	}
	echoMap := make(map[string]interface{})
	echoMap["text"] = text
	echoMap["cover"] = cover
	echoMap["playAddr"] = playAddr
	return Echo(200, "", echoMap)
}
