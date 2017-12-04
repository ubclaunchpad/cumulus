package app

import (
	"fmt"
	"log"
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

	// Set commands
	shell.AddCmd(&ishell.Cmd{
		Name: "send",
		Help: "send coins to another wallet",
		Func: func(ctx *ishell.Context) {
			send(ctx, a)
		},
	})
	shell.AddCmd(&ishell.Cmd{
		Name: "wallet",
		Help: "view the status of a wallet",
		Func: func(ctx *ishell.Context) {
			checkWallet(ctx, a)
		},
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
	shell.AddCmd(&ishell.Cmd{
		Name: "cryptowallet",
		Help: "enable or disable password protection for private key storage",
		Func: func(ctx *ishell.Context) {
			cryptoWallet(ctx, a)
		},
	})

	// Set interrupt handler
	shell.Interrupt(func(ctx *ishell.Context, count int, input string) {
		ctx.Println("Saving app state and flushing logs...")
		a.onExit()
	})
	shell.Start()
	emoji.Println(":cloud: Welcome to the :sunny: Cumulus console :cloud:")
	return shell
}

func encrypt(ctx *ishell.Context, app *App, password string) error {
	err := app.CurrentUser.EncryptPrivateKey(password)
	if err != nil {
		log.Printf("Unable to encrypt private key: %s", err.Error())
		return err
	}

	if err := app.CurrentUser.Save(userFileName); err != nil {
		app.CurrentUser.DecryptPrivateKey(password)
		ctx.Print(err)
		return err
	}

	return nil
}

func decrypt(ctx *ishell.Context, app *App, password string) error {
	err := app.CurrentUser.DecryptPrivateKey(password)
	if err != nil {
		//log.Printf("Unable to decrypt private key: %s", err.Error())
		return err
	}

	if err := app.CurrentUser.Save(userFileName); err != nil {
		app.CurrentUser.EncryptPrivateKey(password)
		ctx.Print(err)
		return err
	}

	return nil
}

func cryptoWallet(ctx *ishell.Context, app *App) {
	if len(ctx.Args) < 1 {
		ctx.Println("Usage: cryptowallet [enable/disable]")
		return
	}

	switch ctx.Args[0] {
	case "enable":
		if app.CurrentUser.CryptoWallet {
			ctx.Print("CryptoWallet is already enabled")
		} else {
			ctx.Print("Please enter password: ")
			password := ctx.ReadPassword()
			err := encrypt(ctx, app, password)
			if err != nil {
				ctx.Print("Unable to decrypt private key")
			} else {
				ctx.Print("Successfully enabled cryptowallet")
			}
		}
	case "disable":
		if !app.CurrentUser.CryptoWallet {
			ctx.Print("CryptoWallet is already disabled")
		} else {
			ctx.Print("Please enter password: ")
			password := ctx.ReadPassword()
			err := decrypt(ctx, app, password)
			if InvalidPassword(err) {
				ctx.Print("Inavalid password, please try again: ")
				password := ctx.ReadPassword()
				err := decrypt(ctx, app, password)
				if err != nil {
					ctx.Println("Unable to decrypt private key")
				}
			} else {
				ctx.Print("Successfully disabled cryptowallet")
			}

		}
	case "status":
		var s string
		if app.CurrentUser.CryptoWallet {
			s = "enabled"
		} else {
			s = "disabled"
		}
		ctx.Printf("cryptowallet status: %s", s)
	default:
		//Fall through
	}
}

func send(ctx *ishell.Context, app *App) {
	if len(ctx.Args) < 2 {
		ctx.Println("Usage: send [amount] [public address]")
		return
	}

	amount, err := strconv.ParseFloat(ctx.Args[0], 64)
	if err != nil {
		ctx.Println(err)
		return
	} else if amount <= 0 {
		ctx.Println("Amount must be a positive decimal value")
		return
	}
	amount *= float64(blockchain.CoinValue)
	addr := ctx.Args[1]

	password := ""
	if app.CurrentUser.CryptoWallet {
		ctx.Println("Please enter password to decrypt private key: ")
		password = ctx.ReadPassword()
		err := decrypt(ctx, app, password)
		if err != nil {
			ctx.Println("Unable to decrypt private key")
			return
		}
	}
	// Try to make a payment.
	ctx.Println("Sending amount", coinValue(uint64(amount)), "to", addr)
	err = app.Pay(addr, uint64(amount))

	if app.CurrentUser.CryptoWallet {
		err := encrypt(ctx, app, password)
		if err != nil {
			ctx.Println("Unable to re-encrypt private key, CryptoWallet disabled")
		}
	}

	if err != nil {
		emoji.Println(":disappointed: ", err)
	} else {
		emoji.Println(":mailbox_with_mail: Its in the mail!")
	}
}

func checkWallet(ctx *ishell.Context, app *App) {
	app.Chain.RLock()
	defer app.Chain.RUnlock()

	wallet := app.CurrentUser.Wallet

	// Show actual and effective balance
	ctx.Println("Balance:", coinValue(wallet.Balance))
	ctx.Println("Effective Balance:", coinValue(wallet.GetEffectiveBalance()))

	// Show list of pending transactions
	if len(wallet.PendingTxns) > 0 {
		ctx.Println("Pending Transactions:")
		for i, txn := range wallet.PendingTxns {
			ctx.Println("\nTransaction ", strconv.Itoa(i))

			var recipient string
			for _, output := range txn.Outputs {
				if output.Recipient != txn.Sender.Repr() {
					recipient = output.Recipient
					break
				}
			}

			ctx.Println("\tAmount:", coinValue(txn.GetTotalOutputFor(recipient)))
			ctx.Println("\tRecipient:", recipient)
		}
	} else {
		ctx.Println("No pending transactions")
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
	usage := func(ctx *ishell.Context) {
		ctx.Println("\nUsage: miner [command]")
		ctx.Println("\nCOMMANDS:")
		ctx.Println("\t start \t Start the miner")
		ctx.Println("\t stop \t Stop the miner")
	}

	if len(ctx.Args) != 1 {
		if app.Miner.State() == miner.Running {
			shell.Println("Miner is running")
		} else if app.Miner.State() == miner.Paused {
			shell.Println("Miner is paused")
		} else {
			shell.Println("Miner is stopped")
		}
		usage(ctx)
		return
	}

	switch ctx.Args[0] {
	case "start":
		if app.Miner.State() == miner.Running {
			shell.Println("Miner is already running")
		} else if app.Miner.State() == miner.Paused {
			app.Miner.ResumeMining()
			shell.Println("Resumed mining")
		} else {
			go app.RunMiner()
			shell.Println("Started miner")
		}
	case "stop":
		if app.Miner.State() == miner.Stopped {
			shell.Println("Miner is already stopped")
			return
		}
		app.Miner.StopMining()
		shell.Println("Stopped miner")
	case "pause":
		wasRunning := app.Miner.PauseIfRunning()
		if wasRunning {
			shell.Println("Paused miner")
		} else {
			shell.Println("Miner was not running")
		}
	default:
		usage(ctx)
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

func editUser(ctx *ishell.Context, app *App) {
	if len(ctx.Args) == 0 {
		ctx.Println("Current User:")
		ctx.Println("Name:", app.CurrentUser.Name)
		ctx.Println("Blocksize:", app.CurrentUser.BlockSize)
		ctx.Println("Address:", app.CurrentUser.Public().Repr())
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
		"between 1000 and 5000000 btyes)")
}

func coinValue(amount uint64) string {
	return fmt.Sprintf("%d (%f cumuli)", amount, float64(amount)/float64(blockchain.CoinValue))
}
