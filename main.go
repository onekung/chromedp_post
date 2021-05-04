package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/fetch"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type (

	//ChromeResponse -
	cResponse struct {
		Status  int64
		URL     string
		Headers network.Headers
		Body    []byte
	}
	//ChromeRequest -
	cRequest struct {
		URL     string
		Method  string
		Headers map[string]string
		Body    []byte
	}
)

var dir = ""

func main() {

	var err error = nil
	dir, err = ioutil.TempDir("", "chromedp-example")
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if dir != "" {
			os.RemoveAll(dir)
		}

	}()

	resp, err := ChromeRequest(&cRequest{
		URL:    "https://api.myip.com",
		Method: "GET",
	})

	if err != nil {
		log.Fatalf("ChromeRequest GET: %v\n", err)
	}
	fmt.Println("GET RESPONSE")
	fmt.Println(string(resp.Body))

	resp = nil

	//Test POST

	header := make(map[string]string)
	header["Content-Type"] = "application/json; charset=utf-8"

	body := []byte(`{ "test": "testdata"}`)

	resp, err = ChromeRequest(&cRequest{
		URL:     "https://httpbin.org/post",
		Method:  "POST",
		Headers: header,
		Body:    body,
	})

	if err != nil {
		log.Fatalf("ChromeRequest POST: %v\n", err)
	}

	fmt.Println("POST RESPONSE")
	fmt.Println(string(resp.Body))
	resp = nil
	header = nil
}

func RunWithTimeOut(ctx *context.Context, timeout time.Duration, tasks chromedp.Tasks) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		timeoutContext, cancel := context.WithTimeout(ctx, timeout*time.Second)
		defer cancel()
		return tasks.Do(timeoutContext)
	}
}

//ChromeRequest -
func ChromeRequest(req *cRequest) (*cResponse, error) {
	var body string
	var res *network.Response = nil
	var ctx context.Context
	var cancel context.CancelFunc
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.DisableGPU,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoFirstRun,
		//chromedp.NoSandbox,
		chromedp.Flag("blink-settings", "imagesEnabled=false"),
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("restore-on-startup", false),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("window-size", "1,1"),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("mute-audio", true),
		chromedp.UserDataDir(dir),
		//chromedp.ProxyServer("http://192.168.1.100:8181"),
	)

	ctx, cancel = chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	if err := chromedp.Run(ctx,
		network.Enable(),
		fetch.Enable(),

		//network.SetExtraHTTPHeaders(ConvertHeaderToMap(req.Header)),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var reqID network.RequestID
			chromedp.ListenTarget(ctx, func(ev interface{}) {
				switch ev.(type) {

				case *network.EventRequestWillBeSent:
					reqEvent := ev.(*network.EventRequestWillBeSent)
					if _, ok := reqEvent.Request.Headers["Referer"]; !ok {
						if strings.HasPrefix(reqEvent.Request.URL, "http") {
							reqID = reqEvent.RequestID
						}
					}

				case *network.EventResponseReceived:
					if resEvent := ev.(*network.EventResponseReceived); resEvent.RequestID == reqID {
						res = resEvent.Response
					}

				case *fetch.EventRequestPaused:
					reqEvent := ev.(*fetch.EventRequestPaused)
					go func() {
						c := chromedp.FromContext(ctx)
						ctx := cdp.WithExecutor(ctx, c.Target)
						if reqEvent.ResourceType == network.ResourceTypeImage || reqEvent.ResourceType == network.ResourceTypeOther {
							fetch.FailRequest(reqEvent.RequestID, network.ErrorReasonBlockedByClient).Do(ctx)
						} else {

							headers := make([]*fetch.HeaderEntry, 0)
							withHeader := false

							if req.Headers != nil && len(req.Headers) > 0 {
								withHeader = true
								for k, v := range req.Headers {
									headers = append(headers, &fetch.HeaderEntry{Name: k, Value: v})
								}
							}
							f := fetch.ContinueRequest(reqEvent.RequestID).WithMethod(req.Method)
							if withHeader {
								f = f.WithHeaders(headers)
							}
							if req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE" {
								if req.Body != nil && len(req.Body) > 0 {
									f = f.WithPostData(base64.StdEncoding.EncodeToString(req.Body))
								}
							}
							f.Do(ctx)
							headers = nil
						}
					}()
				}

			})
			return nil
		}),

		RunWithTimeOut(&ctx, 10, chromedp.Tasks{
			chromedp.Navigate(req.URL),
		}),

		chromedp.WaitReady(":root"),
		chromedp.ActionFunc(func(ctx context.Context) error {

			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}

			body, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}),
	); err != nil {

		chromedp.Stop()
		ctx.Done()
		return nil, err
	}

	chromedp.Stop()
	ctx.Done()
	return &cResponse{
		Status:  res.Status,
		URL:     req.URL,
		Headers: res.Headers,
		Body:    []byte(body),
	}, nil
}
