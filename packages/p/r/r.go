package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Request struct {
	Name string `json:"name"`
}

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

func Main(args map[string]interface{}) *Response {

	path := args["__ow_path"].(string)

	url, statuscode, err := redirect(path)

	if err != nil {
		return &Response{
			StatusCode: int(statuscode),
			Body:       fmt.Sprintf("%s", err),
		}
	}

	headers := make(map[string]string)
	headers["location"] = url

	return &Response{
		Headers:    headers,
		StatusCode: int(http.StatusFound),
	}
}

func redirect(path string) (string, uint, error) {

	log.Infof("Received URL: >%s<", path)

	url, parseErr := url.Parse(path)
	if parseErr != nil {
		log.Errorf("Unable to parse received URL: >%s<", path)
		return "", http.StatusInternalServerError, fmt.Errorf("Unable to parse received URL: >%s<", path)
	}

	log.Debugf("Parsed URL: >%s<", url)

	newURL, assembleErr := assembleNewURL(url)
	if assembleErr != nil {
		log.Errorf("Unable to assemble URL from: >%s< - %s", url, assembleErr)
		return "", http.StatusBadRequest, fmt.Errorf("Unable to assemble URL from: >%s< - %s", path, assembleErr)
	}

	log.Infof("Redirecting to: >%s<", newURL)

	return newURL, http.StatusFound, nil
}

func assembleNewURL(url *url.URL) (string, error) {

	s := strings.SplitN(url.Path, "/", 3)

	log.Debugf("Parsed following parts: >%#v<", s)

	// 0 is empty because we split on "/" and the URL begins with "/"
	// 1 == version
	// 2 == fragment

	if len(s) != 3 {
		err := fmt.Errorf("insufficient parts in provided url %q", s)
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
