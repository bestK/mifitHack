package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
	"github.com/tencentyun/scf-go-lib/events"
	"net/http"
	"net/url"
	"strings"
)

type LoginResponse struct {
	IsBase64Encoded bool    `json:"isBase64Encoded"`
	StatusCode      int     `json:"statusCode"`
	Headers         Headers `json:"headers"`
	Body            string  `json:"body"`
}

type MyResponse struct {
	IsSuccess bool   `json:"isSuccess"`
	Code      int    `json:"code"`
	Data      string `json:"data"`
}

type Headers struct {
	ContentType string `json:"Content-Type"`
}

func Login(ctx context.Context, event events.APIGatewayRequest) (LoginResponse, error) {

	phone := event.QueryString["phone"][0]
	password := event.QueryString["password"][0]

	fmt.Println("LOGIN PARAMS", phone, password)
	loginUrl := "https://api-user.huami.com/registrations/+86" + phone + "/tokens"

	payload := strings.NewReader("state=REDIRECTION&client_id=HuaMi&redirect_uri=https%253A%252F%252Fs3-us-west-2.amazonws.com%252Fhm-registration%252Fsuccesssignin.html&token=access&password=" + password)

	clt := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := clt.Post(loginUrl, "application/x-www-form-urlencoded", payload)
	if err != nil {
		fmt.Print("请求失败->", err.Error())
	}
	location := resp.Header.Get("Location")
	fmt.Println("尝试获取跳转响应头 Location 字段", location)
	u, _ := url.Parse(location)
	m, _ := url.ParseQuery(u.RawQuery)
	status := 400
	authCoe := m.Get("access")
	if authCoe != "" {
		status = 200
	}
	headers := Headers{"application/json"}
	body := setResponse(authCoe)
	fmt.Println("组装返回结果", body)
	return LoginResponse{
		false,
		status,
		headers,
		body,
	}, nil
}

func setResponse(authCode string) string {
	isSuccess := authCode != ""
	mr, _ := json.Marshal(MyResponse{
		isSuccess,
		200,
		authCode,
	})
	return string(mr)
}

func main() {
	cloudfunction.Start(Login)
}
