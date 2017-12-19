/* Proctected by MIT License: https://opensource.org/licenses/MIT
   Author Telegram: @OneFreakDay
   Date: 16/12/2017
   Description:
   A simple telegram bot to connect with a Block Explorer API cryptocoin
*/
package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	//extern dependency
	"github.com/yanzay/tbot"
)

var db Database

/* Obter dados do usuário */
func getData(user string) (*Dados, bool) {
	fmt.Printf("$Procurando %s \n", user)

	for i := 0; i < db.count; i++ {
		if db.dados[i].charid == user {
			return &db.dados[i], false
		}
		i++
	}

	fmt.Printf("$Sem sucesso! \n")

	//alocar
	if db.count == 0 {
		db.dados = make([]Dados, 999)
	}

	//novo
	db.dados[db.count] = Dados{"", "", 0.0, 0.0, 0.0, 0.0, 0.0, false, nil, 0, nil, 0}
	return &db.dados[db.count], true
}

/*Obter anuncios de compra do usuário */
func getDataBuy(user string) (*Buy, bool) {

	fmt.Printf("$Procurando Buys de %s \n", user)

	for i := 0; i < db.count; i++ {
		if db.dados[i].charid == user {

			//buscar um local desativado
			for x := 0; x < db.dados[i].buyc; x++ {
				if db.dados[i].buys[x].end {
					db.dados[i].buys[x].id = x
					return &db.dados[i].buys[x], false
				}
			}
			//novo
			db.dados[i].buyc++
			return &db.dados[i].buys[db.dados[i].buyc-1], true
		}
		i++
	}
	fmt.Printf("$Sem sucesso! \n")
	return nil, true
}

