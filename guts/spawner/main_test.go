package main

import (
  "testing"
)

func ServeDirectory() {
	pwd, err := os.Getwd()
	CheckError(err)
	port := "9999"
	testFilesDir := pwd + "/../../postgres/test-data/test-files/"
	serveCmd := exec.Command("php", "-S", "localhost:"+port)
	serveCmd.Dir = testFilesDir
	go serveCmd.Run() //nolint:all
	for i := 0; i < 60; i++ {
		timeout := time.Second * 5
		conn, err := net.DialTimeout("tcp", net.JoinHostPort("localhost", port), timeout)
		if err != nil {
			time.Sleep(timeout)
		} else {
			if conn != nil {
				err := conn.Close()
				CheckError(err)
				return
			}
		}
	}
	CheckError(fmt.Errorf("Port never came up when trying to serve directory with command:\n%v", serveCmd))
}

