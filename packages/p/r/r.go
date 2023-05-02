package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

var logger = logrus.New()

/*
func main() {
	args := make(map[string]interface{})
	headers := make(map[string]interface{})

	headers["user-agent"] = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_6) AppleWebKit/537.36 (KHTML, like Gecko)"
	headers["do-connecting-ip"] = "192.168.1.2"
	headers["referer"] = "https://pxy.fi/p/r/4/wall"

	args["__ow_path"] = "/4/wall"
	args["__ow_headers"] = headers

	resp := Main(args)

	fmt.Printf("Response: %#v\n", *resp)
}
*/

func Main(args map[string]interface{}) *Response {

	logger.SetLevel(logrus.DebugLevel)

	if os.Getenv("LOG_LEVEL") != "" {
		if os.Getenv("LOG_LEVEL") == "debug" {
			logger.SetLevel(logrus.DebugLevel)
		}
	}

	fmt.Printf("Our log level: %s\n", logger.GetLevel())

	userAgent := ""
	ip := ""
	referer := ""
	path := ""
	requestID := ""

	if args["__ow_path"].(string) != "" {
		path = args["__ow_path"].(string)
	}

	requestHeaders := args["__ow_headers"]
	val, _ := requestHeaders.(map[string]interface{})

	if val["user-agent"].(string) != "" {
		userAgent = val["user-agent"].(string)
	}

	if val["do-connecting-ip"].(string) != "" {
		ip = val["do-connecting-ip"].(string)
	}

	if val["referer"].(string) != "" {
		referer = val["referer"].(string)
	}

	if val["x-request-id"].(string) != "" {
		requestID = val["x-request-id"].(string)
	}

	logger.WithFields(logrus.Fields{
		"ip":         ip,
		"user-agent": userAgent,
		"referer":    referer,
		"request-id": requestID,
	}).Infof("Running with log level: %s", logger.GetLevel())

	logger.WithFields(logrus.Fields{
		"ip":         ip,
		"user-agent": userAgent,
		"referer":    referer,
		"request-id": requestID,
	}).Infof("Received URL: >%s< via >%s<", path, referer)

	emitHeartbeat()

	url, err := parseRedirectURL(path)

	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError, // 500
			Body:       err.Error(),
		}
	}

	targetURL, err := assembleTargetURL(url)

	if err != nil {
		return &Response{
			StatusCode: http.StatusBadRequest, // 400
			Body:       err.Error(),
		}
	}

	headers := make(map[string]string)
	headers["location"] = targetURL

	logger.Infof("Redirecting to: >%s<", targetURL)

	return &Response{
		Headers:    headers,
		StatusCode: http.StatusPermanentRedirect, // 308
	}
}

func parseRedirectURL(path string) (*url.URL, error) {

	redirectURL, parseErr := url.Parse(path)

	if parseErr != nil {
		logger.Errorf("Unable to parse received URL: >%s<", path)
		return nil, fmt.Errorf("Unable to parse received URL: >%s<", path)
	}

	logger.Debugf("Parsed URL: >%s<", redirectURL)

	return redirectURL, nil
}

func assembleTargetURL(url *url.URL) (string, error) {

	s := strings.SplitN(url.Path, "/", 3)

	logger.Debugf("Parsed following parts: >%#v<", s)

	// 0 is empty because we split on "/" and the URL begins with "/"
	// 1 == version
	// 2 == fragment

	url.Host = "pxy.fi"
	url.Scheme = "https"

	if len(s) < 3 {
		logger.Errorf("insufficient parts in provided url: >%s<", url.String())

		// Example:
		// https://pxy.fi/p/r/5
		err := fmt.Errorf("<p>You only made it this far, because the specified URL has insufficient parts to redirect to the documentation</p><p>%s://%s/p/r/<span class=\"my-times\">%s<span></p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: https://pxy.fi/p/r/13/wall</p><p>See more information at: <a href=\"https://github.com/jonasbn/pxy-redirect-ow-function\">GitHub</a></p>", url.Scheme, url.Host, s[1])

		return "", err
	}

	_, err := strconv.Atoi(s[1])
	if err != nil {
		logger.Errorf("first part of url: >%s< is not a number: %q", url.String(), s)

		// Example:
		// https://pxy.fi/p/r/X
		err := fmt.Errorf("<p>You only made it this far, because the specified URL requires a version number as the first part to redirect to the documentation</p><p>%s://%s/p/r/<span class=\"my-times\">%s</span>/%s</p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: https://pxy.fi/p/r/13/wall</p><p>See more information at: <a href=\"https://github.com/jonasbn/pxy-redirect-ow-function\">GitHub</a></p>", url.Scheme, url.Host, s[1], s[2])
		return "", err
	}

	if s[2] == "" {
		logger.Errorf("second part of url: >%s< is not a string: %q", url.String(), s)

		// Example:
		// https://pxy.fi/p/r/0
		err := fmt.Errorf("<p>You only made it this far, because the specified URL requires a part to indicatede the fragment as the second part to redirect to the documentation</p>%s://%s/p/r/%s/<span class=\"my-times\">%s</span></p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: https://pxy.fi/p/r/13/wall</p><p>See more information at: <a href=\"https://github.com/jonasbn/pxy-redirect-ow-function\">GitHub</a></p>", url.Scheme, url.Host, s[1], "x")
		return "", err
	}

	return fmt.Sprintf("https://releases.llvm.org/%s.0.0/tools/clang/docs/DiagnosticsReference.html#%s", s[1], s[2]), nil
}

func emitHeartbeat() {
	logger.Debug("Emitting heartbeat")

	heartbeatToken := os.Getenv("HEARTBEAT_TOKEN")

	url := fmt.Sprintf("https://betteruptime.com/api/v1/heartbeat/%s", heartbeatToken)

	resp, err := http.Get(url)

	if err != nil {
		logger.Errorf("Unable to emit heartbeat: %s", err)
	}

	if resp.StatusCode != 200 {
		logger.Errorf("Emitted heartbeat failed: %s", resp.Status)
	}
	defer resp.Body.Close()
}
