package main

import (
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	//extern dependency
	"github.com/yanzay/tbot"
	"github.com/yanzay/tbot/model"
)

//Block Explorer
const apibe = "https://be.martexcoin.org/ext/getbalance/"

//Max to store
const MAX_DB = 1024
const MAX_DB_BUY = 5
const MAX_DB_SELL = 5

//Masternode balance
const MN_BALANCE = 5000.0

type Sell struct {
	id int
	//identificacao
	xns     string
	contato string

	//preco
	mxt_price float64
	brl_price float64

	end bool
}

type Buy struct {
	id int
	//identificacao
	contato string

	//preco
	mxt_price float64

	end bool
}

//Data from users
type Dados struct {
	//identificacao
	charid     string
	charwallet string

	//dados
	taxa        float64
	porcentagem float64
	last_rent   float64
	rendimento  float64
	balance     float64

	//se foi configurado
	isSet bool
	sells []Sell
	sellc int
	buys  []Buy
	buyc  int
}

//DB
type Database struct {
	dados []Dados
	count int
}

//Convert float64 to str
func floattostr(fv float64) string {
	return strconv.FormatFloat(fv, 'f', 2, 64)
}

//Convert str to float64
func strtofloat(sv string) float64 {
	x, _ := strconv.ParseFloat(sv, 64)
	return x
}

//*Obter saldo da carteira na API
//*Get wallet balance
func getbalance(wall string) string {
	get, err := http.Get(apibe + wall)
	if err != nil {
		return "-1"
	}
	defer get.Body.Close()

	resp, err := ioutil.ReadAll(get.Body)
	if err != nil {
		return "-1"
	}

	return (string(resp))
}

//*comando /wallet
func wallet(m *tbot.Message) {
	carteira := m.Vars["carteira"]
	if len(carteira) < 0 {
		m.Reply("Use: /wallet <carteira>")
		return
	}
	resp := getbalance(carteira)
	if len(resp) > 0 {
		if strings.Contains(resp, "error") {
			m.Reply("Tem certeza que " + carteira + " é uma carteira? :/")
			return
		}
		m.Reply("Saldo: " + resp)

	} else {
		m.Reply("falha ao obter saldo!")
	}
}

// Temp
type MessageVars map[string]string
type MessageOption func(*model.Message)
type Message struct {
	*model.Message
	Vars         MessageVars
	replyChannel chan *model.Message
}

// Reply to user by chat id
// Será usado depois para alertas
func ReplyTo(m *tbot.Message, reply string, options ...MessageOption) {
	msg := &model.Message{
		ChatID: m.ChatID,
		Type:   model.MessageText,
		Data:   reply,
	}
	for _, option := range options {
		option(msg)
	}

	m.Replies <- msg
}
