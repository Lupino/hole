package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"io/ioutil"
	"log"
	"math/big"
	"time"
)

func main() {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(1653),
		Subject: pkix.Name{
			Country:            []string{"China"},
			Organization:       []string{"huabot.com"},
			OrganizationalUnit: []string{"huabot.com"},
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
	ca_b, err := x509.CreateCertificate(rand.Reader, ca, ca, pub, priv)
	if err != nil {
		log.Println("create ca failed", err)
		return
	}
	ca_f := "hole-server/ca.pem"
	log.Println("write to", ca_f)
	ioutil.WriteFile(ca_f, ca_b, 0777)

	priv_f := "hole-server/ca.key"
	priv_b := x509.MarshalPKCS1PrivateKey(priv)
	log.Println("write to", priv_f)
	ioutil.WriteFile(priv_f, priv_b, 0777)

	cert2 := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			Country:            []string{"China"},
			Organization:       []string{"Fuck"},
			OrganizationalUnit: []string{"FuckU"},
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	priv2, _ := rsa.GenerateKey(rand.Reader, 1024)
	pub2 := &priv2.PublicKey
	cert2_b, err2 := x509.CreateCertificate(rand.Reader, cert2, ca, pub2, priv)
	if err2 != nil {
		log.Println("create cert2 failed", err2)
		return
	}

	cert2_f := "hole-local/cert.pem"
	log.Println("write to", cert2_f)
	ioutil.WriteFile(cert2_f, cert2_b, 0777)

	priv2_f := "hole-local/cert.key"
	priv2_b := x509.MarshalPKCS1PrivateKey(priv2)
	log.Println("write to", priv2_f)
	ioutil.WriteFile(priv2_f, priv2_b, 0777)

	ca_c, _ := x509.ParseCertificate(ca_b)
	cert2_c, _ := x509.ParseCertificate(cert2_b)

	err3 := cert2_c.CheckSignatureFrom(ca_c)
	log.Println("check signature", err3 == nil)
}
