package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

func init() {
	log.SetLevel(log.InfoLevel)

	if os.Getenv("LOG_LEVEL") != "" {
		if os.Getenv("LOG_LEVEL") == "debug" {
			log.SetLevel(log.DebugLevel)
		}
	}
}

func Main(args map[string]interface{}) *Response {

	userAgent := ""
	ip := ""
	referer := ""
	path := ""

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

	log.WithFields(log.Fields{
		"ip":         ip,
		"user-agent": userAgent,
		"referer":    referer,
	}).Infof("Running with log level: %s", log.GetLevel())

	log.Infof("Received request to redirect: >%s<", path)

	emitHeartbeat()

	url, err := parseRedirectURL(path, ip, userAgent, referer)

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

	log.Infof("Redirecting to: >%s<", targetURL)

	return &Response{
		Headers:    headers,
		StatusCode: http.StatusPermanentRedirect, // 308
	}
}

func parseRedirectURL(path, ip, userAgent, referer string) (*url.URL, error) {

	log.WithFields(log.Fields{
		"ip":         ip,
		"user-agent": userAgent,
		"referer":    referer,
	}).Info("Received URL: >" + path + "< via >" + referer)

	redirectURL, parseErr := url.Parse(path)

	if parseErr != nil {
		log.Errorf("Unable to parse received URL: >%s<", path)
		return nil, fmt.Errorf("Unable to parse received URL: >%s<", path)
	}

	log.Debugf("Parsed URL: >%s<", redirectURL)

	return redirectURL, nil
}

func assembleTargetURL(url *url.URL) (string, error) {

	s := strings.SplitN(url.Path, "/", 3)

	log.Debugf("Parsed following parts: >%#v<", s)

	// 0 is empty because we split on "/" and the URL begins with "/"
	// 1 == version
	// 2 == fragment

	url.Host = "pxy.fi"
	url.Scheme = "https"

	if len(s) < 3 {
		log.Errorf("insufficient parts in provided url: >%s<", url.String())

		// Example:
		// https://pxy.fi/p/r/5
		err := fmt.Errorf("<p>You only made it this far, because the specified URL has insufficient parts to redirect to the documentation</p><p>%s://%s/p/r/<span class=\"my-times\">%s<span></p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: https://pxy.fi/p/r/13/wall</p><p>See more information at: <a href=\"https://github.com/jonasbn/pxy-redirect-ow-function\">GitHub</a></p>", url.Scheme, url.Host, s[1])

		return "", err
	}

	_, err := strconv.Atoi(s[1])
	if err != nil {
		log.Errorf("first part of url: >%s< is not a number: %q", url.String(), s)

		// Example:
		// https://pxy.fi/p/r/X
		err := fmt.Errorf("<p>You only made it this far, because the specified URL requires a version number as the first part to redirect to the documentation</p><p>%s://%s/p/r/<span class=\"my-times\">%s</span>/%s</p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: https://pxy.fi/p/r/13/wall</p><p>See more information at: <a href=\"https://github.com/jonasbn/pxy-redirect-ow-function\">GitHub</a></p>", url.Scheme, url.Host, s[1], s[2])
		return "", err
	}

	if s[2] == "" {
		log.Errorf("second part of url: >%s< is not a string: %q", url.String(), s)

		// Example:
		// https://pxy.fi/p/r/0
		err := fmt.Errorf("<p>You only made it this far, because the specified URL requires a part to indicatede the fragment as the second part to redirect to the documentation</p>%s://%s/p/r/%s/<span class=\"my-times\">%s</span></p><p>In order to get the redirect to work, please specify both a version and a fragment</p><p>Example: https://pxy.fi/p/r/13/wall</p><p>See more information at: <a href=\"https://github.com/jonasbn/pxy-redirect-ow-function\">GitHub</a></p>", url.Scheme, url.Host, s[1], "x")
		return "", err
	}

	return fmt.Sprintf("https://releases.llvm.org/%s.0.0/tools/clang/docs/DiagnosticsReference.html#%s", s[1], s[2]), nil
}

func emitHeartbeat() {
	log.Debug("Emitting heartbeat")

	heartbeatToken := os.Getenv("HEARTBEAT_TOKEN")

	url := fmt.Sprintf("https://betteruptime.com/api/v1/heartbeat/%s", heartbeatToken)

	resp, err := http.Get(url)

	if err != nil {
		log.Errorf("Unable to emit heartbeat: %s", err)
	}

	if resp.StatusCode != 200 {
		log.Errorf("Emitted heartbeat failed: %s", resp.Status)
	}
}
