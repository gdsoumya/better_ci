package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
)

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func PrintLines(filePath string, values []string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	for _, value := range values {
		fmt.Fprintln(f, value) // print values to f, one per line
	}
	return nil
}

func DirPresent(dir string) bool {
	_, err := os.Stat(dir)
	return !os.IsNotExist(err)
}

func SignBody(secret, body []byte) string {
	computed := hmac.New(sha1.New, secret)
	computed.Write(body)
	return "sha1=" + hex.EncodeToString([]byte(computed.Sum(nil)))
}

func VerifySig(sig, secret string, body []byte) bool {
	return sig == SignBody([]byte(secret), body)
}

func GetPublicIP() (string, error) {
	url := "https://api.ipify.org?format=text"
	resp, err := http.Get(url)
	if err != nil {
		return "", errors.New("ERROR : CANNOT GET PUBLIC IP " + err.Error())
	}
	defer resp.Body.Close()
	ip, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("ERROR : CANNOT GET PUBLIC IP " + err.Error())
	}
	return string(ip), nil
}
