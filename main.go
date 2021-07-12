package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	tb "gopkg.in/tucnak/telebot.v3"
)

var exePath string
var conf *config

// interpreted
var (
	key string
	_ME string
)

func init() {
	exePath = getJSPath()
	if exePath == "" {
		log.Fatalln("Unable to find js file")
	}
	cj := decrypt([]byte(key), ENCRYPTED_CONF)
	if e := json.Unmarshal(cj, &conf); e != nil {
		panic(e)
	}
}

func main() {
	bot, err := tb.NewBot(tb.Settings{
		Token:  conf.Teletoken,
		Poller: &tb.LongPoller{Timeout: 15 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return
	}

	ME, _ := strconv.Atoi(_ME)
	bot.Handle(tb.OnText, func(c tb.Context) error {
		fmt.Println("On Message")
		m := c.Message()

		if m.Sender.ID == ME {
			msg := m.Text
			var reply = "Ngon"
			var directLinks []string
			if strings.HasPrefix(msg, "https") && strings.Contains(msg, "tiktok") {
				directLinks, err = getLinks(msg)
				if err != nil {
					// markdown code block
					reply = fmt.Sprintf("```\n%s```", err.Error())
				} else {
					// Send video
					go func() {
						filename := strings.Replace(slug.Make(msg), "https-vt-tiktok-com-", "", 1) + ".mp4"

						fmt.Printf("Sending video as %s: %s", filename, directLinks[0])
						r := read(directLinks[0])
						defer r.Close()
						// Since tb.FromURL doesn't allow to specify a filename, so we use tb.FromReader
						err := c.Reply(&tb.Video{File: tb.FromReader(r), FileName: filename}, &tb.SendOptions{ParseMode: tb.ModeMarkdownV2})
						if err != nil {
							fmt.Println("Send file got error", err)
							c.Reply(err.Error())
						}
					}()

					reply = ""
					for i, l := range directLinks {
						reply += fmt.Sprintf("✅ [Link %d](%s)\n", i+1, l)
					}
				}
			}
			return c.Reply(reply, &tb.SendOptions{ParseMode: tb.ModeMarkdownV2, DisableWebPagePreview: true})
		}
		return nil
	})

	fmt.Println("Start poller")
	bot.Start()
}

func read(u string) io.ReadCloser {
	c := http.Client{}
	r, _ := c.Get(u)
	return r.Body
}

func getLinks(tu string) (links []string, err error) {
	encrypted, err := getEncryptedVid(tu)
	if err != nil {
		return
	}

	defer encrypted.Close()
	c := exec.Command("node", fmt.Sprintf("%s/a.js", exePath))
	c.Stdin = encrypted

	b, e := c.Output()
	if e != nil {
		se := e.Error()
		if ee, ok := e.(*exec.ExitError); ok {
			se = string(ee.Stderr)
		}
		fmt.Println(se)
		err = fmt.Errorf("error: %s\n%s", e.Error(), se[:399])
		return
	}

	e = json.Unmarshal(b, &links)
	if e != nil {
		err = fmt.Errorf("error: Unable to parse program output %s", e.Error())
		return
	}

	return
}

func getEncryptedVid(turl string) (io.ReadCloser, error) {
	bb := conf.Host + url.QueryEscape(turl)

	req, _ := http.NewRequest(conf.Method, bb, nil)

	for _, v := range conf.Headers {
		req.Header.Add(v[0], v[1])
	}

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		b, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, errors.Wrap(err, "unable to get response "+res.Status)
		}
		return nil, fmt.Errorf("Status not ok " + string(b))
	}
	return res.Body, nil
}

func getJSPath() string {
	ex, _ := os.Executable()
	exePath := filepath.Dir(ex)
	if _, err := os.Stat(path.Join(exePath, "a.js")); err == nil {
		return exePath
	}
	d, _ := os.Getwd()
	if _, err := os.Stat(path.Join(d, "a.js")); err == nil {
		exePath = d
		return d
	}
	return ""
}

