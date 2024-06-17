# Golang bindings for the Telegram Bot API supports custom DNS and telegram api hosts.

[![Go Reference](https://pkg.go.dev/badge/github.com/go-telegram-bot-api/telegram-bot-api/v5.svg)](https://pkg.go.dev/github.com/go-telegram-bot-api/telegram-bot-api/v5)

基于[go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)修改的版本，以支持自定义telegram api host和自定义DNS解析：

-   `func NewBotAPI(token string) (*BotAPI, error)`

-   `func NewBotAPIWithDNS(token string, dnsServers []string) (*BotAPI, error)`

    传入一组自定义DNS服务器用来作为域名解析DNS服务器

-   `func NewBotAPIWithHosts(token string, hosts []string) (*BotAPI, error)`

    传入一组自定义的telegram bot hosts，每次请求时随机使用一个来替换`api.telegram.org`

-   `func NewBotAPIWithHostsAndDNS(token string, hosts []string, dnsServers []string) (*BotAPI, error)`

    同时指定一组自定义DNS服务器地址和一组telegram bot hosts

用法：在`go.mod`中添加然后执行`go mod tidy`：

`require github.com/artlibs/telegram-bot-api v1.0.0`

## Example

First, ensure the library is installed and up to date by running
`go get -u github.com/artlibs/telegram-bot-api`.

This is a very simple bot that just displays any gotten updates,
then replies it to that chat.

```go
package main

import (
	"log"

	tgbotapi "github.com/artlibs/telegram-bot-api"
)

func main() {
    dnsServers := []string{"8.8.8.8:53",}
	bot, err := tgbotapi.NewBotAPIWithDNS("MyAwesomeBotToken", dnsServers)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)
		}
	}
}
```

If you need to use webhooks (if you wish to run on Google App Engine),
you may use a slightly different method.

```go
package main

import (
	"log"
	"net/http"

	"github.com/artlibs/telegram-bot-api"
)

func main() {
    dnsServers := []string{"8.8.8.8:53",}
	bot, err := tgbotapi.NewBotAPIWithDNS("MyAwesomeBotToken", dnsServers)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	wh, _ := tgbotapi.NewWebhookWithCert("https://www.example.com:8443/"+bot.Token, "cert.pem")

	_, err = bot.Request(wh)
	if err != nil {
		log.Fatal(err)
	}

	info, err := bot.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("Telegram callback failed: %s", info.LastErrorMessage)
	}

	updates := bot.ListenForWebhook("/" + bot.Token)
	go http.ListenAndServeTLS("0.0.0.0:8443", "cert.pem", "key.pem", nil)

	for update := range updates {
		log.Printf("%+v\n", update)
	}
}
```

If you need, you may generate a self-signed certificate, as this requires
HTTPS / TLS. The above example tells Telegram that this is your
certificate and that it should be trusted, even though it is not
properly signed.

    openssl req -x509 -newkey rsa:2048 -keyout key.pem -out cert.pem -days 3560 -subj "//O=Org\CN=Test" -nodes

Now that [Let's Encrypt](https://letsencrypt.org) is available,
you may wish to generate your free TLS certificate there.
