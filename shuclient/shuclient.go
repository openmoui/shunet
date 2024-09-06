package shuclient

import (
	"encoding/json"
	"fmt"
	"golang.org/x/net/context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"shunet/config"
	"shunet/rsa"
	"shunet/utils"
	"strings"
	"time"
)

var (
	topSelfLocationHrefPattern = `<script>top\.self\.location\.href='([^']*)'</script>`
	LoginSuccessPattern        = `<title>登录成功</title>`
	defaultHeader              = map[string]string{
		"User-Agent":      "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/126.0.0.0 Safari/537.36",
		"Accept":          "*/*",
		"Accept-Language": "en,zh;q=0.9,zh-CN;q=0.8",
		"Accept-Encoding": "gzip, deflate",
		"Connection":      "keep-alive",
		"DNT":             "1",
		"Pragma":          "no-cache",
		"Cache-Control":   "no-cache",
	}
	log = utils.Log
)

type Client struct {
	cfg                       *config.Config
	rsa                       *rsa.RSAPair
	header                    map[string]string
	httpClient                *http.Client
	hostUrl                   string
	successPageUrl            string
	IsLogin                   bool
	topSelfLocationHref       string
	topSelfLocationHrefParams map[string]string
	referer                   string
	interfaceDoPath           string
	userIndex                 string        // 服务器返回的用户索引
	delayTime                 time.Duration // 重试&心跳时间间隔，单位秒
}

func NewClient(c *config.Config) *Client {
	hc := &http.Client{}
	if jar, err := cookiejar.New(nil); err != nil {
		hc.Jar = jar
	}
	if len(c.Proxy) > 0 {
		// 创建一个代理地址
		proxyURL, err := url.Parse(c.Proxy)
		if err != nil {
			log.Fatal(err)
		}
		// 创建一个 Transport 实例，并配置代理
		hc.Transport = &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}
	}

	delayTime := 60 * time.Second
	if c.DelayTime > 0 {
		delayTime = time.Duration(c.DelayTime) * time.Second
	}
	return &Client{
		cfg:             c,
		header:          defaultHeader,
		hostUrl:         "http://" + c.Host,
		interfaceDoPath: "http://" + c.Host + "/eportal/InterFace.do?method=",
		httpClient:      hc,
		IsLogin:         false,
		delayTime:       delayTime,
	}
}

func setReqHeader(header map[string]string, r *http.Request) *http.Request {
	for k, v := range header {
		r.Header.Set(k, v)
	}
	return r
}

func (c *Client) EnterLoginPage() (page string, err error) {
	req, err := http.NewRequest("GET", c.hostUrl, nil)
	if err != nil {
		return "", err
	}
	req = setReqHeader(c.header, req)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	c.referer = resp.Request.URL.String()
	defer resp.Body.Close()

	// 检测返回的body是否经过压缩，并返回解压的内容
	if page, err = utils.DecodeContent(resp); err != nil {
		return "", err
	}

	if redirectURL, err := utils.Match(page, topSelfLocationHrefPattern); err != nil {
		// 登录成功获取userIndex
		if strings.Contains(page, LoginSuccessPattern) {
			c.successPageUrl = resp.Request.URL.String()
			if params, err := utils.DencodeParams(c.successPageUrl); err != nil {
				return page, err
			} else {
				if v, ok := params["userIndex"]; ok {
					c.userIndex = v
					c.IsLogin = true
					return page, nil
				} else {
					return page, fmt.Errorf("userIndex not found")
				}
			}
		} else {
			return page, err
		}
	} else {
		c.topSelfLocationHref = redirectURL
		if params, err := utils.DencodeParams(redirectURL); err != nil {
			return "", err
		} else {
			c.topSelfLocationHrefParams = params
		}
	}

	if v, ok := c.topSelfLocationHrefParams["mac"]; ok {
		c.cfg.Mac = v
	}

	if page, err = c.enterTopSelfLocation(); err != nil {
		return "", err
	}
	return page, nil
}

func (c *Client) enterTopSelfLocation() (page string, err error) {
	if c.topSelfLocationHref == "" {
		return "", fmt.Errorf("topSelfLocationHref is empty")
	}
	req, err := http.NewRequest("GET", c.topSelfLocationHref, nil)
	if err != nil {
		return "", err
	}
	req = setReqHeader(c.header, req)
	req.Header.Set("Referer", c.referer)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	c.referer = req.URL.String()
	defer resp.Body.Close()

	// 检测返回的body是否经过压缩，并返回解压的内容
	if page, err = utils.DecodeContent(resp); err != nil {
		return "", err
	}

	return page, nil
}

