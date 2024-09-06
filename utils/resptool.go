package utils

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"fmt"
	"golang.org/x/net/html/charset"
	"io"
	"net/http"
	"net/url"
	"regexp"
)

func ConvertGBKToUTF8(body []byte) (string, error) {
	reader, err := charset.NewReaderLabel("gbk", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	utf8Bytes, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(utf8Bytes), nil
}

// 检测返回的body是否经过压缩，并返回解压的内容
func DecodeContent(res *http.Response) (s string, err error) {
	var bodyReader io.Reader
	switch res.Header.Get("Content-Encoding") {
	case "gzip":
		if bodyReader, err = gzip.NewReader(res.Body); err != nil {
			return "", err
		}
	case "deflate":
		bodyReader = flate.NewReader(res.Body)
	default:
		bodyReader = res.Body
	}

	body, err := io.ReadAll(bodyReader)
	if err != nil {
		return "", err
	}

	switch res.Header.Get("Content-Type") {
	case "text/html;charset=GBK":
		if s, err = ConvertGBKToUTF8(body); err != nil {
			return "", err
		}
	default:
		s = string(body)
	}
	return s, nil
}

func Match(str string, pattern string) (string, error) {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(str)
	if len(matches) < 2 {
		return "", fmt.Errorf("no match found")
	}
	return matches[1], nil
}

func DencodeParams(rawURL string) (map[string]string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	params := make(map[string]string)
	for key, values := range parsedURL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params, nil
}

func EncodeParams(params map[string]string) string {
	values := url.Values{}
	for key, value := range params {
		values.Set(key, value)
	}
	return values.Encode()
}