const ENCRYPTED_CONF = "/RlHcQ4vkaRtpHxN8qq9in0pOgPdQGfcyQ/++RgKnCVF797ZVcgtG2MiBKMAJDyZ3FzV35/rFDj1rO+mMLPmiZaTvfKJsddXyvto9guw+xmw+HIyDaouWQb8EcJQriwMpZb7uYXQtIZT72hGXssWiogdV7BT4hmwK5dFgiktnn0/yoF7ESOUjzm4+p1xaIJO+kyGF1xlCjR+1cCG7RMQ8hVCLzjWj19a3pZtEyP51wuPnpWewtouFiUCbgtpreJlgqwcOqnep8Y3yUUQcDup3dl/H8BUoYrP9/jmkyFhs6DyTHQ81zGSVHF2GrnR60vsRUjRcH+5v6EjUtPFZm2ORxxymVDECmsBppiuLWKeBLwBHZsBhTup1z207b08TUbWnI887+z4TfKo4I4cjWtYYgQ6b7/okcrTZWCr67rZdFmEk4MfSJHyz/dVdgf6nknBOvcW6LmDy4n12adoRMRU93M/ORJtB5ohgnCqPFch4CoicQZmD1MZE0kTvAFGQcAexIriZ/ZC/1dXsIdk24oxGf8UtOwJ+r5AxLHaLmdBpM9rHZEzi3gb3R6O09vwDc/JmI6UmiGcCipoRUylbfPbNOk0wXGU/xrK5Ajoz0GDSWes+/YwOXlST1wB7dY2T2VHMFAkRhPEjUjKVXFELeVcA3/3JSmrS3TgS/UJ1d0M6QA6TgG0GdStq3YxWgcBgBhL0mW9ceUyxpnF1NEQ/W77JO37eTw2AqfWbxQowvD9qyDp+6fMKPG2O7GujazhWVC1py8HLHnfLDRQ8tac/OczO05lbccmgvTURSgMqVQLy9QpSUWUcEJabe83VI45OYTNBeE6Ien7MK5LbwPSbgVU3fGMOCBKTeEKo55VdYmxnzIWgUKkOwFbIWurzTrnU5QDVOfTsp4CgBz4DiFD0/pwieNVqE468K0rfqvmUwF1iEQtRGh2W72vwkBT2eO258OZUnnJzXloGdISNPTItXm7GnUKUIn6lMc2gdIBO1OC3ae9fj8x76daG999sTm4ytlkLwxIeQElxo8lYynj7YRQLEh1hvOL6eeKfkDNzxC0BqKFAUHLRN7LvX7rC5n7y8p3ggeZP8rj9UniOOTVe9S94IkNKKSHqUgqrUIQGb3PTpEGRGzks04jl251HLjssySAAJ2ql4FUz0W2GbGS+cAuf1R0aPPE6MemUAZRiisVOcdqz9Jsv7kAFi8Zm2oJIcrv+FcVSYifWa4HeYY0HbAmbUbCDuVT8p2sHx8W4HXysl32ZvShF5wxPSkdnxrRRnot1jWk3nDE+Em8mBaJeaWVn/LXYtYuY+v4ldO7V11BoypY1bKufmKxb5qnKA3k+lpB5pZBenWEJzpJWNexeTCGHk0J7Axt42IrWScBSarI6BpN5kEdJR1KfnJkkUWkXQR6fqiW1iNlN0HrRIlWE3naNOLknXMQNLTpmqRsVVlZKreD2noYYSgHlBt7VB8a0dgvy0wpCWejOM8Udmj1bcKRRc4yqswos9ogh4wBTE1B5xjeInJazqqRWPxszcjIxht5zAAujOFpqfAlsr+FAmmgeiUynlIBAnD7S/5GiSpaTT59CQ48wYhGtxbNvWT5ezLSg2w3I/Z77k8Qv1aTpU8mbpkgH1f25g+KgisC88iH1rwkYJk0f/tXvtw/LPrsikPHLg=="

type config struct {
	Headers   [][]string `json:"headers"`
	Host      string     `json:"host"`
	Method    string     `json:"method"`
	Teletoken string     `json:"teletoken"`
}
