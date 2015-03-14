// Copyright (c) 2015 RightScale, Inc. - see LICENSE

package main

// Omega: Alt+937

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// Iterate through all recorded test cases and play them back
var _ = Describe("Recorded request", func() {

	// Open the recording file
	f, err := os.Open("recording.json")
	if err != nil {
		fmt.Fprintf(os.Stdout, "Cannot open recording: %s\n", err.Error())
		os.Exit(1)
	}

	// Iterate through test cases
	for {
		// Read a test case, which is a json struct
		var testCase MyRecording
		err := json.NewDecoder(f).Decode(&testCase)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Json decode: %s\n", err.Error())
			break
		}

		// Create an http handler function that checks that we're indeed getting the
		// expected request and then returns the recorded response
		handler := func(w http.ResponseWriter, req *http.Request) {
			defer GinkgoRecover()
			reqBody, _ := ioutil.ReadAll(req.Body)
			Ω(req.Method).Should(Equal(testCase.RR.Verb))
			Ω(req.URL).Should(Equal(testCase.RR.Uri))
			Ω(reqBody).Should(Equal(testCase.RR.ReqBody))
			for k, _ := range testCase.RR.ReqHeader {
				Ω(req.Header.Get(k)).Should(Equal(testCase.RR.ReqHeader.Get(k)))
			}

			w.WriteHeader(testCase.RR.Status)
			w.Write([]byte(testCase.RR.RespBody))
		}

		// Perform the test by running main() with the command line args set
		It(strings.Join(testCase.CmdArgs, " "), func() {
			server := httptest.NewServer(http.HandlerFunc(handler))
			defer server.Close()

			os.Args = append([]string{
				"rs-api", "--rl10",
				//"--debug",
				"--host", strings.TrimPrefix(server.URL, "http://")},
				testCase.CmdArgs...)
			fmt.Fprintf(os.Stderr, "testing \"%s\"\n", strings.Join(os.Args, `" "`))

			// capture stdout and intercept calls to osExit
			stdoutBuf := bytes.Buffer{}
			osStdout = &stdoutBuf
			exitCode := 99
			osExit = func(code int) { exitCode = code }
			rsClientInternal = nil
			//rightscale().SetInsecure()

			main()

			// Verify that stdout and the exit code are correct
			fmt.Fprintf(os.Stderr, "Exit %d %d\n", exitCode, testCase.ExitCode)
			Ω(exitCode).Should(Equal(testCase.ExitCode))
			fmt.Fprintf(os.Stderr, "stdout <<%s>> <<%s>>\n", stdoutBuf.String(), testCase.Stdout)
			Ω(stdoutBuf.String()).Should(Equal(testCase.Stdout))
		})
	}

})
