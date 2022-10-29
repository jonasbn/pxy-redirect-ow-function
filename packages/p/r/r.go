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

func Main(args map[string]interface{}) *Response {

	path := args["__ow_path"].(string)

	url, err := parseRedirectURL(path)

	if err != nil {
		return &Response{
			StatusCode: http.StatusInternalServerError,
		}
	}

	if url.String() == "/" || url.String() == "/index.html" {
		log.Infof("Non-redirectable URL >%s< served", url.String())

		content, err := os.ReadFile("static/index.html")
		if err != nil {
			return &Response{
				StatusCode: http.StatusInternalServerError,
			}
		}
		return &Response{
			StatusCode: http.StatusOK,
			Body:       string(content),
		}
	}

	targetURL, err := redirect(url)

	if err != nil {
		return &Response{
			StatusCode: http.StatusBadRequest,
			Body:       fmt.Sprintf("%s", err),
		}
	}

	headers := make(map[string]string)
	headers["location"] = targetURL

	log.Infof("Redirecting to: >%s<", targetURL)

	return &Response{
		Headers:    headers,
		StatusCode: int(http.StatusFound),
	}
}

func parseRedirectURL(path string) (*url.URL, error) {
	log.Infof("Received URL: >%s<", path)

	redirectURL, parseErr := url.Parse(path)
	if parseErr != nil {
		log.Errorf("Unable to parse received URL: >%s<", path)
		return nil, fmt.Errorf("Unable to parse received URL: >%s<", path)
	}

	log.Debugf("Parsed URL: >%s<", redirectURL)

	return redirectURL, nil
}

func redirect(url *url.URL) (string, error) {

	redirectURL, err := assembleRedirectURL(url)

	if err != nil {
		log.Errorf("Unable to assemble URL from: >%s< - %s", url.String(), err)
		return "", fmt.Errorf("Unable to assemble URL from: >%s< - %s", url.String(), err)
	}

	return redirectURL, nil
}

func assembleRedirectURL(url *url.URL) (string, error) {

	s := strings.SplitN(url.Path, "/", 3)

	log.Debugf("Parsed following parts: >%#v<", s)

	// 0 is empty because we split on "/" and the URL begins with "/"
	// 1 == version
	// 2 == fragment

	if len(s) < 3 {
		err := fmt.Errorf("insufficient parts in provided url %q", s)
		return "", err
	}

	if len(s) > 3 {
		err := fmt.Errorf("excessive parts in provided url %q", s)
		return "", err
	}

	_, err := strconv.Atoi(s[1])
	if err != nil {
		err := fmt.Errorf("first part of url is not a number: %q", s)
		return "", err
	}

	if s[2] == "" {
		err := fmt.Errorf("second part of url is not a string: %q", s)
		return "", err
	}

	return fmt.Sprintf("https://releases.llvm.org/%s.0.0/tools/clang/docs/DiagnosticsReference.html#%s", s[1], s[2]), nil
}
