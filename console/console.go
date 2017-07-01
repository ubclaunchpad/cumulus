package console

import (
	"strconv"

	"github.com/abiosoft/ishell"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/conn"
	"github.com/ubclaunchpad/cumulus/peer"
)

var shell *ishell.Shell

// Run starts the Cumulus console. This should be run only once as a
// goroutine, and logging should be redirected away from stdout before it is run.
func Run() {
	shell = ishell.New()
	shell.AddCmd(&ishell.Cmd{
		Name: "create",
		Help: "create a new wallet hash or transaction",
		Func: create,
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "check",
		Help: "check the status of a transaction or wallet",
		Func: check,
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "listen-address",
		Help: "show the address this host is listening on",
		Func: listenAddr,
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "peers",
		Help: "show the peers this host is connected to",
		Func: peers,
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "connect",
		Help: "connect to another peer",
		Func: connect,
	})

	shell.Start()
	shell.Println("Cumulus Console")
}

func create(ctx *ishell.Context) {
	choice := ctx.MultiChoice([]string{
		"Wallet",
		"Transaction",
	}, "What would you like to create?")
	if choice == 0 {
		ctx.Println("New Wallet:", blockchain.NewWallet().Public())
	} else {
		shell.Print("Sender wallet ID: ")
		senderID := shell.ReadLine()
		shell.Print("Recipient wallet ID: ")
		recipientID := shell.ReadLine()
		shell.Print("Amount to send: ")
		amount, err := strconv.ParseFloat(shell.ReadLine(), 64)
		if err != nil {
			shell.Println("Invalid number format. Please enter an amount in decimal format.")
			return
		}

		// TODO: make transaction, add it to the pool, broadcast it
		ctx.Printf(`\nNew Transaction: \nSenderID: %s \nRecipiendID: %s\nAmount: %f"`,
			senderID, recipientID, amount)
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

func listenAddr(ctx *ishell.Context) {
	shell.Println("Listening on", peer.ListenAddr)
}

func peers(tcx *ishell.Context) {
	shell.Println("Connected to", peer.PStore.Addrs())
}

func connect(ctx *ishell.Context) {
	if len(ctx.Args) == 0 {
		shell.Println("Usage: connect [IP address]:[TCP port]")
		return
	}

	addr := ctx.Args[0]
	c, err := conn.Dial(addr)
	if err != nil {
		shell.Println("Failed to dial peer", addr, ":", err)
		return
	}
	peer.ConnectionHandler(c)
	if peer.PStore.Get(addr) == nil {
		shell.Println("Failed to extablish connection. See logs for details.")
	} else {
		shell.Println("Connected to", addr)
	}
}
