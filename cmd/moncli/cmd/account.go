package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/glynternet/go-accounting/account"
	"github.com/glynternet/go-accounting/balance"
	"github.com/glynternet/go-money/currency"
	"github.com/glynternet/mon/pkg/date"
	"github.com/glynternet/mon/pkg/filter"
	"github.com/glynternet/mon/pkg/storage"
	"github.com/glynternet/mon/pkg/table"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	keyDate               = "date"
	keyAmount             = "amount"
	keyNote               = "note"
	keyName               = "name"
	keyCurrency           = "currency"
	keyOpened             = "opened"
	keyClosed             = "closed"
	keyLimit              = "limit"
	keyOpeningBalance     = "opening-balance"
	keyOpeningBalanceNote = "opening-balance-note"
	keyClosingBalance     = "closing-balance"
	keyClosingBalanceNote = "closing-balance-note"
)

var (
	accountOpened = date.Flag()
	accountClosed = date.Flag()
)

var accountCmd = &cobra.Command{
	Use:   "account [ID]",
	Short: "retrieve account info",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return errors.Wrap(err, "parsing account id")
		}

		a, err := newClient().SelectAccount(uint(id))
		if err != nil {
			return errors.Wrap(err, "selecting account")
		}

		table.Accounts(storage.Accounts{*a}, os.Stdout)
		return nil
	},
}

var accountAddCmd = &cobra.Command{
	Use:   "add [NAME]",
	Short: "add an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cc, err := currency.NewCode(viper.GetString(keyCurrency))
		if err != nil {
			return errors.Wrap(err, "creating new currency code")
		}

		opened := time.Now()
		if accountOpened.Time != nil {
			opened = *accountOpened.Time
		}

		var ops []account.Option
		if accountClosed.Time != nil {
			ops = append(ops, account.CloseTime(*accountClosed.Time))
		}

		a, err := account.New(
			args[0],
			*cc,
			opened,
			ops...,
		)
		if err != nil {
			return errors.Wrap(err, "creating new account for insert")
		}

		i, err := newClient().InsertAccount(*a)
		if err != nil {
			return errors.Wrap(err, "inserting new account")
		}
		table.Accounts(storage.Accounts{*i}, os.Stdout)
		return nil
	},
}

var accountOpenCmd = &cobra.Command{
	Use:   "open [NAME]",
	Short: "open an account with a balance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cc, err := currency.NewCode(viper.GetString(keyCurrency))
		if err != nil {
			return errors.Wrap(err, "creating new currency code")
		}

		opened := time.Now()
		if accountOpened.Time != nil {
			opened = *accountOpened.Time
		}

		a, err := account.New(args[0], *cc, opened)
		if err != nil {
			return errors.Wrap(err, "creating new account for insert")
		}

		c := newClient()

		i, err := c.InsertAccount(*a)
		if err != nil {
			return errors.Wrap(err, "inserting new account")
		}

		b, err := c.InsertBalance(
			(*i).ID,
			balance.Balance{
				Date:   i.Account.Opened(),
				Amount: viper.GetInt(keyOpeningBalance),
			},
			viper.GetString(keyOpeningBalanceNote),
		)
		if err != nil {
			return errors.Wrap(err, "inserting balance")
		}

		table.Accounts(storage.Accounts{*i}, os.Stdout)
		table.Balances(storage.Balances{*b}, os.Stdout)
		return nil
	},
}

var accountReopenCmd = &cobra.Command{
	Use:   "reopen [ID]",
	Short: "reopen an account",
	Long:  "reopen removes an account's closed date",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return errors.Wrap(err, "parsing account id")
		}

		c := newClient()

		a, err := c.SelectAccount(uint(id))
		if err != nil {
			return errors.Wrap(err, "selecting account")
		}

		aux := a.Account
		us, err := account.New(aux.Name(), aux.CurrencyCode(), aux.Opened())
		if err != nil {
			return errors.Wrap(err, "creating updates account from selected account")
		}

		b, err := c.UpdateAccount(a.ID, *us)
		if err != nil {
			return errors.Wrap(err, "applying updates")
		}

		fmt.Println("Reopened:")
		table.Accounts(storage.Accounts{*b}, os.Stdout)
		return nil
	},
}

var accountDeleteCmd = &cobra.Command{
	Use:   "delete [ID]",
	Short: "delete an account",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return errors.Wrap(err, "parsing account id")
		}

		c := newClient()

		a, err := c.SelectAccount(uint(id))
		if err != nil {
			return errors.Wrap(err, "selecting account")
		}

		err = c.DeleteAccount(a.ID)
		if err != nil {
			return errors.Wrap(err, "deleting account")
		}

		fmt.Println("Deleted:")
		table.Accounts(storage.Accounts{*a}, os.Stdout)
		return nil
	},
}

