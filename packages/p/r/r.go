package main

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Response struct {
	StatusCode int               `json:"statusCode,omitempty"`
	Headers    map[string]string `json:"headers,omitempty"`
	Body       string            `json:"body,omitempty"`
}

const tpl = `
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8" />
    <title>CodePen - error, success, warning and alert Messages</title>
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/twitter-bootstrap/4.1.3/css/bootstrap.min.css"
    />
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/5.11.2/css/all.min.css"
    />
    <link
      rel="stylesheet"
      href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome-animation/0.2.1/font-awesome-animation.min.css"
    />
    <style>
      body {
        background: #000e29;
      }

      .alert > .start-icon {
        margin-right: 0;
        min-width: 20px;
        text-align: center;
      }

      .alert > .start-icon {
        margin-right: 5px;
      }

      .greencross {
        font-size: 18px;
        color: #25ff0b;
        text-shadow: none;
      }

      .alert-simple.alert-success {
        border: 1px solid rgba(36, 241, 6, 0.46);
        background-color: rgba(7, 149, 66, 0.12156862745098039);
        box-shadow: 0px 0px 2px #259c08;
        color: #0ad406;
        text-shadow: 2px 1px #00040a;
        transition: 0.5s;
        cursor: pointer;
      }
      .alert-success:hover {
        background-color: rgba(7, 149, 66, 0.35);
        transition: 0.5s;
      }
      .alert-simple.alert-info {
        border: 1px solid rgba(6, 44, 241, 0.46);
        background-color: rgba(7, 73, 149, 0.12156862745098039);
        box-shadow: 0px 0px 2px #0396ff;
        color: #0396ff;
        text-shadow: 2px 1px #00040a;
        transition: 0.5s;
        cursor: pointer;
      }

      .alert-info:hover {
        background-color: rgba(7, 73, 149, 0.35);
        transition: 0.5s;
      }

      .blue-cross {
        font-size: 18px;
        color: #0bd2ff;
        text-shadow: none;
      }

      .alert-simple.alert-warning {
        border: 1px solid rgba(241, 142, 6, 0.81);
        background-color: rgba(220, 128, 1, 0.16);
        box-shadow: 0px 0px 2px #ffb103;
        color: #ffb103;
        text-shadow: 2px 1px #00040a;
        transition: 0.5s;
        cursor: pointer;
      }

      .alert-warning:hover {
        background-color: rgba(220, 128, 1, 0.33);
        transition: 0.5s;
      }

      .warning {
        font-size: 18px;
        color: #ffb40b;
        text-shadow: none;
      }

      .alert-simple.alert-danger {
        border: 1px solid rgba(241, 6, 6, 0.81);
        background-color: rgba(220, 17, 1, 0.16);
        box-shadow: 0px 0px 2px #ff0303;
        color: #ff0303;
        text-shadow: 2px 1px #00040a;
        transition: 0.5s;
        cursor: pointer;
      }

      .alert-danger:hover {
        background-color: rgba(220, 17, 1, 0.33);
        transition: 0.5s;
      }

      .danger {
        font-size: 18px;
        color: #ff0303;
        text-shadow: none;
      }

      .alert-simple.alert-primary {
        border: 1px solid rgba(6, 241, 226, 0.81);
        background-color: rgba(1, 204, 220, 0.16);
        box-shadow: 0px 0px 2px #03fff5;
        color: #03d0ff;
        text-shadow: 2px 1px #00040a;
        transition: 0.5s;
        cursor: pointer;
      }

      .alert-primary:hover {
        background-color: rgba(1, 204, 220, 0.33);
        transition: 0.5s;
      }

      .alertprimary {
        font-size: 18px;
        color: #03d0ff;
        text-shadow: none;
      }

      .square_box {
        position: absolute;
        -webkit-transform: rotate(45deg);
        -ms-transform: rotate(45deg);
        transform: rotate(45deg);
        border-top-left-radius: 45px;
        opacity: 0.302;
      }

      .square_box.box_three {
        background-image: -moz-linear-gradient(
          -90deg,
          #290a59 0%,
          #3d57f4 100%
        );
        background-image: -webkit-linear-gradient(
          -90deg,
          #290a59 0%,
          #3d57f4 100%
        );
        background-image: -ms-linear-gradient(-90deg, #290a59 0%, #3d57f4 100%);
        opacity: 0.059;
        left: -80px;
        top: -60px;
        width: 500px;
        height: 500px;
        border-radius: 45px;
      }

      .square_box.box_four {
        background-image: -moz-linear-gradient(
          -90deg,
          #290a59 0%,
          #3d57f4 100%
        );
        background-image: -webkit-linear-gradient(
          -90deg,
          #290a59 0%,
          #3d57f4 100%
        );
        background-image: -ms-linear-gradient(-90deg, #290a59 0%, #3d57f4 100%);
        opacity: 0.059;
        left: 150px;
        top: -25px;
        width: 550px;
        height: 550px;
        border-radius: 45px;
      }

      .alert:before {
        content: "";
        position: absolute;
        width: 0;
        height: calc(100% - 44px);
        border-left: 1px solid;
        border-right: 2px solid;
        border-bottom-right-radius: 3px;
        border-top-right-radius: 3px;
        left: 0;
        top: 50%;
        transform: translate(0, -50%);
        height: 20px;
      }

      .fa-times {
        -webkit-animation: blink-1 2s infinite both;
        animation: blink-1 2s infinite both;
      }

      /**
    * ----------------------------------------
    * animation blink-1
    * ----------------------------------------
    */
      @-webkit-keyframes blink-1 {
        0%,
        50%,
        100% {
          opacity: 1;
        }
        25%,
        75% {
          opacity: 0;
        }
      }
      @keyframes blink-1 {
        0%,
        50%,
        100% {
          opacity: 1;
        }
        25%,
        75% {
          opacity: 0;
        }
      }
    </style>
  </head>
  <body>
    <!-- partial:index.partial.html -->
    <section>
      <div class="square_box box_three"></div>
      <div class="square_box box_four"></div>
      <div class="container mt-5">
        <div class="row">
		{{if eq .PageType "error"}}
		<div class="col-sm-12">
		<div
		class="alert fade alert-simple alert-danger alert-dismissible text-left font__family-montserrat font__size-16 font__weight-light brk-library-rendered rendered show"
		role="alert"
		data-brk-library="component__alert"
		>
		<button
			type="button"
			class="close font__size-18"
			data-dismiss="alert"
		>
			<span aria-hidden="true">
			<i class="fa fa-times danger"></i>
			</span>
			<span class="sr-only">Close</span>
		</button>
		<i class="start-icon far fa-times-circle faa-pulse animated"></i>
		<strong class="font__weight-semibold">Oh snap!</strong> {{ .Message }}.
		</div>
		</div>
		{{else if eq .PageType "warning"}}
		<div class="col-sm-12">
		<div
		class="alert fade alert-simple alert-warning alert-dismissible text-left font__family-montserrat font__size-16 font__weight-light brk-library-rendered rendered show"
		role="alert"
		data-brk-library="component__alert"
		>
		<button
			type="button"
			class="close font__size-18"
			data-dismiss="alert"
		>
			<span aria-hidden="true">
			<i class="fa fa-times warning"></i>
			</span>
			<span class="sr-only">Close</span>
		</button>
		<i
			class="start-icon fa fa-exclamation-triangle faa-flash animated"
		></i>
		<strong class="font__weight-semibold">Warning!</strong> {{ .Message }}
		</div>
		</div>
		{{else}}
		<div class="col-sm-12">
		<div
		  class="alert fade alert-simple alert-info alert-dismissible text-left font__family-montserrat font__size-16 font__weight-light brk-library-rendered rendered show"
		  role="alert"
		  data-brk-library="component__alert"
		>
		  <button
			type="button"
			class="close font__size-18"
			data-dismiss="alert"
		  >
			<span aria-hidden="true">
			  <i class="fa fa-times blue-cross"></i>
			</span>
			<span class="sr-only">Close</span>
		  </button>
		  <i class="start-icon fa fa-info-circle faa-shake animated"></i>
		  <strong class="font__weight-semibold">Heads up!</strong> {{ .Message }}.
		</div>
		</div>
		{{end}}
        </div>
      </div>
    </section>
    <!-- partial -->
  </body>
</html>`