/*Obter anuncios de venda do usuário */
func getDataSell(user string) (*Sell, bool) {

	fmt.Printf("$Procurando Sells de %s \n", user)

	for i := 0; i < db.count; i++ {
		if db.dados[i].charid == user {
			//algum local desativado?
			for x := 0; x < db.dados[i].sellc; x++ {
				if db.dados[i].sells[x].end {
					return &db.dados[i].sells[x], false
				}
			}
			//novo
			db.dados[i].sellc++
			return &db.dados[i].sells[db.dados[i].sellc-1], true
		}
		i++
	}
	fmt.Printf("$Sem sucesso! \n")
	return nil, true
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
func mnrend(m *tbot.Message) {

	//verificar se o usuário fez /setup
	user, inativo := getData(m.From.UserName)
	if inativo || !user.isSet {
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

	bot.HandleFunc("/sobre", sobre)
	bot.HandleFunc("/start", sobre)

	bot.HandleFunc("/wallet {carteira}", wallet)

	//rendimento do masternode
	bot.HandleFunc("/setup {carteira} {taxa} {porc}", setuser)
	bot.HandleFunc("/me", mnrend)

	//anunciar/finalizar/listar venda
	bot.HandleFunc("/newsell {XNS} {MXT} {BRL} {contato}", newsell)
	bot.HandleFunc("/endsell {id}", endsell)
	bot.HandleFunc("/lsells", listsells)

	//anunciar/finalizar/listar compra
	bot.HandleFunc("/newbuy {MXT} {contato}", newbuy)
	bot.HandleFunc("/endbuy {id}", endbuy)
	bot.HandleFunc("/lbuys", listbuys)
	bot.ListenAndServe()

}

func newsell(m *tbot.Message) {
	user, nexiste := getData(m.From.UserName)
	if nexiste {
		m.Reply("Usuário cadastrado! Bem vindo!")
		m.Reply("Tente não acomular sells!")

		*user = Dados{
			charid: m.From.UserName, //nome do usuário
			sells:  make([]Sell, MAX_DB_SELL),
			buys:   make([]Buy, MAX_DB_BUY),
		}

		db.count++
	}
	if user.sellc >= MAX_DB_SELL {
		m.Replyf("Você excedeu %d pedidos.", MAX_DB_SELL)
		return
	}

	//dados
	xnsn := m.Vars["XNS"]
	mxtf := m.Vars["MXT"]
	brlf := m.Vars["BRL"]
	contn := m.Vars["contato"]

	//log
	fmt.Printf("#BUY {%s} {%s} {%s}\n", m.From.UserName, mxtf, brlf, contn)

	//fail?
	if len(xnsn) < 0 || len(mxtf) < 0 || len(brlf) < 0 || len(contn) < 0 {
		m.Reply("Use: /annsell <xns_n> <mxt_pedido> <brl_pedido> <contato> ")
		return
	} else {
		//cadastro
		bi, novo := getDataSell(m.From.UserName)

		*bi = Sell{
			xns:       xnsn,
			contato:   contn,
			mxt_price: strtofloat(mxtf),
			brl_price: strtofloat(brlf),
			end:       false,
		}

		//output
		m.Reply("--------------------")
		m.Reply("▌▌▌▌EXTRATO▌▌▌▌")
		if novo {
			m.Replyf("▌ID: %d", user.sellc)
		} else {
			m.Replyf("▌ID: %d", bi.id)
		}
		m.Reply("▌de: " + bi.contato)
		m.Reply("▌MN: XNs" + bi.xns)
		m.Replyf("▌mxt_price: %f", bi.mxt_price)
		m.Replyf("▌brl_price: %f", bi.brl_price)
		m.Reply("--------------------")
		m.Reply("Nao apague o extrato!")
		m.Reply("--------------------")
	}

}

func endsell(m *tbot.Message) {
	user, nexiste := getData(m.From.UserName)
	if nexiste {
		m.Reply("Você não está cadastrado!")
		return
	}
	if user.sellc <= 0 {
		m.Reply("Você não efetuou nenhum anúncio de venda.")
		return
	}

	//dados
	bid, err := strconv.Atoi(m.Vars["id"])

	//fail?
	if err != nil {
		m.Reply("Use: /endsell <id>")
		return
	} else {
		//cadastro
		user.sells[bid].end = true

		//output
		m.Reply("Feito!")
	}
}

func newbuy(m *tbot.Message) {
	user, inexiste := getData(m.From.UserName)
	if inexiste {
		m.Reply("Usuário cadastrado! Bem vindo!")
		m.Reply("Tente não acomular buys!")

		*user = Dados{
			charid: m.From.UserName, //nome do usuário
			sells:  make([]Sell, MAX_DB_SELL),
			buys:   make([]Buy, MAX_DB_BUY),
		}

		db.count++
	}
	if user.buyc >= MAX_DB_BUY {
		m.Replyf("Você excedeu %d pedidos.", MAX_DB_BUY)
		return
	}

	//dados
	mxtf := m.Vars["MXT"]
	contn := m.Vars["contato"]

	//log
	fmt.Printf("#SELL {%s} {%s}\n", m.From.UserName, mxtf, contn)

	//invalido
	if len(mxtf) < 0 || len(contn) < 0 {
		m.Reply("Use: /annbuy <mxt_pedido> <contato> ")
		return

	} else {
		//cadastro
		bi, novo := getDataBuy(m.From.UserName)
		*bi = Buy{
			contato:   contn,
			mxt_price: strtofloat(mxtf),
			end:       false,
		}

		//output
		m.Reply("--------------------")
		m.Reply("▌▌▌▌EXTRATO▌▌▌▌")
		if novo {
			m.Replyf("▌ID: %d", user.buyc)
		} else {
			m.Replyf("▌ID: %d", bi.id)
		}
		m.Reply("▌de: " + bi.contato)
		m.Replyf("▌mxt_price: %f", bi.mxt_price)
		m.Reply("--------------------")
		m.Reply("Nao apague o extrato!")
		m.Reply("--------------------")
	}

}

func endbuy(m *tbot.Message) {
	user, nexiste := getData(m.From.UserName)
	if nexiste {
		m.Reply("Você não está cadastrado!")
		return
	}
	if user.buyc <= 0 {
		m.Reply("Você não efetuou nenhum anuncio de compra.")
		return
	}

	//dados
	bid, err := strconv.Atoi(m.Vars["id"])

	//fail?
	if err != nil || bid < 0 || bid > MAX_DB_BUY {
		m.Reply("Use: /endbuy <id>")
		return
	} else {

		//cadastro
		user.buys[bid].end = true

		//output
		m.Reply("Feito!")
	}
}

//listar ordens de venda
func listsells(m *tbot.Message) {
	m.Reply("ID - XNs - MXT pedido - BRL pedido - Contato ")
	for i := 0; i < db.count; i++ {
		fmt.Print(db.dados[i].sellc)
		for x := 0; x < db.dados[i].sellc; x++ {
			if !db.dados[i].sells[x].end {
				m.Replyf("%d - %s - %f - %f - %s", x,
					db.dados[i].sells[x].xns,
					db.dados[i].sells[x].mxt_price,
					db.dados[i].sells[x].brl_price,
					db.dados[i].sells[x].contato)
			}
		}
	}
}

//listar ordens de compra
func listbuys(m *tbot.Message) {
	m.Reply("ID - lance de MXT - Contato ")
	for i := 0; i < db.count; i++ {
		for x := 0; x < db.dados[i].buyc; x++ {
			if !db.dados[i].buys[x].end {
				m.Replyf("%d - %f - %s", x,
					db.dados[i].buys[x].mxt_price,
					db.dados[i].buys[x].contato)
			}
		}
	}
}

//about
func sobre(m *tbot.Message) {
	m.Reply("MN BOT - 0.0.2 - @OneFreakDay")
	m.Replyf("------------------------------")
	m.Replyf("----  Comandos básicos   -----")
	m.Replyf("------------------------------")
	m.Replyf("/wallet <carteira> - Obter saldo de uma carteira.")
	m.Replyf("------------------------------------\n")
	m.Replyf("/newsell <xns_n> <mxt_pedido> <brl_pedido> <contato> - Anunciar venda de XNs\n")
	m.Replyf("/endsell <id> - Finalizar anuncio de venda de XNs\n")
	m.Replyf("/mysells - Listar suas vendas XNs\n")
	m.Replyf("------------------------------------\n")
	m.Replyf("/newbuy <mxt_max> <contato> - Anunciar compra de XNs\n")
	m.Replyf("/endbuy <id> - Finalizar anuncio de compra de XNs\n")
	m.Replyf("/mybuys - Listar suas compras XNs\n")
	m.Replyf("------------------------------------\n")
	m.Replyf("/lbuys - Listar ofertas de compra\n")
	m.Replyf("/lsells - Listar ofertas de venda\n")

	m.Replyf("\n----------------------------------------------------------")
	m.Replyf("----    Configuração para rendimento de masternode   -----")
	m.Replyf("----------------------------------------------------------")
	m.Replyf("- 1º use:")
	m.Replyf("/setup <carteira> <taxa> <porcentagem>\n Use para configurar dados do Masternode. \n- taxa = Taxa do Masternode \n- porcentagem = sua porcentagem do Masternode\n")
	m.Replyf("- 2º agora você só precisa usar:")
	m.Replyf("/me - Sua Próxima recompensa\n")
	m.Replyf("*Seu nome de usuário é salvo apenas para poder identificar suas configurações.")
	m.Replyf("*Dados são apagados quando o bot fica offline.")
	m.Reply("Código: https://github.com/apollomatheus/mxtmnbot")
}
