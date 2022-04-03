package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
)

// extractIP gets the real ip address from a request
// taking all custom headers into account
func extractIP(r *http.Request) (string, error) {
	ip := ""
	// use the passed ip from nginx here
	// as nginx will always be in front of this server
	// it can not be forged (except in dev)
	realIP := r.Header.Get("X-Forwarded-For")
	if realIP != "" {
		parts := strings.Split(realIP, ",")
		if len(parts) > 0 {
			ip = strings.TrimSpace(parts[0])
		} else {
			return "", fmt.Errorf("parts is 0, this should not happen: %s", realIP)
		}
	} else {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return "", fmt.Errorf("error on extracting host from ip: %w", err)
		}
		ip = host
	}

	if ip == "" {
		return "", fmt.Errorf("could not get ip")
	}
	return ip, nil
}

// copied from
// https://github.com/gorilla/handlers/blob/master/logging.go

// https://github.com/gorilla/handlers/issues/202
// https://github.com/gorilla/handlers/pull/203

const lowerhex = "0123456789abcdef"

func appendQuoted(buf []byte, s string) []byte {
	var runeTmp [utf8.UTFMax]byte
	for width := 0; len(s) > 0; s = s[width:] {
		r := rune(s[0])
		width = 1
		if r >= utf8.RuneSelf {
			r, width = utf8.DecodeRuneInString(s)
		}
		if width == 1 && r == utf8.RuneError {
			buf = append(buf, `\x`...)
			buf = append(buf, lowerhex[s[0]>>4])
			buf = append(buf, lowerhex[s[0]&0xF])
			continue
		}
		if r == rune('"') || r == '\\' { // always backslashed
			buf = append(buf, '\\')
			buf = append(buf, byte(r))
			continue
		}
		if strconv.IsPrint(r) {
			n := utf8.EncodeRune(runeTmp[:], r)
			buf = append(buf, runeTmp[:n]...)
			continue
		}
		switch r {
		case '\a':
			buf = append(buf, `\a`...)
		case '\b':
			buf = append(buf, `\b`...)
		case '\f':
			buf = append(buf, `\f`...)
		case '\n':
			buf = append(buf, `\n`...)
		case '\r':
			buf = append(buf, `\r`...)
		case '\t':
			buf = append(buf, `\t`...)
		case '\v':
			buf = append(buf, `\v`...)
		default:
			switch {
			case r < ' ':
				buf = append(buf, `\x`...)
				buf = append(buf, lowerhex[s[0]>>4])
				buf = append(buf, lowerhex[s[0]&0xF])
			case r > utf8.MaxRune:
				r = 0xFFFD
				fallthrough
			case r < 0x10000:
				buf = append(buf, `\u`...)
				for s := 12; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			default:
				buf = append(buf, `\U`...)
				for s := 28; s >= 0; s -= 4 {
					buf = append(buf, lowerhex[r>>uint(s)&0xF])
				}
			}
		}
	}
	return buf
}

func buildCustomLogLine(req *http.Request, url url.URL, ts time.Time, status, size int) []byte {
	username := "-"
	if url.User != nil {
		if name := url.User.Username(); name != "" {
			username = name
		}
	}

	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}

	// use the passed ip from nginx here
	// ignore errors here as we have no error handler
	realIP, err := extractIP(req)
	if err != nil {
		// ignore errors
		log.Error(err)
	}
	if realIP != "" {
		host = realIP
	}

	uri := req.RequestURI

	// Requests using the CONNECT method over HTTP/2.0 must use
	// the authority field (aka r.Host) to identify the target.
	// Refer: https://httpwg.github.io/specs/rfc7540.html#CONNECT
	if req.ProtoMajor == 2 && req.Method == "CONNECT" {
		uri = req.Host
	}
	if uri == "" {
		uri = url.RequestURI()
	}

	buf := make([]byte, 0, 3*(len(host)+len(username)+len(req.Method)+len(uri)+len(req.Proto)+50)/2)
	buf = append(buf, host...)
	buf = append(buf, " - "...)
	buf = append(buf, username...)
	buf = append(buf, " ["...)
	buf = append(buf, ts.Format("02/Jan/2006:15:04:05 -0700")...)
	buf = append(buf, `] "`...)
	buf = appendQuoted(buf, req.Host)
	buf = append(buf, `" "`...)
	buf = append(buf, req.Method...)
	buf = append(buf, " "...)
	buf = appendQuoted(buf, uri)
	buf = append(buf, " "...)
	buf = append(buf, req.Proto...)
	buf = append(buf, `" `...)
	buf = append(buf, strconv.Itoa(status)...)
	buf = append(buf, " "...)
	buf = append(buf, strconv.Itoa(size)...)
	buf = append(buf, ` "`...)
	buf = appendQuoted(buf, req.Referer())
	buf = append(buf, `" "`...)
	buf = appendQuoted(buf, req.UserAgent())
	buf = append(buf, '"')

	return buf
}

func customLogFormatter(writer io.Writer, params handlers.LogFormatterParams) {
	buf := buildCustomLogLine(params.Request, params.URL, params.TimeStamp, params.StatusCode, params.Size)
	buf = append(buf, '\n')
	_, err := writer.Write(buf)
	if err != nil {
		log.Errorf("error writing log: %v", err)
	}
}
