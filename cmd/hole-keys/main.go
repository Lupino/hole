package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"io/ioutil"
	"log"
	"math/big"
	"time"
)

var prefix string

func init() {
	flag.StringVar(&prefix, "prefix", "", "The certificate file prefix.")
	flag.Parse()
}

func main() {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Country:            []string{"China"},
			Organization:       []string{"HoleHUB"},
			OrganizationalUnit: []string{"holehub.com"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		SubjectKeyId:          []byte{1, 2, 3, 4, 5},
		BasicConstraintsValid: true,
		IsCA:        true,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}

	priv, _ := rsa.GenerateKey(rand.Reader, 1024)
	pub := &priv.PublicKey
	caB, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		log.Println("create ca failed", err)
		return
	}
	caF := prefix + "ca.pem"
	log.Println("write to", caF)
	ioutil.WriteFile(caF, caB, 0600)

	privF := prefix + "ca.key"
	privB := x509.MarshalPKCS1PrivateKey(priv)
	log.Println("write to", privF)
	ioutil.WriteFile(privF, privB, 0600)

	cert2 := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Country:            []string{"China"},
			Organization:       []string{"HoleHUB"},
			OrganizationalUnit: []string{"holehub.com"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	priv2, _ := rsa.GenerateKey(rand.Reader, 1024)
	pub2 := &priv2.PublicKey
	cert2B, err2 := x509.CreateCertificate(rand.Reader, cert2, ca, pub2, priv)
	if err2 != nil {
		log.Println("create cert2 failed", err2)
		return
	}

	cert2F := prefix + "cert.pem"
	log.Println("write to", cert2F)
	ioutil.WriteFile(cert2F, cert2B, 0600)

	priv2F := prefix + "cert.key"
	priv2B := x509.MarshalPKCS1PrivateKey(priv2)
	log.Println("write to", priv2F)
	ioutil.WriteFile(priv2F, priv2B, 0600)

	caC, _ := x509.ParseCertificate(caB)
	cert2C, _ := x509.ParseCertificate(cert2B)

	err3 := cert2C.CheckSignatureFrom(caC)
	log.Println("check signature", err3 == nil)
}
