/* Proctected by MIT License: https://opensource.org/licenses/MIT
   Author Telegram: @OneFreakDay
   Date: 16/12/2017
   Description:
   A simple telegram bot to connect with a Block Explorer API cryptocoin
*/
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	//extern dependency
	"github.com/yanzay/tbot"
)

//Block Explorer
const apibe = "https://be.martexcoin.org/ext/getbalance/"

//Max users
const MAX_DB = 1024

//Masternode balance
const MN_BALANCE = 5000.0

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
}

//DB
type Database struct {
	dados []Dados
	count int
}

//...
var db Database

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
//* /wallet command
func wallet(m *tbot.Message) {
	carteira := m.Vars["carteira"]
	if len(carteira) < 0 {
		m.Reply("Use: /wallet <carteira>")
		return
	}
	resp := getbalance(carteira)
	if len(resp) > 0 {
		m.Reply("Saldo: " + resp)

	} else {
		m.Reply("falha ao obter saldo!")
	}
}

/*Obter dados, caso existir*/
func getData(user string) (*Dados, bool) {

	fmt.Printf("Procurando %s \n", user)
	i := 0
	for i < db.count {
		if db.dados[i].charid == user {
			return &db.dados[i], false
		}
		i++
	}
	fmt.Printf("Sem sucesso! \n")
	if db.count == 0 {
		db.dados = make([]Dados, 999)
	}
	db.dados[db.count] = Dados{"", "", 0.0, 0.0, 0.0, 0.0, 0.0, false}
	return &db.dados[db.count], true
}

//adiciona usuário ou configura novos dados - /setup
func setuser(m *tbot.Message) {

	//obtem dados do comando /setup
	saldo := getbalance(m.Vars["carteira"])
	tax := strtofloat(m.Vars["taxa"])
	porcent := strtofloat(m.Vars["porc"])

	//verifica se os dados foram recebidos do BE
	if len(saldo) < 1 || strings.HasSuffix(saldo, "error") || strtofloat(saldo) < 5000 {
		m.Reply("Setup: sem sucesso, acho que tem um inseto por aqui NEO!")
		m.Reply("Output: " + saldo)
		return
	} else {
		//se nao foi cadastrado, cadastra-lo
		if db.count < MAX_DB {
			//obter dados ou struct
			user, nexiste := getData(m.From.UserName)

			if nexiste {
				m.Reply("Usuário novo? Bem vindo! Aguarde...")
				time.Sleep(2000 * time.Millisecond)
				m.Reply("$ Hackeando a NAS4!...")
				time.Sleep(2000 * time.Millisecond)
				m.Reply("hehe.. ")
			}

			//novo user
			*user = Dados{
				charid:      m.From.UserName,    //nome do usuário
				charwallet:  m.Vars["carteira"], //carteira do usuário
				balance:     strtofloat(saldo),  //saldo
				taxa:        tax,                //taxa
				porcentagem: porcent,            //porcentagem
				last_rent:   0.0,                //ultima renda verificada
				rendimento:  0.0,                //ultima renda verificada
				isSet:       true,               //está configurado!
			}

			db.count++
			m.Replyf("Hey [%s], o setup foi um sucesso e a NASA hackeada! Tente usar /me para iniciar o foguete.", m.From.FirstName)
			return
		} else {
			m.Reply("Setup: sem sucesso :sob: ! - Usuários máximos cadastrados, tente mais tarde! :clock1:")
		}
	}
	return
}

//comando /me - não requer argumentos, é preciso configra-lo
//*Get masternode rent
func mnrend(m *tbot.Message) {

	//verificar se o usuário fez /setup
	user, inativo := getData(m.From.UserName)
	if inativo {
		m.Reply("Essse comando é automático.\nPortanto configure usando: \n/setup <carteira> <taxa> <porcentagem> !")
		return
	}
	user.balance = strtofloat(getbalance(user.charwallet))
	if user.balance >= 0 {
		fmt.Printf("[%s] Carteira: %s %f\n", m.From.UserName, user.charwallet, user.balance)

		user.last_rent = user.rendimento

		//sem taxa e sem porcentagem
		if user.taxa == 0 && user.porcentagem == 0 {
			user.rendimento = user.balance - MN_BALANCE

			//sem taxa e com porcentagem
		} else if user.taxa == 0 {
			user.rendimento = user.balance - MN_BALANCE
			user.rendimento = (user.rendimento * user.porcentagem) / 100

			//com taxa e com porcentagem
		} else {
			user.rendimento = user.balance - MN_BALANCE
			user.rendimento = (user.rendimento * user.porcentagem) / 100
			user.rendimento = user.rendimento - (user.rendimento * (user.taxa / 100))
		}

	} else {
		m.Reply("falha ao obter saldo!")
		return
	}

	m.Reply("Próximo recompensa: " + floattostr(user.rendimento) + "\nSem taxa: " + floattostr(user.rendimento+(user.rendimento*(user.taxa/100))))

}

func main() {
	bot, err := tbot.NewServer(os.Getenv("TELEGRAM_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	db.count = 0
  
  //Telegram command handlers
	bot.HandleFunc("/sobre", sobre)
	bot.HandleFunc("/start", sobre)
	bot.HandleFunc("/wallet {carteira}", wallet)
	bot.HandleFunc("/setup {carteira} {taxa} {porc}", setuser)
	bot.HandleFunc("/me", mnrend)
	bot.ListenAndServe()

}

//about
func sobre(m *tbot.Message) {
	m.Replyf("------------------------------")
	m.Replyf("----  Comandos básicos   -----")
	m.Replyf("------------------------------")
	m.Replyf("/wallet <carteira> - Obter saldo da carteira.")
	m.Replyf("/eraseme - Limpar seus dados do bot(nome de usuário, carteira, saldo). P.S: Inativo\n")
	m.Replyf("----------------------------")
	m.Replyf("----    Configuração   -----")
	m.Replyf("----------------------------")
	m.Replyf("- 1º use:")
	m.Replyf("/setup <carteira> <taxa> <porcentagem>\n Use para configurar dados. \n- taxa = Taxa do Masternode \n- porcentagem = sua porcentagem do Masternode\n")
	m.Replyf("- 2º agora você só precisa usar:")
	m.Replyf("/me - use quando quiser, após configurar com /setup\n")
	m.Replyf("\n\nP.S: Seu nome de usuário é salvo apenas para poder identificar suas configurações.")
	m.Reply("MN BOT - 0.0.1 - @OneFreakDay - https://github.com/apollomatheus/mxtmnbot")
}
