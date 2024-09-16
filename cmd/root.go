package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/JackalLabs/mulberry/relay"
	"github.com/spf13/cobra"
)

func RootCMD() *cobra.Command {
	r := &cobra.Command{
		Use:   "mulberry",
		Short: "Mulberry is a Jackal EVM Relay",
		Long: `Mulberry is a Jackal EVM Relay that delivers packets from 
EVM chains to the Jackal network ot bridge storage capabilities cross-chain.`,
	}

	r.AddCommand(StartCMD(), WalletCMD())

	r.PersistentFlags().String(FLAG_HOME, "$HOME/.mulberry", "where the mulberry config can be found")

	return r
}

func StartCMD() *cobra.Command {
	r := &cobra.Command{
		Use:   "start",
		Short: "Starts the Mulberry Relay service",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := getHome(cmd)
			if err != nil {
				return err
			}

			a, err := relay.MakeApp(home)
			if err != nil {
				return err
			}

			log.Print("Starting relay...")

			return a.Start()
		},
	}

	return r
}

func WalletCMD() *cobra.Command {
	r := &cobra.Command{
		Use:   "wallet",
		Short: "Commands to manage the internal wallet",
	}
	r.AddCommand(AddressCMD())

	return r
}

func getHome(cmd *cobra.Command) (string, error) {
	home, err := cmd.Flags().GetString(FLAG_HOME)
	if err != nil {
		return home, err
	}

	return os.ExpandEnv(home), nil
}

func AddressCMD() *cobra.Command {
	r := &cobra.Command{
		Use:   "address",
		Short: "View the relay address",
		RunE: func(cmd *cobra.Command, args []string) error {
			home, err := getHome(cmd)
			if err != nil {
				return err
			}

			a, err := relay.MakeApp(home)
			if err != nil {
				return err
			}

			fmt.Printf("Wallet Address: %s\n", a.Address())

			return nil
		},
	}

	return r
}
