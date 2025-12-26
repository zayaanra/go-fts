package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"net/http"
	"bytes"
	"github.com/spf13/cobra"
)

var sendCmd = &cobra.Command{
	Use:	"send",
	Short: 	"Send a file",
	Long: 	"Send a file to a computer",
	Args: 	cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		target := "http://" + args[0] + ":" + "8000"
		fmt.Fprintln(os.Stdout, "Sending to: ", target)
		
		var data = []byte(`{"msg": "Hello World"}`)


		req, err := http.NewRequest("POST", target, bytes.NewBuffer(data))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Something went wrong: ", err)
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Something went wrong: ", err)
		}
		fmt.Printf("Status Code: %d\n", resp.StatusCode)
		fmt.Printf("Response Body: %s\n", body)
	},
}

func init() {
	ftsCmd.AddCommand(sendCmd)
}