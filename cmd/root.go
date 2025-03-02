package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tilt-validator",
	Short: "Tilt Validator Simulation",
	Long:  `A simulation of the Tilt Validator network with threshold signatures.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cyan := color.New(color.FgCyan).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()

		fmt.Printf(`
%s
   _______ _ __ 
  |__   __(_) | |_   
     | |   _| | __| 
     | |  | | | |_  
     |_|  |_|_|\__| %s
%s
    Validator Network Simulation
%s

`, cyan("════════════════════════════════════"),
			yellow("v1.0.0"),
			cyan("════════════════════════════════════"),
			cyan("════════════════════════════════════"))
	},
}

func Execute() error {
	return rootCmd.Execute()
}