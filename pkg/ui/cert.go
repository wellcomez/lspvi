// Copyright 2024 wellcomez
// SPDX-License-Identifier: gplv3

package mainui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"zen108.com/lspvi/pkg/ui/common"
)

type Cert struct {
	root      string
	serverkey string
	servercrt string
}

func NewCert() *Cert {
	root, err := common.CreateLspviRoot()
	if err != nil {
		return nil
	}
	c := &Cert{root: root}
	c.serverkey = filepath.Join(c.root, "server.key")
	c.servercrt = filepath.Join(c.root, "server.crt")
	return c
}
func (c *Cert) Getcert() error {
	yes := false
	if _, err := os.Stat(c.serverkey); err != nil {
		yes = true
	}
	if _, err := os.Stat(c.servercrt); err != nil {
		yes = true
	}
	if yes {
		cmd := exec.Command("openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes", "-keyout", c.serverkey, "-out", c.servercrt, "-days", "365", "-subj", "/CN=localhost")
		err := cmd.Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// generatePrivateKey 生成一个 RSA 私钥
func generatePrivateKey(keyFile string) error {
	cmd := exec.Command("openssl", "genrsa", "-out", keyFile, "2048")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error generating private key: %v, output: %s", err, out)
	}
	return nil
}

// generateSelfSignedCert 生成一个自签名证书
func generateSelfSignedCert(certFile, keyFile string) error {
	cmd := exec.Command("openssl", "req", "-x509", "-new", "-nodes", "-key", keyFile, "-sha256", "-days", "3650", "-out", certFile, "-subj", "/C=CN/ST=Beijing/L=Beijing/O=Example Inc./CN=192.168.2.16")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error generating self-signed certificate: %v, output: %s", err, out)
	}
	return nil
}
