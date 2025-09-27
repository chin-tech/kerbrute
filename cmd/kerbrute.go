package cmd

import (
	"fmt"
	"os"

	"github.com/chin-tech/kerbrute/util"
	"github.com/spf13/cobra"
)

var WarningString = "\n[WARNING] - Failed Kerberos Pre-Auth attempts count toward failed logins and WILL lock out accounts!\n"

var rootCmd = &cobra.Command{
	Use:   "kerbrute",
	Short: "A tool to perform various bruteforce attacks against Windows Kerberos",
	Long: util.PrintBanner(`
This tool is designed to assist in quickly bruteforcing valid Active Directory accounts through Kerberos Pre-Authentication.
It is designed to be used on an internal Windows domain with access to one of the Domain Controllers.` + WarningString),
}

var completionCmd = &cobra.Command{
	Use:   "completion",
	Short: "Generates command-line completion scripts inside the CWD",
	Long: `Generates command-line completion scripts inside the CWD: BASH, ZSH, FISH, POWERSHELL.
	Bash: Needs the bash-completion package and to be sourced in your shell`,
	Run: func(cmd *cobra.Command, args []string) {

		rootCmd.GenBashCompletionFile("kerbrute_completion.sh")
		rootCmd.GenZshCompletionFile("kerbrute_completion.zsh")
		rootCmd.GenFishCompletionFile("kerbrute_completion.fish", false)
		rootCmd.GenPowerShellCompletionFile("kerbrute_completion.ps1")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&domain, "domain", "d", "", "The full domain to use (e.g. contoso.com)")
	rootCmd.PersistentFlags().StringVar(&domainController, "dc", "", "The location of the Domain Controller (KDC) to target. If blank, will lookup via DNS")
	rootCmd.PersistentFlags().StringVarP(&logFileName, "output", "o", "", "File to write logs to. Optional.")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Log failures and errors")
	rootCmd.PersistentFlags().BoolVar(&safe, "safe", false, "Safe mode. Will abort if any user comes back as locked out. Default: FALSE")
	rootCmd.PersistentFlags().IntVarP(&threads, "threads", "t", 10, "Threads to use")
	rootCmd.PersistentFlags().IntVarP(&delay, "delay", "", 0, "Delay in millisecond between each attempt. Will always use single thread if set")
	rootCmd.PersistentFlags().BoolVar(&downgrade, "downgrade", false, "Force downgraded encryption type (arcfour-hmac-md5)")
	rootCmd.PersistentFlags().StringVar(&hashFileName, "hash-file", "", "File to save AS-REP hashes to (if any captured), otherwise just logged")
	rootCmd.AddCommand(completionCmd)
	if delay != 0 {
		threads = 1
	}

}
