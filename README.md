# chromedp_post
golang chromedp library post data to chrome browser

#Response

GET RESPONSE
<html><head></head><body>{"ip":"xxx.x.xxx.xx","country":"Thailand","cc":"TH"}</body></html>


POST RESPONSE
<html><head></head><body><pre style="word-wrap: break-word; white-space: pre-wrap;">{
  "args": {}, 
  "data": "{ \"test\": \"testdata\"}", 
  "files": {}, 
  "form": {}, 
  "headers": {
    "Accept-Encoding": "gzip, deflate, br", 
    "Accept-Language": "en-US,en;q=0.9", 
    "Content-Length": "21", 
    "Content-Type": "application/json; charset=utf-8", 
    "Host": "httpbin.org", 
    "Sec-Fetch-Dest": "document", 
    "Sec-Fetch-Mode": "navigate", 
    "Sec-Fetch-Site": "none", 
    "Sec-Fetch-User": "?1", 
    "User-Agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/90.0.4430.93 Safari/537.36", 
    "X-Amzn-Trace-Id": "Root=1-6088cf90-3affc1070096b40305cc8087"
  }, 
  "json": {
    "test": "testdata"
  }, 
  "origin": "xxx.x.xxx.xx", 
  "url": "https://httpbin.org/post"
}
</pre></body></html>