var accountCloseCmd = &cobra.Command{
	Use:   "close [ID]",
	Short: "close an account with a balance",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		closed := time.Now()
		if balanceDate.Time != nil {
			closed = *balanceDate.Time
		}

		id, err := parseID(args[0])
		if err != nil {
			return errors.Wrap(err, "parsing account id")
		}

		c := newClient()

		a, err := c.SelectAccount(uint(id))
		if err != nil {
			return errors.Wrap(err, "selecting account")
		}

		b, err := c.InsertBalance(
			(*a).ID,
			balance.Balance{
				Date:   closed,
				Amount: viper.GetInt(keyClosingBalance),
			},
			viper.GetString(keyClosingBalanceNote),
		)
		if err != nil {
			return errors.Wrap(err, "inserting balance")
		}

		us, err := account.New(
			a.Account.Name(),
			a.Account.CurrencyCode(),
			a.Account.Opened(),
			account.CloseTime(b.Date),
		)
		if err != nil {
			return errors.Wrap(err, "creating updates account")
		}

		u, err := c.UpdateAccount(a.ID, *us)
		if err != nil {
			return errors.Wrap(err, "updating account")
		}

		table.Accounts(storage.Accounts{*u}, os.Stdout)
		table.Balances(storage.Balances{*b}, os.Stdout)
		return nil
	},
}

var accountUpdateCmd = &cobra.Command{
	Use:   "update [ID]",
	Short: "update an account",
	Long: `update an account with the given details. 
All of the details of an account must be provided, even if they are exactly 
the same as the original account`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return errors.Wrap(err, "parsing account id")
		}

		c := newClient()
		a, err := c.SelectAccount(uint(id))
		if err != nil {
			return errors.Wrap(err, "selecting account to update")
		}

		opened := time.Now()
		if accountOpened.Time != nil {
			opened = *accountOpened.Time
		}

		var ops []account.Option
		if accountClosed.Time != nil {
			ops = append(ops, account.CloseTime(*accountClosed.Time))
		}

		cc, err := currency.NewCode(viper.GetString(keyCurrency))
		if err != nil {
			return errors.Wrap(err, "creating new currency code")
		}

		us, err := account.New(viper.GetString(keyName), *cc, opened, ops...)
		if err != nil {
			return errors.Wrap(err, "creating account for update")
		}

		u, err := c.UpdateAccount(a.ID, *us)
		if err != nil {
			return errors.Wrap(err, "updating account")
		}

		fmt.Println("ORIGINAL")
		table.Accounts(storage.Accounts{*a}, os.Stdout)

		fmt.Println("UPDATED")
		table.Accounts(storage.Accounts{*u}, os.Stdout)
		return nil
	},
}

var accountRenameCmd = &cobra.Command{
	Use:   "rename [ID] [NEW NAME]",
	Short: "rename an account",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return errors.Wrap(err, "parsing account id")
		}

		c := newClient()
		a, err := c.SelectAccount(uint(id))
		if err != nil {
			return errors.Wrap(err, "selecting account")
		}

		var ops []account.Option
		if a.Account.Closed().Valid {
			ops = append(ops, account.CloseTime(a.Account.Closed().Time))
		}

		us, err := account.New(
			args[1],
			a.Account.CurrencyCode(),
			a.Account.Opened(),
			ops...,
		)
		if err != nil {
			return errors.Wrap(err, "creating new account for update")
		}

		u, err := c.UpdateAccount(a.ID, *us)
		if err != nil {
			return errors.Wrap(err, "updating account")
		}

		fmt.Println("ORIGINAL")
		table.Accounts(storage.Accounts{*a}, os.Stdout)

		fmt.Println("UPDATED")
		table.Accounts(storage.Accounts{*u}, os.Stdout)
		return nil
	},
}

var accountBalancesCmd = &cobra.Command{
	Use:  "balances [ID]",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return errors.Wrap(err, "parsing account id")
		}

		c := newClient()
		a, err := c.SelectAccount(uint(id))
		if err != nil {
			return errors.Wrap(err, "selecting account")
		}

		table.Accounts(storage.Accounts{*a}, os.Stdout)

		bs, err := c.SelectAccountBalances((*a).ID)
		if err != nil {
			return errors.Wrap(err, "selecting account balances")
		}

		limit := viper.GetInt(keyLimit)
		if limit > len(*bs) {
			limit = len(*bs)
		}
		if limit != 0 {
			*bs = (*bs)[len(*bs)-limit:]
		}

		table.Balances(*bs, os.Stdout)
		return nil
	},
}

