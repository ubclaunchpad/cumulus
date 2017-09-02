package app

import (
	"strconv"

	"github.com/abiosoft/ishell"
	"github.com/ubclaunchpad/cumulus/blockchain"
	"github.com/ubclaunchpad/cumulus/miner"
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
	shell.AddCmd(&ishell.Cmd{
		Name: "miner",
		Help: "view or toggle miner status",
		Func: func(ctx *ishell.Context) {
			toggleMiner(ctx, a)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "user",
		Help: "view or edit current user's info",
		Func: func(ctx *ishell.Context) {
			editUser(ctx, a)
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

func toggleMiner(ctx *ishell.Context, app *App) {
	if len(ctx.Args) != 1 {
		if app.Miner.State() == miner.Running {
			shell.Println("Miner is running.")
		} else if app.Miner.State() == miner.Paused {
			shell.Println("Miner is paused.")
		} else {
			shell.Println("Miner is stopped.")
		}
		shell.Println("Use 'miner start' or 'miner stop' to start or stop the miner.")
		return
	}

	switch ctx.Args[0] {
	case "start":
		if app.Miner.State() == miner.Running {
			shell.Println("Miner is already running.")
		} else if app.Miner.State() == miner.Paused {
			app.Miner.ResumeMining()
			shell.Println("Resumed mining.")
		} else {
			go app.RunMiner()
			shell.Println("Started miner.")
		}
	case "stop":
		if app.Miner.State() == miner.Stopped {
			shell.Println("Miner is already stopped.")
			return
		}
		app.Miner.StopMining()
		shell.Println("Stopped miner.")
	case "pause":
		wasRunning := app.Miner.PauseIfRunning()
		if wasRunning {
			shell.Println("Paused miner.")
		} else {
			shell.Println("Miner was not running.")
		}
	default:
		shell.Println("Usage: miner [start] | [stop]")
	}
}

func createWallet(ctx *ishell.Context, app *App) {
	// Create a new wallet and set as CurrentUser's wallet.
	wallet := blockchain.NewWallet()
	app.CurrentUser.Wallet = wallet
	emoji.Println(":credit_card: New wallet created!")

	// Give a printout of the address(es).
	emoji.Print(":mailbox:")
	ctx.Println(" Address: " + wallet.Public().Repr())
	emoji.Println(":fist: Emoji Address: " + wallet.Public().Emoji())
	ctx.Println("")
}

func createTransaction(ctx *ishell.Context, app *App) {
	// Read in the recipient address.
	emoji.Print(":credit_card:")
	ctx.Println(" Enter recipient wallet address")
	toAddress := shell.ReadLine()

	// Get amount from user.
	emoji.Print(":dollar:")
	ctx.Println(" Enter amount to send: ")
	amount, err := strconv.ParseUint(shell.ReadLine(), 10, 64)
	if err != nil {
		emoji.Println(":disappointed: Invalid number format. Please enter an amount in decimal format.")
		return
	}

	// Try to make a payment.
	err = app.Pay(toAddress, amount)
	if err != nil {
		emoji.Println(":disappointed: Transaction failed!")
		ctx.Println(err.Error())
	} else {
		emoji.Println(":mailbox_with_mail: Its in the mail!")
	}
}

func editUser(ctx *ishell.Context, app *App) {
	if len(ctx.Args) == 0 {
		ctx.Println("Current User:")
		ctx.Println("Name:", app.CurrentUser.Name)
		ctx.Println("Blocksize:", app.CurrentUser.BlockSize)
		ctx.Println("Address:", app.CurrentUser.Public())
		emoji.Println("Emoji Address:", app.CurrentUser.Public().Emoji())
	} else if len(ctx.Args) == 2 {
		if ctx.Args[0] == "name" {
			app.CurrentUser.Name = ctx.Args[1]
			if err := app.CurrentUser.Save(userFileName); err != nil {
				ctx.Print(err)
			}
			return
		} else if ctx.Args[0] == "blocksize" {
			size, err := strconv.ParseUint(ctx.Args[1], 10, 32)
			if err != nil {
				ctx.Println(err)
			} else if size < MinBlockSize || size > MaxBlockSize {
				ctx.Println("Block size must be between", MinBlockSize, "and",
					MaxBlockSize, "btyes")
				return
			}
			app.CurrentUser.BlockSize = (uint32)(size)
			if err := app.CurrentUser.Save(userFileName); err != nil {
				ctx.Print(err)
			}
			return
		}
	}

	ctx.Println("\nUsage: user [command] [value]")
	ctx.Println("\nCOMMANDS:")
	ctx.Println("\t name      \t Set the current user's name")
	ctx.Println("\t blocksize \t Set the current user's blocksize (must be " +
		"between 1000 and 5000000 btyes")
}
