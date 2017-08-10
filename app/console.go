package app

import (
	"strconv"

	"github.com/abiosoft/ishell"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/peer"
	"gopkg.in/kyokomi/emoji.v1"
)

var (
	shell *ishell.Shell
)

// RunConsole starts the Cumulus console. This should be run only once as a
// goroutine, and logging should be redirected away from stdout before it is run.
// It takes a pointer to a PeerStore so we can use the PeerStore to interact
// with other peers and give the user info about the running instance.
func RunConsole(a *App) *ishell.Shell {
	shell = ishell.New()

	shell.AddCmd(&ishell.Cmd{
		Name: "create",
		Help: "create a new wallet hash or transaction",
		Func: func(ctx *ishell.Context) {
			create(ctx, a)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "check",
		Help: "check the status of a transaction or wallet",
		Func: check,
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "address",
		Help: "show the address this host is listening on",
		Func: func(ctx *ishell.Context) {
			listenAddr(ctx, a)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "peers",
		Help: "show the peers this host is connected to",
		Func: func(ctx *ishell.Context) {
			peers(ctx, a)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "connect",
		Help: "connect to another peer",
		Func: func(ctx *ishell.Context) {
			connect(ctx, a)
		},
	})

	shell.Start()
	emoji.Println(":cloud: Welcome to the :sunny: Cumulus console :cloud:")
	return shell
}

func create(ctx *ishell.Context, app *App) {
	choice := ctx.MultiChoice([]string{
		"Wallet",
		"Transaction",
	}, "What would you like to create?")
	if choice == 0 {
		createWallet(ctx, app)
	} else {
		createTransaction(ctx, app)
	}
}

func check(ctx *ishell.Context) {
	choice := ctx.MultiChoice([]string{
		"Wallet",
		"Transaction",
	}, "What would you like to check the status of?")
	if choice == 0 {
		ctx.Println("Wallet status: ")
	} else {
		ctx.Println("Transaction status: ")
	}
}

func listenAddr(ctx *ishell.Context, a *App) {
	shell.Println("Listening on", a.PeerStore.ListenAddr)
}

func peers(tcx *ishell.Context, a *App) {
	shell.Println("Connected to", a.PeerStore.Addrs())
}

func connect(ctx *ishell.Context, a *App) {
	if len(ctx.Args) == 0 {
		shell.Println("Usage: connect [IP address]:[TCP port]")
		return
	}

	addr := ctx.Args[0]
	_, err := peer.Connect(addr, a.PeerStore)
	if err != nil {
		shell.Println("Failed to extablish connection:", err)
	} else {
		shell.Println("Connected to", addr)
	}
}

func createWallet(ctx *ishell.Context, app *App) {
	wallet := blockchain.NewWallet()
	app.CurrentUser.Wallet = wallet
	emoji.Println(":credit_card: New wallet created!")
	emoji.Print(":mailbox:")
	ctx.Println(" Address: " + wallet.Public().Repr())
	emoji.Println(":fist: Emoji Address: " + wallet.Public().Emoji())
	ctx.Println("")
}

func createTransaction(ctx *ishell.Context, app *App) {
	emoji.Print(":credit_card:")
	ctx.Println(" Enter recipient wallet address")
	toAddress := shell.ReadLine()
	// TODO: Error handling address input.

	emoji.Print(":dollar:")
	ctx.Println(" Enter amount to send: ")
	amount, err := strconv.ParseUint(shell.ReadLine(), 10, 64)
	if err != nil {
		emoji.Println(":disappointed: Invalid number format. Please enter an amount in decimal format.")
		return
	}

	// TODO: Check if we have enough coins to make the purchase.
	err = app.Pay(toAddress, amount)
	if err != nil {
		emoji.Println(":disappointed: Transaction failed!")
		ctx.Println(err.Error)
	} else {
		emoji.Println(":mailbox_with_mail: Its in the mail!")
	}
}