var balanceDate = date.Flag()
var accountBalanceInsertCmd = &cobra.Command{
	Use:  "balance-insert [ID]",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return errors.Wrap(err, "parsing account id")
		}

		c := newClient()
		a, err := c.SelectAccount(uint(id))
		if err != nil {
			return errors.Wrap(err, "selecting account")
		}

		t := time.Now()
		if balanceDate.Time != nil {
			t = *balanceDate.Time
		}

		b, err := c.InsertBalance(
			(*a).ID,
			balance.Balance{
				Date:   t,
				Amount: viper.GetInt(keyAmount),
			},
			viper.GetString(keyNote),
		)
		if err != nil {
			return errors.Wrap(err, "inserting balance")
		}

		table.Accounts(storage.Accounts{*a}, os.Stdout)
		table.Balances(storage.Balances{*b}, os.Stdout)
		return nil
	},
}

var accountBalanceCmd = &cobra.Command{
	Use:  "balance [ID]",
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := parseID(args[0])
		if err != nil {
			return errors.Wrap(err, "parsing account id")
		}

		c := newClient()
		a, err := c.SelectAccount(uint(id))
		if err != nil {
			return errors.Wrap(err, "selecting account")
		}

		t := time.Now()
		if balanceDate.Time != nil {
			t = *balanceDate.Time
		}

		bs, err := accountBalancesAtTime(c, *a, t)
		if err != nil {
			return errors.Wrapf(err, "getting balance at time:%+v for account:%+v", t, a)
		}
		fmt.Println(bs.InnerBalances().Sum())
		return nil
	},
}

// accountBalancesAtTime retrieves the balances that existed for the account at
// a given time.
func accountBalancesAtTime(store storage.Storage, a storage.Account, at time.Time) (storage.Balances, error) {
	bs, err := store.SelectAccountBalances(a.ID)
	if err != nil {
		return storage.Balances{}, errors.Wrapf(err, "selecting balances for account: %+v", a)
	}
	c := filter.BalanceNot(filter.BalanceAfter(at))
	filtered := c.Filter(*bs)

	return filtered, err
}

func init() {
	// TODO: find out how to use same flag on different subcommands instead of
	// TODO: making is persistent here. The issue may arise from using viper to
	// TODO: retrieve them. The issue doesn't happen with custom flags that are
	// TODO: retrieved using a global variable
	accountCmd.PersistentFlags().String(keyCurrency, "", "account currency")
	err := viper.BindPFlags(accountCmd.PersistentFlags())
	if err != nil {
		log.Fatal(errors.Wrap(err, "binding pflags"))
	}
	rootCmd.AddCommand(accountCmd)

	accountAddCmd.Flags().VarP(accountOpened, keyOpened, "o", "account opened date")
	accountAddCmd.Flags().VarP(accountClosed, keyClosed, "c", "account closed date")

	accountOpenCmd.Flags().VarP(accountOpened, keyOpened, "o", "account opened date")
	accountOpenCmd.Flags().IntP(keyOpeningBalance, "b", 0, "account opening balance")
	accountOpenCmd.Flags().String(keyOpeningBalanceNote, "", "note to attach to account opening balance")

	accountCloseCmd.Flags().VarP(balanceDate, keyDate, "d", "account closed date")
	accountCloseCmd.Flags().IntP(keyClosingBalance, "b", 0, "account closing balance")
	accountCloseCmd.Flags().String(keyClosingBalanceNote, "", "note to attach to account closing balance")

	accountUpdateCmd.Flags().StringP(keyName, "n", "", "account name")
	accountUpdateCmd.Flags().VarP(accountOpened, keyOpened, "o", "account opened date")
	accountUpdateCmd.Flags().VarP(accountClosed, keyClosed, "c", "account closed date")

	accountBalancesCmd.Flags().UintP(keyLimit, "l", 0, "limit results")

	// TODO: Stop multiple usage of the flag like in this article: http://blog.ralch.com/tutorial/golang-custom-flags/
	accountBalanceInsertCmd.Flags().VarP(balanceDate, keyDate, "d", "date of balance to insert")
	accountBalanceInsertCmd.Flags().IntP(keyAmount, "a", 0, "amount of balance to insert")
	accountBalanceInsertCmd.Flags().String(keyNote, "", "note to attach to balance")

	accountBalanceCmd.Flags().VarP(balanceDate, keyDate, "d", "date at which to retrieve balance")

	for _, c := range []*cobra.Command{
		accountAddCmd,
		accountOpenCmd,
		accountReopenCmd,
		accountCloseCmd,
		accountUpdateCmd,
		accountDeleteCmd,
		accountRenameCmd,
		accountBalancesCmd,
		accountBalanceInsertCmd,
		accountBalanceCmd,
	} {
		err := viper.BindPFlags(c.Flags())
		if err != nil {
			log.Fatal(errors.Wrap(err, "binding pflags"))
		}
		accountCmd.AddCommand(c)
	}
}