func (c *Client) interfaceDo(method string, formData map[string]string) (*http.Response, error) {
	data := url.Values{}
	for key, value := range formData {
		data.Set(key, value)
	}

	req, err := http.NewRequest("POST", c.interfaceDoPath+method, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req = setReqHeader(c.header, req)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Referer", c.referer)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) GetPageInfo() (*PageInfo, error) {
	if c.topSelfLocationHrefParams == nil {
		return nil, fmt.Errorf("topSelfLocationHrefParams is nil")
	}

	param := make(map[string]string, 1)
	param["queryString"] = utils.EncodeParams(c.topSelfLocationHrefParams)
	resp, err := c.interfaceDo("pageInfo", param)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	pageInfo := &PageInfo{}
	err = json.Unmarshal(body, pageInfo)
	if err != nil {
		return nil, err
	}

	if len(pageInfo.PublicKeyExponent) > 0 {
		c.cfg.PublicKeyExponent = pageInfo.PublicKeyExponent
	}
	if len(pageInfo.PublicKeyModulus) > 0 {
		c.cfg.PublicKeyModulus = pageInfo.PublicKeyModulus
	}
	if len(pageInfo.PasswordEncrypt) > 0 {
		c.cfg.PasswordEncrypt = pageInfo.PasswordEncrypt
	}

	c.rsa = rsa.NewRSAPair(c.cfg.PublicKeyExponent, "", c.cfg.PublicKeyModulus)
	return pageInfo, nil
}

func (c *Client) Login() (*LoginResponse, error) {
	if c.rsa == nil {
		return nil, fmt.Errorf("rsa is nil")
	}
	param := make(map[string]string, 8)
	param["queryString"] = utils.EncodeParams(c.topSelfLocationHrefParams)
	param["userId"] = c.cfg.UserId
	param["password"] = c.rsa.EncryptedPassword(c.cfg.Password, c.cfg.Mac)
	param["service"] = "shu"
	param["operatorPwd"] = ""
	param["operatorUserId"] = ""
	param["validcode"] = ""
	param["passwordEncrypt"] = c.cfg.PasswordEncrypt

	resp, err := c.interfaceDo("login", param)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	loginResponse := &LoginResponse{}
	if err = json.Unmarshal(body, loginResponse); err != nil {
		return nil, err
	}
	if loginResponse.Result == "success" {
		c.IsLogin = true
	}
	c.userIndex = loginResponse.UserIndex
	return loginResponse, nil
}

func (c *Client) LogOut() (*GeneralResponse, error) {
	if len(c.userIndex) == 0 {
		return nil, fmt.Errorf("userIndex is empty")
	}
	param := make(map[string]string, 1)
	param["userIndex"] = c.userIndex

	resp, err := c.interfaceDo("logout", param)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	logoutResponse := &GeneralResponse{}
	if err = json.Unmarshal(body, logoutResponse); err != nil {
		return nil, err
	}
	if logoutResponse.Result == "success" {
		c.IsLogin = false
	}
	return logoutResponse, nil
}

func (c *Client) KeepAlive() (*GeneralResponse, error) {
	if len(c.userIndex) == 0 {
		return nil, fmt.Errorf("userIndex is empty")
	}
	param := make(map[string]string, 1)
	param["userIndex"] = c.userIndex

	resp, err := c.interfaceDo("keepalive", param)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	keepAliveResponse := &GeneralResponse{}
	if err = json.Unmarshal(body, keepAliveResponse); err != nil {
		return nil, err
	}
	return keepAliveResponse, nil
}

func (c *Client) Run(ctx context.Context) {
	// 启动时记录Pid
	c.cfg.Pid = os.Getpid()
	if err := c.cfg.Save(); err != nil {
		log.Fatalf("Failed to save config: %v", err)
	}
	ticker := time.Tick(c.delayTime)
	for {
		switch c.IsLogin {
		case true:
			resp, err := c.KeepAlive()
			if err != nil {
				c.IsLogin = false
				log.Errorf("err: %+v", err)
				break
			}
			if resp.Result != "success" {
				c.IsLogin = false
			}
			log.Info("KeepAlive")
		case false:
			if _, err := c.EnterLoginPage(); err != nil {
				c.IsLogin = false
				log.Errorf("EnterLoginPage err: %+v", err)
				break
			}
			log.Info("EnterLoginPage")

			if c.IsLogin {
				log.Warning("already login, skip login")
				break
			}

			if _, err := c.GetPageInfo(); err != nil {
				c.IsLogin = false
				log.Errorf("GetPageInfo err: %v", err)
				break
			}
			log.Info("GetPageInfo")

			resp, err := c.Login()
			if err != nil {
				c.IsLogin = false
				log.Errorf("Login err: %+v", err)
				break
			}
			if resp.Result != "success" {
				c.IsLogin = false
				log.Warning("Login fail, but no err")
			} else {
				log.Info("Login success")
			}

		}
		log.Infof("Sleep %v", c.delayTime.String())
		select {
		case <-ctx.Done():
			log.Info("Client.Run Receive stop signal, Client Run exit")
			if !c.IsLogin {
				log.Info("Already logout!")
			} else {
				if _, err := c.LogOut(); err != nil {
					log.Errorf("Client.Run Logout err: %+v", err)
				}
			}
			// 退出时清空Pid
			c.cfg.Pid = 0
			if err := c.cfg.Save(); err != nil {
				log.Errorf("Client.Run Save err: %+v", err)
			}
			return
		case <-ticker:
		}
	}
}

func (c *Client) save() error {
	return c.cfg.Save()
}