type TmplData struct {
	Message  string
	PageType string
}

func Main(args map[string]interface{}) *Response {

	log.SetLevel(log.DebugLevel)

	path := args["__ow_path"].(string)

	url, err := parseRedirectURL(path)

	if err != nil {

		b, renderErr := renderPage(err.Error(), "error")

		if renderErr != nil {
			return &Response{
				StatusCode: http.StatusInternalServerError,
				Body:       renderErr.Error(),
			}
		}

		return &Response{
			StatusCode: http.StatusInternalServerError,
			Body:       b.String(),
		}
	}

	if url.String() == "/" || url.String() == "/index.html" {
		log.Infof("Non-redirectable URL >%s< served", url.String())

		b, renderErr := renderPage("hello world", "info")

		if renderErr != nil {
			return &Response{
				StatusCode: http.StatusInternalServerError,
				Body:       renderErr.Error(),
			}
		}

		return &Response{
			StatusCode: http.StatusOK,
			Body:       b.String(),
		}
	}

	targetURL, err := redirect(url)

	if err != nil {

		b, renderErr := renderPage(err.Error(), "warning")

		if renderErr != nil {
			return &Response{
				StatusCode: http.StatusInternalServerError,
				Body:       renderErr.Error(),
			}
		}

		return &Response{
			StatusCode: http.StatusBadRequest,
			Body:       b.String(),
		}
	}

	headers := make(map[string]string)
	headers["location"] = targetURL

	log.Infof("Redirecting to: >%s<", targetURL)

	return &Response{
		Headers:    headers,
		StatusCode: http.StatusFound,
	}
}

func renderPage(message string, pagetype string) (bytes.Buffer, error) {

	var b bytes.Buffer

	tmpl, err := template.New("").Parse(tpl)
	if err != nil {
		log.Errorf("Unable to parse HTML template: %s", err.Error())
		return b, err
	}

	data := TmplData{Message: message, PageType: pagetype}

	err = tmpl.Execute(&b, &data)

	if err != nil {
		log.Errorf("Unable to render HTML page: %s", err.Error())
		return b, err
	}

	return b, nil
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
