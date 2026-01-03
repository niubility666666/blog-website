package utils

import (
   "crypto/tls"
   "fmt"
   "time"
)

type CertificateInfo struct {
   Domain     string
   NotBefore  time.Time
   NotAfter   time.Time
   IsValid    bool
   ExpireDays int
}

func GetCertificateInfo(domain string) (*CertificateInfo, error) {
   conn, err := tls.Dial("tcp", domain+":443", &tls.Config{
       InsecureSkipVerify: true,
   })
   if err != nil {
       return nil, fmt.Errorf("连接失败: %v", err)
   }
   defer conn.Close()

   certs := conn.ConnectionState().PeerCertificates
   if len(certs) == 0 {
       return nil, fmt.Errorf("未获取到证书")
   }

   cert := certs[0]
   now := time.Now()
   expireDays := int(cert.NotAfter.Sub(now).Hours() / 24)

   return &CertificateInfo{
       Domain:     domain,
       NotBefore:  cert.NotBefore,
       NotAfter:   cert.NotAfter,
       IsValid:    now.After(cert.NotBefore) && now.Before(cert.NotAfter),
       ExpireDays: expireDays,
   }, nil
}
   